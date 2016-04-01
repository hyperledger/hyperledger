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

// Package trie implements Merkle Patricia Tries.
package trie

import (
	"bytes"
	"fmt"
	"hash"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"
	"github.com/ethereum/go-ethereum/rlp"
)

const defaultCacheCapacity = 800

var (
	// The global cache stores decoded trie nodes by hash as they get loaded.
	globalCache = newARC(defaultCacheCapacity)

	// This is the known root hash of an empty trie.
	emptyRoot = common.HexToHash("56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421")

	// This is the known hash of an empty state trie entry.
	emptyState = crypto.Keccak256Hash(nil)
)

// ClearGlobalCache clears the global trie cache
func ClearGlobalCache() {
	globalCache.Clear()
}

// Database must be implemented by backing stores for the trie.
type Database interface {
	DatabaseWriter
	// Get returns the value for key from the database.
	Get(key []byte) (value []byte, err error)
}

// DatabaseWriter wraps the Put method of a backing store for the trie.
type DatabaseWriter interface {
	// Put stores the mapping key->value in the database.
	// Implementations must not hold onto the value bytes, the trie
	// will reuse the slice across calls to Put.
	Put(key, value []byte) error
}

// Trie is a Merkle Patricia Trie.
// The zero value is an empty trie with no database.
// Use New to create a trie that sits on top of a database.
//
// Trie is not safe for concurrent use.
type Trie struct {
	root         node
	db           Database
	originalRoot common.Hash
	*hasher
}

// New creates a trie with an existing root node from db.
//
// If root is the zero hash or the sha3 hash of an empty string, the
// trie is initially empty and does not require a database. Otherwise,
// New will panic if db is nil and returns a MissingNodeError if root does
// not exist in the database. Accessing the trie loads nodes from db on demand.
func New(root common.Hash, db Database) (*Trie, error) {
	trie := &Trie{db: db, originalRoot: root}
	if (root != common.Hash{}) && root != emptyRoot {
		if db == nil {
			panic("trie.New: cannot use existing root without a database")
		}
		if v, _ := trie.db.Get(root[:]); len(v) == 0 {
			return nil, &MissingNodeError{
				RootHash: root,
				NodeHash: root,
			}
		}
		trie.root = hashNode(root.Bytes())
	}
	return trie, nil
}

// Iterator returns an iterator over all mappings in the trie.
func (t *Trie) Iterator() *Iterator {
	return NewIterator(t)
}

// Get returns the value for key stored in the trie.
// The value bytes must not be modified by the caller.
func (t *Trie) Get(key []byte) []byte {
	res, err := t.TryGet(key)
	if err != nil && glog.V(logger.Error) {
		glog.Errorf("Unhandled trie error: %v", err)
	}
	return res
}

// TryGet returns the value for key stored in the trie.
// The value bytes must not be modified by the caller.
// If a node was not found in the database, a MissingNodeError is returned.
func (t *Trie) TryGet(key []byte) ([]byte, error) {
	key = compactHexDecode(key)
	pos := 0
	tn := t.root
	for pos < len(key) {
		switch n := tn.(type) {
		case shortNode:
			if len(key)-pos < len(n.Key) || !bytes.Equal(n.Key, key[pos:pos+len(n.Key)]) {
				return nil, nil
			}
			tn = n.Val
			pos += len(n.Key)
		case fullNode:
			tn = n[key[pos]]
			pos++
		case nil:
			return nil, nil
		case hashNode:
			var err error
			tn, err = t.resolveHash(n, key[:pos], key[pos:])
			if err != nil {
				return nil, err
			}
		default:
			panic(fmt.Sprintf("%T: invalid node: %v", tn, tn))
		}
	}
	return tn.(valueNode), nil
}

// Update associates key with value in the trie. Subsequent calls to
// Get will return value. If value has length zero, any existing value
// is deleted from the trie and calls to Get will return nil.
//
// The value bytes must not be modified by the caller while they are
// stored in the trie.
func (t *Trie) Update(key, value []byte) {
	if err := t.TryUpdate(key, value); err != nil && glog.V(logger.Error) {
		glog.Errorf("Unhandled trie error: %v", err)
	}
}

// TryUpdate associates key with value in the trie. Subsequent calls to
// Get will return value. If value has length zero, any existing value
// is deleted from the trie and calls to Get will return nil.
//
// The value bytes must not be modified by the caller while they are
// stored in the trie.
//
// If a node was not found in the database, a MissingNodeError is returned.
func (t *Trie) TryUpdate(key, value []byte) error {
	k := compactHexDecode(key)
	if len(value) != 0 {
		n, err := t.insert(t.root, nil, k, valueNode(value))
		if err != nil {
			return err
		}
		t.root = n
	} else {
		n, err := t.delete(t.root, nil, k)
		if err != nil {
			return err
		}
		t.root = n
	}
	return nil
}

func (t *Trie) insert(n node, prefix, key []byte, value node) (node, error) {
	if len(key) == 0 {
		return value, nil
	}
	switch n := n.(type) {
	case shortNode:
		matchlen := prefixLen(key, n.Key)
		// If the whole key matches, keep this short node as is
		// and only update the value.
		if matchlen == len(n.Key) {
			nn, err := t.insert(n.Val, append(prefix, key[:matchlen]...), key[matchlen:], value)
			if err != nil {
				return nil, err
			}
			return shortNode{n.Key, nn}, nil
		}
		// Otherwise branch out at the index where they differ.
		var branch fullNode
		var err error
		branch[n.Key[matchlen]], err = t.insert(nil, append(prefix, n.Key[:matchlen+1]...), n.Key[matchlen+1:], n.Val)
		if err != nil {
			return nil, err
		}
		branch[key[matchlen]], err = t.insert(nil, append(prefix, key[:matchlen+1]...), key[matchlen+1:], value)
		if err != nil {
			return nil, err
		}
		// Replace this shortNode with the branch if it occurs at index 0.
		if matchlen == 0 {
			return branch, nil
		}
		// Otherwise, replace it with a short node leading up to the branch.
		return shortNode{key[:matchlen], branch}, nil

	case fullNode:
		nn, err := t.insert(n[key[0]], append(prefix, key[0]), key[1:], value)
		if err != nil {
			return nil, err
		}
		n[key[0]] = nn
		return n, nil

	case nil:
		return shortNode{key, value}, nil

	case hashNode:
		// We've hit a part of the trie that isn't loaded yet. Load
		// the node and insert into it. This leaves all child nodes on
		// the path to the value in the trie.
		//
		// TODO: track whether insertion changed the value and keep
		// n as a hash node if it didn't.
		rn, err := t.resolveHash(n, prefix, key)
		if err != nil {
			return nil, err
		}
		return t.insert(rn, prefix, key, value)

	default:
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

// Delete removes any existing value for key from the trie.
func (t *Trie) Delete(key []byte) {
	if err := t.TryDelete(key); err != nil && glog.V(logger.Error) {
		glog.Errorf("Unhandled trie error: %v", err)
	}
}

// TryDelete removes any existing value for key from the trie.
// If a node was not found in the database, a MissingNodeError is returned.
func (t *Trie) TryDelete(key []byte) error {
	k := compactHexDecode(key)
	n, err := t.delete(t.root, nil, k)
	if err != nil {
		return err
	}
	t.root = n
	return nil
}

// delete returns the new root of the trie with key deleted.
// It reduces the trie to minimal form by simplifying
// nodes on the way up after deleting recursively.
func (t *Trie) delete(n node, prefix, key []byte) (node, error) {
	switch n := n.(type) {
	case shortNode:
		matchlen := prefixLen(key, n.Key)
		if matchlen < len(n.Key) {
			return n, nil // don't replace n on mismatch
		}
		if matchlen == len(key) {
			return nil, nil // remove n entirely for whole matches
		}
		// The key is longer than n.Key. Remove the remaining suffix
		// from the subtrie. Child can never be nil here since the
		// subtrie must contain at least two other values with keys
		// longer than n.Key.
		child, err := t.delete(n.Val, append(prefix, key[:len(n.Key)]...), key[len(n.Key):])
		if err != nil {
			return nil, err
		}
		switch child := child.(type) {
		case shortNode:
			// Deleting from the subtrie reduced it to another
			// short node. Merge the nodes to avoid creating a
			// shortNode{..., shortNode{...}}. Use concat (which
			// always creates a new slice) instead of append to
			// avoid modifying n.Key since it might be shared with
			// other nodes.
			return shortNode{concat(n.Key, child.Key...), child.Val}, nil
		default:
			return shortNode{n.Key, child}, nil
		}

	case fullNode:
		nn, err := t.delete(n[key[0]], append(prefix, key[0]), key[1:])
		if err != nil {
			return nil, err
		}
		n[key[0]] = nn
		// Check how many non-nil entries are left after deleting and
		// reduce the full node to a short node if only one entry is
		// left. Since n must've contained at least two children
		// before deletion (otherwise it would not be a full node) n
		// can never be reduced to nil.
		//
		// When the loop is done, pos contains the index of the single
		// value that is left in n or -2 if n contains at least two
		// values.
		pos := -1
		for i, cld := range n {
			if cld != nil {
				if pos == -1 {
					pos = i
				} else {
					pos = -2
					break
				}
			}
		}
		if pos >= 0 {
			if pos != 16 {
				// If the remaining entry is a short node, it replaces
				// n and its key gets the missing nibble tacked to the
				// front. This avoids creating an invalid
				// shortNode{..., shortNode{...}}.  Since the entry
				// might not be loaded yet, resolve it just for this
				// check.
				cnode, err := t.resolve(n[pos], prefix, []byte{byte(pos)})
				if err != nil {
					return nil, err
				}
				if cnode, ok := cnode.(shortNode); ok {
					k := append([]byte{byte(pos)}, cnode.Key...)
					return shortNode{k, cnode.Val}, nil
				}
			}
			// Otherwise, n is replaced by a one-nibble short node
			// containing the child.
			return shortNode{[]byte{byte(pos)}, n[pos]}, nil
		}
		// n still contains at least two values and cannot be reduced.
		return n, nil

	case nil:
		return nil, nil

	case hashNode:
		// We've hit a part of the trie that isn't loaded yet. Load
		// the node and delete from it. This leaves all child nodes on
		// the path to the value in the trie.
		//
		// TODO: track whether deletion actually hit a key and keep
		// n as a hash node if it didn't.
		rn, err := t.resolveHash(n, prefix, key)
		if err != nil {
			return nil, err
		}
		return t.delete(rn, prefix, key)

	default:
		panic(fmt.Sprintf("%T: invalid node: %v (%v)", n, n, key))
	}
}

func concat(s1 []byte, s2 ...byte) []byte {
	r := make([]byte, len(s1)+len(s2))
	copy(r, s1)
	copy(r[len(s1):], s2)
	return r
}

func (t *Trie) resolve(n node, prefix, suffix []byte) (node, error) {
	if n, ok := n.(hashNode); ok {
		return t.resolveHash(n, prefix, suffix)
	}
	return n, nil
}

func (t *Trie) resolveHash(n hashNode, prefix, suffix []byte) (node, error) {
	if v, ok := globalCache.Get(n); ok {
		return v, nil
	}
	enc, err := t.db.Get(n)
	if err != nil || enc == nil {
		return nil, &MissingNodeError{
			RootHash:  t.originalRoot,
			NodeHash:  common.BytesToHash(n),
			Key:       compactHexEncode(append(prefix, suffix...)),
			PrefixLen: len(prefix),
			SuffixLen: len(suffix),
		}
	}
	dec := mustDecodeNode(n, enc)
	if dec != nil {
		globalCache.Put(n, dec)
	}
	return dec, nil
}

// Root returns the root hash of the trie.
// Deprecated: use Hash instead.
func (t *Trie) Root() []byte { return t.Hash().Bytes() }

// Hash returns the root hash of the trie. It does not write to the
// database and can be used even if the trie doesn't have one.
func (t *Trie) Hash() common.Hash {
	root, _ := t.hashRoot(nil)
	return common.BytesToHash(root.(hashNode))
}

// Commit writes all nodes to the trie's database.
// Nodes are stored with their sha3 hash as the key.
//
// Committing flushes nodes from memory.
// Subsequent Get calls will load nodes from the database.
func (t *Trie) Commit() (root common.Hash, err error) {
	if t.db == nil {
		panic("Commit called on trie with nil database")
	}
	return t.CommitTo(t.db)
}

// CommitTo writes all nodes to the given database.
// Nodes are stored with their sha3 hash as the key.
//
// Committing flushes nodes from memory. Subsequent Get calls will
// load nodes from the trie's database. Calling code must ensure that
// the changes made to db are written back to the trie's attached
// database before using the trie.
func (t *Trie) CommitTo(db DatabaseWriter) (root common.Hash, err error) {
	n, err := t.hashRoot(db)
	if err != nil {
		return (common.Hash{}), err
	}
	t.root = n
	return common.BytesToHash(n.(hashNode)), nil
}

func (t *Trie) hashRoot(db DatabaseWriter) (node, error) {
	if t.root == nil {
		return hashNode(emptyRoot.Bytes()), nil
	}
	if t.hasher == nil {
		t.hasher = newHasher()
	}
	return t.hasher.hash(t.root, db, true)
}

type hasher struct {
	tmp *bytes.Buffer
	sha hash.Hash
}

func newHasher() *hasher {
	return &hasher{tmp: new(bytes.Buffer), sha: sha3.NewKeccak256()}
}

func (h *hasher) hash(n node, db DatabaseWriter, force bool) (node, error) {
	hashed, err := h.replaceChildren(n, db)
	if err != nil {
		return hashNode{}, err
	}
	if n, err = h.store(hashed, db, force); err != nil {
		return hashNode{}, err
	}
	return n, nil
}

// hashChildren replaces child nodes of n with their hashes if the encoded
// size of the child is larger than a hash.
func (h *hasher) replaceChildren(n node, db DatabaseWriter) (node, error) {
	var err error
	switch n := n.(type) {
	case shortNode:
		n.Key = compactEncode(n.Key)
		if _, ok := n.Val.(valueNode); !ok {
			if n.Val, err = h.hash(n.Val, db, false); err != nil {
				return n, err
			}
		}
		if n.Val == nil {
			// Ensure that nil children are encoded as empty strings.
			n.Val = valueNode(nil)
		}
		return n, nil
	case fullNode:
		for i := 0; i < 16; i++ {
			if n[i] != nil {
				if n[i], err = h.hash(n[i], db, false); err != nil {
					return n, err
				}
			} else {
				// Ensure that nil children are encoded as empty strings.
				n[i] = valueNode(nil)
			}
		}
		if n[16] == nil {
			n[16] = valueNode(nil)
		}
		return n, nil
	default:
		return n, nil
	}
}

func (h *hasher) store(n node, db DatabaseWriter, force bool) (node, error) {
	// Don't store hashes or empty nodes.
	if _, isHash := n.(hashNode); n == nil || isHash {
		return n, nil
	}
	h.tmp.Reset()
	if err := rlp.Encode(h.tmp, n); err != nil {
		panic("encode error: " + err.Error())
	}
	if h.tmp.Len() < 32 && !force {
		// Nodes smaller than 32 bytes are stored inside their parent.
		return n, nil
	}
	// Larger nodes are replaced by their hash and stored in the database.
	h.sha.Reset()
	h.sha.Write(h.tmp.Bytes())
	key := hashNode(h.sha.Sum(nil))
	if db != nil {
		err := db.Put(key, h.tmp.Bytes())
		return key, err
	}
	return key, nil
}
