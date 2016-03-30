// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

// Contains the active peer-set of the downloader, maintaining both failures
// as well as reputation metrics to prioritize the block retrievals.

package downloader

import (
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ethereum/go-ethereum/common"
)

const (
	maxLackingHashes = 4096 // Maximum number of entries allowed on the list or lacking items
	throughputImpact = 0.1  // The impact a single measurement has on a peer's final throughput value.
)

// Hash and block fetchers belonging to eth/61 and below
type relativeHashFetcherFn func(common.Hash) error
type absoluteHashFetcherFn func(uint64, int) error
type blockFetcherFn func([]common.Hash) error

// Block header and body fetchers belonging to eth/62 and above
type relativeHeaderFetcherFn func(common.Hash, int, int, bool) error
type absoluteHeaderFetcherFn func(uint64, int, int, bool) error
type blockBodyFetcherFn func([]common.Hash) error
type receiptFetcherFn func([]common.Hash) error
type stateFetcherFn func([]common.Hash) error

var (
	errAlreadyFetching   = errors.New("already fetching blocks from peer")
	errAlreadyRegistered = errors.New("peer is already registered")
	errNotRegistered     = errors.New("peer is not registered")
)

// peer represents an active peer from which hashes and blocks are retrieved.
type peer struct {
	id   string      // Unique identifier of the peer
	head common.Hash // Hash of the peers latest known block

	blockIdle   int32 // Current block activity state of the peer (idle = 0, active = 1)
	receiptIdle int32 // Current receipt activity state of the peer (idle = 0, active = 1)
	stateIdle   int32 // Current node data activity state of the peer (idle = 0, active = 1)

	blockThroughput   float64 // Number of blocks (bodies) measured to be retrievable per second
	receiptThroughput float64 // Number of receipts measured to be retrievable per second
	stateThroughput   float64 // Number of node data pieces measured to be retrievable per second

	blockStarted   time.Time // Time instance when the last block (body)fetch was started
	receiptStarted time.Time // Time instance when the last receipt fetch was started
	stateStarted   time.Time // Time instance when the last node data fetch was started

	lacking map[common.Hash]struct{} // Set of hashes not to request (didn't have previously)

	getRelHashes relativeHashFetcherFn // [eth/61] Method to retrieve a batch of hashes from an origin hash
	getAbsHashes absoluteHashFetcherFn // [eth/61] Method to retrieve a batch of hashes from an absolute position
	getBlocks    blockFetcherFn        // [eth/61] Method to retrieve a batch of blocks

	getRelHeaders  relativeHeaderFetcherFn // [eth/62] Method to retrieve a batch of headers from an origin hash
	getAbsHeaders  absoluteHeaderFetcherFn // [eth/62] Method to retrieve a batch of headers from an absolute position
	getBlockBodies blockBodyFetcherFn      // [eth/62] Method to retrieve a batch of block bodies

	getReceipts receiptFetcherFn // [eth/63] Method to retrieve a batch of block transaction receipts
	getNodeData stateFetcherFn   // [eth/63] Method to retrieve a batch of state trie data

	version int // Eth protocol version number to switch strategies
	lock    sync.RWMutex
}

// newPeer create a new downloader peer, with specific hash and block retrieval
// mechanisms.
func newPeer(id string, version int, head common.Hash,
	getRelHashes relativeHashFetcherFn, getAbsHashes absoluteHashFetcherFn, getBlocks blockFetcherFn, // eth/61 callbacks, remove when upgrading
	getRelHeaders relativeHeaderFetcherFn, getAbsHeaders absoluteHeaderFetcherFn, getBlockBodies blockBodyFetcherFn,
	getReceipts receiptFetcherFn, getNodeData stateFetcherFn) *peer {
	return &peer{
		id:      id,
		head:    head,
		lacking: make(map[common.Hash]struct{}),

		getRelHashes: getRelHashes,
		getAbsHashes: getAbsHashes,
		getBlocks:    getBlocks,

		getRelHeaders:  getRelHeaders,
		getAbsHeaders:  getAbsHeaders,
		getBlockBodies: getBlockBodies,

		getReceipts: getReceipts,
		getNodeData: getNodeData,

		version: version,
	}
}

// Reset clears the internal state of a peer entity.
func (p *peer) Reset() {
	p.lock.Lock()
	defer p.lock.Unlock()

	atomic.StoreInt32(&p.blockIdle, 0)
	atomic.StoreInt32(&p.receiptIdle, 0)
	atomic.StoreInt32(&p.stateIdle, 0)

	p.blockThroughput = 0
	p.receiptThroughput = 0
	p.stateThroughput = 0

	p.lacking = make(map[common.Hash]struct{})
}

// Fetch61 sends a block retrieval request to the remote peer.
func (p *peer) Fetch61(request *fetchRequest) error {
	// Sanity check the protocol version
	if p.version != 61 {
		panic(fmt.Sprintf("block fetch [eth/61] requested on eth/%d", p.version))
	}
	// Short circuit if the peer is already fetching
	if !atomic.CompareAndSwapInt32(&p.blockIdle, 0, 1) {
		return errAlreadyFetching
	}
	p.blockStarted = time.Now()

	// Convert the hash set to a retrievable slice
	hashes := make([]common.Hash, 0, len(request.Hashes))
	for hash, _ := range request.Hashes {
		hashes = append(hashes, hash)
	}
	go p.getBlocks(hashes)

	return nil
}

// FetchBodies sends a block body retrieval request to the remote peer.
func (p *peer) FetchBodies(request *fetchRequest) error {
	// Sanity check the protocol version
	if p.version < 62 {
		panic(fmt.Sprintf("body fetch [eth/62+] requested on eth/%d", p.version))
	}
	// Short circuit if the peer is already fetching
	if !atomic.CompareAndSwapInt32(&p.blockIdle, 0, 1) {
		return errAlreadyFetching
	}
	p.blockStarted = time.Now()

	// Convert the header set to a retrievable slice
	hashes := make([]common.Hash, 0, len(request.Headers))
	for _, header := range request.Headers {
		hashes = append(hashes, header.Hash())
	}
	go p.getBlockBodies(hashes)

	return nil
}

// FetchReceipts sends a receipt retrieval request to the remote peer.
func (p *peer) FetchReceipts(request *fetchRequest) error {
	// Sanity check the protocol version
	if p.version < 63 {
		panic(fmt.Sprintf("body fetch [eth/63+] requested on eth/%d", p.version))
	}
	// Short circuit if the peer is already fetching
	if !atomic.CompareAndSwapInt32(&p.receiptIdle, 0, 1) {
		return errAlreadyFetching
	}
	p.receiptStarted = time.Now()

	// Convert the header set to a retrievable slice
	hashes := make([]common.Hash, 0, len(request.Headers))
	for _, header := range request.Headers {
		hashes = append(hashes, header.Hash())
	}
	go p.getReceipts(hashes)

	return nil
}

// FetchNodeData sends a node state data retrieval request to the remote peer.
func (p *peer) FetchNodeData(request *fetchRequest) error {
	// Sanity check the protocol version
	if p.version < 63 {
		panic(fmt.Sprintf("node data fetch [eth/63+] requested on eth/%d", p.version))
	}
	// Short circuit if the peer is already fetching
	if !atomic.CompareAndSwapInt32(&p.stateIdle, 0, 1) {
		return errAlreadyFetching
	}
	p.stateStarted = time.Now()

	// Convert the hash set to a retrievable slice
	hashes := make([]common.Hash, 0, len(request.Hashes))
	for hash, _ := range request.Hashes {
		hashes = append(hashes, hash)
	}
	go p.getNodeData(hashes)

	return nil
}

// SetBlocksIdle sets the peer to idle, allowing it to execute new block retrieval
// requests. Its estimated block retrieval throughput is updated with that measured
// just now.
func (p *peer) SetBlocksIdle(delivered int) {
	p.setIdle(p.blockStarted, delivered, &p.blockThroughput, &p.blockIdle)
}

// SetBodiesIdle sets the peer to idle, allowing it to execute block body retrieval
// requests. Its estimated body retrieval throughput is updated with that measured
// just now.
func (p *peer) SetBodiesIdle(delivered int) {
	p.setIdle(p.blockStarted, delivered, &p.blockThroughput, &p.blockIdle)
}

// SetReceiptsIdle sets the peer to idle, allowing it to execute new receipt
// retrieval requests. Its estimated receipt retrieval throughput is updated
// with that measured just now.
func (p *peer) SetReceiptsIdle(delivered int) {
	p.setIdle(p.receiptStarted, delivered, &p.receiptThroughput, &p.receiptIdle)
}

// SetNodeDataIdle sets the peer to idle, allowing it to execute new state trie
// data retrieval requests. Its estimated state retrieval throughput is updated
// with that measured just now.
func (p *peer) SetNodeDataIdle(delivered int) {
	p.setIdle(p.stateStarted, delivered, &p.stateThroughput, &p.stateIdle)
}

// setIdle sets the peer to idle, allowing it to execute new retrieval requests.
// Its estimated retrieval throughput is updated with that measured just now.
func (p *peer) setIdle(started time.Time, delivered int, throughput *float64, idle *int32) {
	// Irrelevant of the scaling, make sure the peer ends up idle
	defer atomic.StoreInt32(idle, 0)

	p.lock.Lock()
	defer p.lock.Unlock()

	// If nothing was delivered (hard timeout / unavailable data), reduce throughput to minimum
	if delivered == 0 {
		*throughput = 0
		return
	}
	// Otherwise update the throughput with a new measurement
	measured := float64(delivered) / (float64(time.Since(started)+1) / float64(time.Second)) // +1 (ns) to ensure non-zero divisor
	*throughput = (1-throughputImpact)*(*throughput) + throughputImpact*measured
}

// BlockCapacity retrieves the peers block download allowance based on its
// previously discovered throughput.
func (p *peer) BlockCapacity() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return int(math.Max(1, math.Min(p.blockThroughput*float64(blockTargetRTT)/float64(time.Second), float64(MaxBlockFetch))))
}

// ReceiptCapacity retrieves the peers receipt download allowance based on its
// previously discovered throughput.
func (p *peer) ReceiptCapacity() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return int(math.Max(1, math.Min(p.receiptThroughput*float64(receiptTargetRTT)/float64(time.Second), float64(MaxReceiptFetch))))
}

// NodeDataCapacity retrieves the peers state download allowance based on its
// previously discovered throughput.
func (p *peer) NodeDataCapacity() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return int(math.Max(1, math.Min(p.stateThroughput*float64(stateTargetRTT)/float64(time.Second), float64(MaxStateFetch))))
}

// MarkLacking appends a new entity to the set of items (blocks, receipts, states)
// that a peer is known not to have (i.e. have been requested before). If the
// set reaches its maximum allowed capacity, items are randomly dropped off.
func (p *peer) MarkLacking(hash common.Hash) {
	p.lock.Lock()
	defer p.lock.Unlock()

	for len(p.lacking) >= maxLackingHashes {
		for drop, _ := range p.lacking {
			delete(p.lacking, drop)
			break
		}
	}
	p.lacking[hash] = struct{}{}
}

// Lacks retrieves whether the hash of a blockchain item is on the peers lacking
// list (i.e. whether we know that the peer does not have it).
func (p *peer) Lacks(hash common.Hash) bool {
	p.lock.RLock()
	defer p.lock.RUnlock()

	_, ok := p.lacking[hash]
	return ok
}

// String implements fmt.Stringer.
func (p *peer) String() string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return fmt.Sprintf("Peer %s [%s]", p.id,
		fmt.Sprintf("blocks %3.2f/s, ", p.blockThroughput)+
			fmt.Sprintf("receipts %3.2f/s, ", p.receiptThroughput)+
			fmt.Sprintf("states %3.2f/s, ", p.stateThroughput)+
			fmt.Sprintf("lacking %4d", len(p.lacking)),
	)
}

// peerSet represents the collection of active peer participating in the block
// download procedure.
type peerSet struct {
	peers map[string]*peer
	lock  sync.RWMutex
}

// newPeerSet creates a new peer set top track the active download sources.
func newPeerSet() *peerSet {
	return &peerSet{
		peers: make(map[string]*peer),
	}
}

// Reset iterates over the current peer set, and resets each of the known peers
// to prepare for a next batch of block retrieval.
func (ps *peerSet) Reset() {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	for _, peer := range ps.peers {
		peer.Reset()
	}
}

// Register injects a new peer into the working set, or returns an error if the
// peer is already known.
//
// The method also sets the starting throughput values of the new peer to the
// average of all existing peers, to give it a realistic change of being used
// for data retrievals.
func (ps *peerSet) Register(p *peer) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if _, ok := ps.peers[p.id]; ok {
		return errAlreadyRegistered
	}
	if len(ps.peers) > 0 {
		p.blockThroughput, p.receiptThroughput, p.stateThroughput = 0, 0, 0

		for _, peer := range ps.peers {
			peer.lock.RLock()
			p.blockThroughput += peer.blockThroughput
			p.receiptThroughput += peer.receiptThroughput
			p.stateThroughput += peer.stateThroughput
			peer.lock.RUnlock()
		}
		p.blockThroughput /= float64(len(ps.peers))
		p.receiptThroughput /= float64(len(ps.peers))
		p.stateThroughput /= float64(len(ps.peers))
	}
	ps.peers[p.id] = p
	return nil
}

// Unregister removes a remote peer from the active set, disabling any further
// actions to/from that particular entity.
func (ps *peerSet) Unregister(id string) error {
	ps.lock.Lock()
	defer ps.lock.Unlock()

	if _, ok := ps.peers[id]; !ok {
		return errNotRegistered
	}
	delete(ps.peers, id)
	return nil
}

// Peer retrieves the registered peer with the given id.
func (ps *peerSet) Peer(id string) *peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return ps.peers[id]
}

// Len returns if the current number of peers in the set.
func (ps *peerSet) Len() int {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	return len(ps.peers)
}

// AllPeers retrieves a flat list of all the peers within the set.
func (ps *peerSet) AllPeers() []*peer {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	list := make([]*peer, 0, len(ps.peers))
	for _, p := range ps.peers {
		list = append(list, p)
	}
	return list
}

// BlockIdlePeers retrieves a flat list of all the currently idle peers within the
// active peer set, ordered by their reputation.
func (ps *peerSet) BlockIdlePeers() ([]*peer, int) {
	idle := func(p *peer) bool {
		return atomic.LoadInt32(&p.blockIdle) == 0
	}
	throughput := func(p *peer) float64 {
		p.lock.RLock()
		defer p.lock.RUnlock()
		return p.blockThroughput
	}
	return ps.idlePeers(61, 61, idle, throughput)
}

// BodyIdlePeers retrieves a flat list of all the currently body-idle peers within
// the active peer set, ordered by their reputation.
func (ps *peerSet) BodyIdlePeers() ([]*peer, int) {
	idle := func(p *peer) bool {
		return atomic.LoadInt32(&p.blockIdle) == 0
	}
	throughput := func(p *peer) float64 {
		p.lock.RLock()
		defer p.lock.RUnlock()
		return p.blockThroughput
	}
	return ps.idlePeers(62, 64, idle, throughput)
}

// ReceiptIdlePeers retrieves a flat list of all the currently receipt-idle peers
// within the active peer set, ordered by their reputation.
func (ps *peerSet) ReceiptIdlePeers() ([]*peer, int) {
	idle := func(p *peer) bool {
		return atomic.LoadInt32(&p.receiptIdle) == 0
	}
	throughput := func(p *peer) float64 {
		p.lock.RLock()
		defer p.lock.RUnlock()
		return p.receiptThroughput
	}
	return ps.idlePeers(63, 64, idle, throughput)
}

// NodeDataIdlePeers retrieves a flat list of all the currently node-data-idle
// peers within the active peer set, ordered by their reputation.
func (ps *peerSet) NodeDataIdlePeers() ([]*peer, int) {
	idle := func(p *peer) bool {
		return atomic.LoadInt32(&p.stateIdle) == 0
	}
	throughput := func(p *peer) float64 {
		p.lock.RLock()
		defer p.lock.RUnlock()
		return p.stateThroughput
	}
	return ps.idlePeers(63, 64, idle, throughput)
}

// idlePeers retrieves a flat list of all currently idle peers satisfying the
// protocol version constraints, using the provided function to check idleness.
// The resulting set of peers are sorted by their measure throughput.
func (ps *peerSet) idlePeers(minProtocol, maxProtocol int, idleCheck func(*peer) bool, throughput func(*peer) float64) ([]*peer, int) {
	ps.lock.RLock()
	defer ps.lock.RUnlock()

	idle, total := make([]*peer, 0, len(ps.peers)), 0
	for _, p := range ps.peers {
		if p.version >= minProtocol && p.version <= maxProtocol {
			if idleCheck(p) {
				idle = append(idle, p)
			}
			total++
		}
	}
	for i := 0; i < len(idle); i++ {
		for j := i + 1; j < len(idle); j++ {
			if throughput(idle[i]) < throughput(idle[j]) {
				idle[i], idle[j] = idle[j], idle[i]
			}
		}
	}
	return idle, total
}
