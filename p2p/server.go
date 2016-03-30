// Copyright 2014 The go-ethereum Authors
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

// Package p2p implements the Ethereum p2p network protocols.
package p2p

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
)

const (
	defaultDialTimeout      = 15 * time.Second
	refreshPeersInterval    = 30 * time.Second
	staticPeerCheckInterval = 15 * time.Second

	// Maximum number of concurrently handshaking inbound connections.
	maxAcceptConns = 50

	// Maximum number of concurrently dialing outbound connections.
	maxActiveDialTasks = 16

	// Maximum time allowed for reading a complete message.
	// This is effectively the amount of time a connection can be idle.
	frameReadTimeout = 30 * time.Second

	// Maximum amount of time allowed for writing a complete message.
	frameWriteTimeout = 20 * time.Second
)

var errServerStopped = errors.New("server stopped")

var srvjslog = logger.NewJsonLogger()

// Server manages all peer connections.
//
// The fields of Server are used as configuration parameters.
// You should set them before starting the Server. Fields may not be
// modified while the server is running.
type Server struct {
	// This field must be set to a valid secp256k1 private key.
	PrivateKey *ecdsa.PrivateKey

	// MaxPeers is the maximum number of peers that can be
	// connected. It must be greater than zero.
	MaxPeers int

	// MaxPendingPeers is the maximum number of peers that can be pending in the
	// handshake phase, counted separately for inbound and outbound connections.
	// Zero defaults to preset values.
	MaxPendingPeers int

	// Discovery specifies whether the peer discovery mechanism should be started
	// or not. Disabling is usually useful for protocol debugging (manual topology).
	Discovery bool

	// Name sets the node name of this server.
	// Use common.MakeName to create a name that follows existing conventions.
	Name string

	// Bootstrap nodes are used to establish connectivity
	// with the rest of the network.
	BootstrapNodes []*discover.Node

	// Static nodes are used as pre-configured connections which are always
	// maintained and re-connected on disconnects.
	StaticNodes []*discover.Node

	// Trusted nodes are used as pre-configured connections which are always
	// allowed to connect, even above the peer limit.
	TrustedNodes []*discover.Node

	// NodeDatabase is the path to the database containing the previously seen
	// live nodes in the network.
	NodeDatabase string

	// Protocols should contain the protocols supported
	// by the server. Matching protocols are launched for
	// each peer.
	Protocols []Protocol

	// If ListenAddr is set to a non-nil address, the server
	// will listen for incoming connections.
	//
	// If the port is zero, the operating system will pick a port. The
	// ListenAddr field will be updated with the actual address when
	// the server is started.
	ListenAddr string

	// If set to a non-nil value, the given NAT port mapper
	// is used to make the listening port available to the
	// Internet.
	NAT nat.Interface

	// If Dialer is set to a non-nil value, the given Dialer
	// is used to dial outbound peer connections.
	Dialer *net.Dialer

	// If NoDial is true, the server will not dial any peers.
	NoDial bool

	// Hooks for testing. These are useful because we can inhibit
	// the whole protocol stack.
	newTransport func(net.Conn) transport
	newPeerHook  func(*Peer)

	lock    sync.Mutex // protects running
	running bool

	ntab         discoverTable
	listener     net.Listener
	ourHandshake *protoHandshake
	lastLookup   time.Time

	// These are for Peers, PeerCount (and nothing else).
	peerOp     chan peerOpFunc
	peerOpDone chan struct{}

	quit          chan struct{}
	addstatic     chan *discover.Node
	posthandshake chan *conn
	addpeer       chan *conn
	delpeer       chan *Peer
	loopWG        sync.WaitGroup // loop, listenLoop
}

type peerOpFunc func(map[discover.NodeID]*Peer)

type connFlag int

const (
	dynDialedConn connFlag = 1 << iota
	staticDialedConn
	inboundConn
	trustedConn
)

// conn wraps a network connection with information gathered
// during the two handshakes.
type conn struct {
	fd net.Conn
	transport
	flags connFlag
	cont  chan error      // The run loop uses cont to signal errors to setupConn.
	id    discover.NodeID // valid after the encryption handshake
	caps  []Cap           // valid after the protocol handshake
	name  string          // valid after the protocol handshake
}

type transport interface {
	// The two handshakes.
	doEncHandshake(prv *ecdsa.PrivateKey, dialDest *discover.Node) (discover.NodeID, error)
	doProtoHandshake(our *protoHandshake) (*protoHandshake, error)
	// The MsgReadWriter can only be used after the encryption
	// handshake has completed. The code uses conn.id to track this
	// by setting it to a non-nil value after the encryption handshake.
	MsgReadWriter
	// transports must provide Close because we use MsgPipe in some of
	// the tests. Closing the actual network connection doesn't do
	// anything in those tests because NsgPipe doesn't use it.
	close(err error)
}

func (c *conn) String() string {
	s := c.flags.String() + " conn"
	if (c.id != discover.NodeID{}) {
		s += fmt.Sprintf(" %x", c.id[:8])
	}
	s += " " + c.fd.RemoteAddr().String()
	return s
}

func (f connFlag) String() string {
	s := ""
	if f&trustedConn != 0 {
		s += " trusted"
	}
	if f&dynDialedConn != 0 {
		s += " dyn dial"
	}
	if f&staticDialedConn != 0 {
		s += " static dial"
	}
	if f&inboundConn != 0 {
		s += " inbound"
	}
	if s != "" {
		s = s[1:]
	}
	return s
}

func (c *conn) is(f connFlag) bool {
	return c.flags&f != 0
}

// Peers returns all connected peers.
func (srv *Server) Peers() []*Peer {
	var ps []*Peer
	select {
	// Note: We'd love to put this function into a variable but
	// that seems to cause a weird compiler error in some
	// environments.
	case srv.peerOp <- func(peers map[discover.NodeID]*Peer) {
		for _, p := range peers {
			ps = append(ps, p)
		}
	}:
		<-srv.peerOpDone
	case <-srv.quit:
	}
	return ps
}

// PeerCount returns the number of connected peers.
func (srv *Server) PeerCount() int {
	var count int
	select {
	case srv.peerOp <- func(ps map[discover.NodeID]*Peer) { count = len(ps) }:
		<-srv.peerOpDone
	case <-srv.quit:
	}
	return count
}

// AddPeer connects to the given node and maintains the connection until the
// server is shut down. If the connection fails for any reason, the server will
// attempt to reconnect the peer.
func (srv *Server) AddPeer(node *discover.Node) {
	select {
	case srv.addstatic <- node:
	case <-srv.quit:
	}
}

// Self returns the local node's endpoint information.
func (srv *Server) Self() *discover.Node {
	srv.lock.Lock()
	defer srv.lock.Unlock()

	// If the server's not running, return an empty node
	if !srv.running {
		return &discover.Node{IP: net.ParseIP("0.0.0.0")}
	}
	// If the node is running but discovery is off, manually assemble the node infos
	if srv.ntab == nil {
		// Inbound connections disabled, use zero address
		if srv.listener == nil {
			return &discover.Node{IP: net.ParseIP("0.0.0.0"), ID: discover.PubkeyID(&srv.PrivateKey.PublicKey)}
		}
		// Otherwise inject the listener address too
		addr := srv.listener.Addr().(*net.TCPAddr)
		return &discover.Node{
			ID:  discover.PubkeyID(&srv.PrivateKey.PublicKey),
			IP:  addr.IP,
			TCP: uint16(addr.Port),
		}
	}
	// Otherwise return the live node infos
	return srv.ntab.Self()
}

// Stop terminates the server and all active peer connections.
// It blocks until all active connections have been closed.
func (srv *Server) Stop() {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if !srv.running {
		return
	}
	srv.running = false
	if srv.listener != nil {
		// this unblocks listener Accept
		srv.listener.Close()
	}
	close(srv.quit)
	srv.loopWG.Wait()
}

// Start starts running the server.
// Servers can not be re-used after stopping.
func (srv *Server) Start() (err error) {
	srv.lock.Lock()
	defer srv.lock.Unlock()
	if srv.running {
		return errors.New("server already running")
	}
	srv.running = true
	glog.V(logger.Info).Infoln("Starting Server")

	// static fields
	if srv.PrivateKey == nil {
		return fmt.Errorf("Server.PrivateKey must be set to a non-nil key")
	}
	if srv.newTransport == nil {
		srv.newTransport = newRLPX
	}
	if srv.Dialer == nil {
		srv.Dialer = &net.Dialer{Timeout: defaultDialTimeout}
	}
	srv.quit = make(chan struct{})
	srv.addpeer = make(chan *conn)
	srv.delpeer = make(chan *Peer)
	srv.posthandshake = make(chan *conn)
	srv.addstatic = make(chan *discover.Node)
	srv.peerOp = make(chan peerOpFunc)
	srv.peerOpDone = make(chan struct{})

	// node table
	if srv.Discovery {
		ntab, err := discover.ListenUDP(srv.PrivateKey, srv.ListenAddr, srv.NAT, srv.NodeDatabase)
		if err != nil {
			return err
		}
		if err := ntab.SetFallbackNodes(srv.BootstrapNodes); err != nil {
			return err
		}
		srv.ntab = ntab
	}

	dynPeers := (srv.MaxPeers + 1) / 2
	if !srv.Discovery {
		dynPeers = 0
	}
	dialer := newDialState(srv.StaticNodes, srv.ntab, dynPeers)

	// handshake
	srv.ourHandshake = &protoHandshake{Version: baseProtocolVersion, Name: srv.Name, ID: discover.PubkeyID(&srv.PrivateKey.PublicKey)}
	for _, p := range srv.Protocols {
		srv.ourHandshake.Caps = append(srv.ourHandshake.Caps, p.cap())
	}
	// listen/dial
	if srv.ListenAddr != "" {
		if err := srv.startListening(); err != nil {
			return err
		}
	}
	if srv.NoDial && srv.ListenAddr == "" {
		glog.V(logger.Warn).Infoln("I will be kind-of useless, neither dialing nor listening.")
	}

	srv.loopWG.Add(1)
	go srv.run(dialer)
	srv.running = true
	return nil
}

func (srv *Server) startListening() error {
	// Launch the TCP listener.
	listener, err := net.Listen("tcp", srv.ListenAddr)
	if err != nil {
		return err
	}
	laddr := listener.Addr().(*net.TCPAddr)
	srv.ListenAddr = laddr.String()
	srv.listener = listener
	srv.loopWG.Add(1)
	go srv.listenLoop()
	// Map the TCP listening port if NAT is configured.
	if !laddr.IP.IsLoopback() && srv.NAT != nil {
		srv.loopWG.Add(1)
		go func() {
			nat.Map(srv.NAT, srv.quit, "tcp", laddr.Port, laddr.Port, "ethereum p2p")
			srv.loopWG.Done()
		}()
	}
	return nil
}

type dialer interface {
	newTasks(running int, peers map[discover.NodeID]*Peer, now time.Time) []task
	taskDone(task, time.Time)
	addStatic(*discover.Node)
}

func (srv *Server) run(dialstate dialer) {
	defer srv.loopWG.Done()
	var (
		peers   = make(map[discover.NodeID]*Peer)
		trusted = make(map[discover.NodeID]bool, len(srv.TrustedNodes))

		tasks        []task
		pendingTasks []task
		taskdone     = make(chan task, maxActiveDialTasks)
	)
	// Put trusted nodes into a map to speed up checks.
	// Trusted peers are loaded on startup and cannot be
	// modified while the server is running.
	for _, n := range srv.TrustedNodes {
		trusted[n.ID] = true
	}

	// Some task list helpers.
	delTask := func(t task) {
		for i := range tasks {
			if tasks[i] == t {
				tasks = append(tasks[:i], tasks[i+1:]...)
				break
			}
		}
	}
	scheduleTasks := func(new []task) {
		pt := append(pendingTasks, new...)
		start := maxActiveDialTasks - len(tasks)
		if len(pt) < start {
			start = len(pt)
		}
		if start > 0 {
			tasks = append(tasks, pt[:start]...)
			for _, t := range pt[:start] {
				t := t
				glog.V(logger.Detail).Infoln("new task:", t)
				go func() { t.Do(srv); taskdone <- t }()
			}
			copy(pt, pt[start:])
			pendingTasks = pt[:len(pt)-start]
		}
	}

running:
	for {
		// Query the dialer for new tasks and launch them.
		now := time.Now()
		nt := dialstate.newTasks(len(pendingTasks)+len(tasks), peers, now)
		scheduleTasks(nt)

		select {
		case <-srv.quit:
			// The server was stopped. Run the cleanup logic.
			glog.V(logger.Detail).Infoln("<-quit: spinning down")
			break running
		case n := <-srv.addstatic:
			// This channel is used by AddPeer to add to the
			// ephemeral static peer list. Add it to the dialer,
			// it will keep the node connected.
			glog.V(logger.Detail).Infoln("<-addstatic:", n)
			dialstate.addStatic(n)
		case op := <-srv.peerOp:
			// This channel is used by Peers and PeerCount.
			op(peers)
			srv.peerOpDone <- struct{}{}
		case t := <-taskdone:
			// A task got done. Tell dialstate about it so it
			// can update its state and remove it from the active
			// tasks list.
			glog.V(logger.Detail).Infoln("<-taskdone:", t)
			dialstate.taskDone(t, now)
			delTask(t)
		case c := <-srv.posthandshake:
			// A connection has passed the encryption handshake so
			// the remote identity is known (but hasn't been verified yet).
			if trusted[c.id] {
				// Ensure that the trusted flag is set before checking against MaxPeers.
				c.flags |= trustedConn
			}
			glog.V(logger.Detail).Infoln("<-posthandshake:", c)
			// TODO: track in-progress inbound node IDs (pre-Peer) to avoid dialing them.
			c.cont <- srv.encHandshakeChecks(peers, c)
		case c := <-srv.addpeer:
			// At this point the connection is past the protocol handshake.
			// Its capabilities are known and the remote identity is verified.
			glog.V(logger.Detail).Infoln("<-addpeer:", c)
			err := srv.protoHandshakeChecks(peers, c)
			if err != nil {
				glog.V(logger.Detail).Infof("Not adding %v as peer: %v", c, err)
			} else {
				// The handshakes are done and it passed all checks.
				p := newPeer(c, srv.Protocols)
				peers[c.id] = p
				go srv.runPeer(p)
			}
			// The dialer logic relies on the assumption that
			// dial tasks complete after the peer has been added or
			// discarded. Unblock the task last.
			c.cont <- err
		case p := <-srv.delpeer:
			// A peer disconnected.
			glog.V(logger.Detail).Infoln("<-delpeer:", p)
			delete(peers, p.ID())
		}
	}

	// Terminate discovery. If there is a running lookup it will terminate soon.
	if srv.ntab != nil {
		srv.ntab.Close()
	}
	// Disconnect all peers.
	for _, p := range peers {
		p.Disconnect(DiscQuitting)
	}
	// Wait for peers to shut down. Pending connections and tasks are
	// not handled here and will terminate soon-ish because srv.quit
	// is closed.
	glog.V(logger.Detail).Infof("ignoring %d pending tasks at spindown", len(tasks))
	for len(peers) > 0 {
		p := <-srv.delpeer
		glog.V(logger.Detail).Infoln("<-delpeer (spindown):", p)
		delete(peers, p.ID())
	}
}

func (srv *Server) protoHandshakeChecks(peers map[discover.NodeID]*Peer, c *conn) error {
	// Drop connections with no matching protocols.
	if len(srv.Protocols) > 0 && countMatchingProtocols(srv.Protocols, c.caps) == 0 {
		return DiscUselessPeer
	}
	// Repeat the encryption handshake checks because the
	// peer set might have changed between the handshakes.
	return srv.encHandshakeChecks(peers, c)
}

func (srv *Server) encHandshakeChecks(peers map[discover.NodeID]*Peer, c *conn) error {
	switch {
	case !c.is(trustedConn|staticDialedConn) && len(peers) >= srv.MaxPeers:
		return DiscTooManyPeers
	case peers[c.id] != nil:
		return DiscAlreadyConnected
	case c.id == srv.Self().ID:
		return DiscSelf
	default:
		return nil
	}
}

type tempError interface {
	Temporary() bool
}

// listenLoop runs in its own goroutine and accepts
// inbound connections.
func (srv *Server) listenLoop() {
	defer srv.loopWG.Done()
	glog.V(logger.Info).Infoln("Listening on", srv.listener.Addr())

	// This channel acts as a semaphore limiting
	// active inbound connections that are lingering pre-handshake.
	// If all slots are taken, no further connections are accepted.
	tokens := maxAcceptConns
	if srv.MaxPendingPeers > 0 {
		tokens = srv.MaxPendingPeers
	}
	slots := make(chan struct{}, tokens)
	for i := 0; i < tokens; i++ {
		slots <- struct{}{}
	}

	for {
		// Wait for a handshake slot before accepting.
		<-slots

		var (
			fd  net.Conn
			err error
		)
		for {
			fd, err = srv.listener.Accept()
			if tempErr, ok := err.(tempError); ok && tempErr.Temporary() {
				glog.V(logger.Debug).Infof("Temporary read error: %v", err)
				continue
			} else if err != nil {
				glog.V(logger.Debug).Infof("Read error: %v", err)
				return
			}
			break
		}
		fd = newMeteredConn(fd, true)
		glog.V(logger.Debug).Infof("Accepted conn %v\n", fd.RemoteAddr())

		// Spawn the handler. It will give the slot back when the connection
		// has been established.
		go func() {
			srv.setupConn(fd, inboundConn, nil)
			slots <- struct{}{}
		}()
	}
}

// setupConn runs the handshakes and attempts to add the connection
// as a peer. It returns when the connection has been added as a peer
// or the handshakes have failed.
func (srv *Server) setupConn(fd net.Conn, flags connFlag, dialDest *discover.Node) {
	// Prevent leftover pending conns from entering the handshake.
	srv.lock.Lock()
	running := srv.running
	srv.lock.Unlock()
	c := &conn{fd: fd, transport: srv.newTransport(fd), flags: flags, cont: make(chan error)}
	if !running {
		c.close(errServerStopped)
		return
	}
	// Run the encryption handshake.
	var err error
	if c.id, err = c.doEncHandshake(srv.PrivateKey, dialDest); err != nil {
		glog.V(logger.Debug).Infof("%v faild enc handshake: %v", c, err)
		c.close(err)
		return
	}
	// For dialed connections, check that the remote public key matches.
	if dialDest != nil && c.id != dialDest.ID {
		c.close(DiscUnexpectedIdentity)
		glog.V(logger.Debug).Infof("%v dialed identity mismatch, want %x", c, dialDest.ID[:8])
		return
	}
	if err := srv.checkpoint(c, srv.posthandshake); err != nil {
		glog.V(logger.Debug).Infof("%v failed checkpoint posthandshake: %v", c, err)
		c.close(err)
		return
	}
	// Run the protocol handshake
	phs, err := c.doProtoHandshake(srv.ourHandshake)
	if err != nil {
		glog.V(logger.Debug).Infof("%v failed proto handshake: %v", c, err)
		c.close(err)
		return
	}
	if phs.ID != c.id {
		glog.V(logger.Debug).Infof("%v wrong proto handshake identity: %x", c, phs.ID[:8])
		c.close(DiscUnexpectedIdentity)
		return
	}
	c.caps, c.name = phs.Caps, phs.Name
	if err := srv.checkpoint(c, srv.addpeer); err != nil {
		glog.V(logger.Debug).Infof("%v failed checkpoint addpeer: %v", c, err)
		c.close(err)
		return
	}
	// If the checks completed successfully, runPeer has now been
	// launched by run.
}

// checkpoint sends the conn to run, which performs the
// post-handshake checks for the stage (posthandshake, addpeer).
func (srv *Server) checkpoint(c *conn, stage chan<- *conn) error {
	select {
	case stage <- c:
	case <-srv.quit:
		return errServerStopped
	}
	select {
	case err := <-c.cont:
		return err
	case <-srv.quit:
		return errServerStopped
	}
}

// runPeer runs in its own goroutine for each peer.
// it waits until the Peer logic returns and removes
// the peer.
func (srv *Server) runPeer(p *Peer) {
	glog.V(logger.Debug).Infof("Added %v\n", p)
	srvjslog.LogJson(&logger.P2PConnected{
		RemoteId:            p.ID().String(),
		RemoteAddress:       p.RemoteAddr().String(),
		RemoteVersionString: p.Name(),
		NumConnections:      srv.PeerCount(),
	})

	if srv.newPeerHook != nil {
		srv.newPeerHook(p)
	}
	discreason := p.run()
	// Note: run waits for existing peers to be sent on srv.delpeer
	// before returning, so this send should not select on srv.quit.
	srv.delpeer <- p

	glog.V(logger.Debug).Infof("Removed %v (%v)\n", p, discreason)
	srvjslog.LogJson(&logger.P2PDisconnected{
		RemoteId:       p.ID().String(),
		NumConnections: srv.PeerCount(),
	})
}

// NodeInfo represents a short summary of the information known about the host.
type NodeInfo struct {
	ID    string `json:"id"`    // Unique node identifier (also the encryption key)
	Name  string `json:"name"`  // Name of the node, including client type, version, OS, custom data
	Enode string `json:"enode"` // Enode URL for adding this peer from remote peers
	IP    string `json:"ip"`    // IP address of the node
	Ports struct {
		Discovery int `json:"discovery"` // UDP listening port for discovery protocol
		Listener  int `json:"listener"`  // TCP listening port for RLPx
	} `json:"ports"`
	ListenAddr string                 `json:"listenAddr"`
	Protocols  map[string]interface{} `json:"protocols"`
}

// Info gathers and returns a collection of metadata known about the host.
func (srv *Server) NodeInfo() *NodeInfo {
	node := srv.Self()

	// Gather and assemble the generic node infos
	info := &NodeInfo{
		Name:       srv.Name,
		Enode:      node.String(),
		ID:         node.ID.String(),
		IP:         node.IP.String(),
		ListenAddr: srv.ListenAddr,
		Protocols:  make(map[string]interface{}),
	}
	info.Ports.Discovery = int(node.UDP)
	info.Ports.Listener = int(node.TCP)

	// Gather all the running protocol infos (only once per protocol type)
	for _, proto := range srv.Protocols {
		if _, ok := info.Protocols[proto.Name]; !ok {
			nodeInfo := interface{}("unknown")
			if query := proto.NodeInfo; query != nil {
				nodeInfo = proto.NodeInfo()
			}
			info.Protocols[proto.Name] = nodeInfo
		}
	}
	return info
}

// PeersInfo returns an array of metadata objects describing connected peers.
func (srv *Server) PeersInfo() []*PeerInfo {
	// Gather all the generic and sub-protocol specific infos
	infos := make([]*PeerInfo, 0, srv.PeerCount())
	for _, peer := range srv.Peers() {
		if peer != nil {
			infos = append(infos, peer.Info())
		}
	}
	// Sort the result array alphabetically by node identifier
	for i := 0; i < len(infos); i++ {
		for j := i + 1; j < len(infos); j++ {
			if infos[i].ID > infos[j].ID {
				infos[i], infos[j] = infos[j], infos[i]
			}
		}
	}
	return infos
}
