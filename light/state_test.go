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

package light

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/ethdb"
	"github.com/ethereum/go-ethereum/trie"
	"golang.org/x/net/context"
)

type testOdr struct {
	OdrBackend
	sdb, ldb ethdb.Database
}

func (odr *testOdr) Database() ethdb.Database {
	return odr.ldb
}

func (odr *testOdr) Retrieve(ctx context.Context, req OdrRequest) error {
	switch req := req.(type) {
	case *TrieRequest:
		t, _ := trie.New(req.root, odr.sdb)
		req.proof = t.Prove(req.key)
		trie.ClearGlobalCache()
	case *NodeDataRequest:
		req.data, _ = odr.sdb.Get(req.hash[:])
	}
	req.StoreResult(odr.ldb)
	return nil
}

func makeTestState() (common.Hash, ethdb.Database) {
	sdb, _ := ethdb.NewMemDatabase()
	st, _ := state.New(common.Hash{}, sdb)
	for i := byte(0); i < 100; i++ {
		so := st.GetOrNewStateObject(common.Address{i})
		for j := byte(0); j < 100; j++ {
			val := common.Hash{i, j}
			so.SetState(common.Hash{j}, val)
			so.SetNonce(100)
		}
		so.AddBalance(big.NewInt(int64(i)))
		so.SetCode([]byte{i, i, i})
		so.Update()
		st.UpdateStateObject(so)
	}
	root, _ := st.Commit()
	return root, sdb
}

func TestLightStateOdr(t *testing.T) {
	root, sdb := makeTestState()
	ldb, _ := ethdb.NewMemDatabase()
	odr := &testOdr{sdb: sdb, ldb: ldb}
	ls := NewLightState(root, odr)
	ctx := context.Background()
	trie.ClearGlobalCache()

	for i := byte(0); i < 100; i++ {
		addr := common.Address{i}
		err := ls.AddBalance(ctx, addr, big.NewInt(1000))
		if err != nil {
			t.Fatalf("Error adding balance to acc[%d]: %v", i, err)
		}
		err = ls.SetState(ctx, addr, common.Hash{100}, common.Hash{i, 100})
		if err != nil {
			t.Fatalf("Error setting storage of acc[%d]: %v", i, err)
		}
	}

	addr := common.Address{100}
	_, err := ls.CreateStateObject(ctx, addr)
	if err != nil {
		t.Fatalf("Error creating state object: %v", err)
	}
	err = ls.SetCode(ctx, addr, []byte{100, 100, 100})
	if err != nil {
		t.Fatalf("Error setting code: %v", err)
	}
	err = ls.AddBalance(ctx, addr, big.NewInt(1100))
	if err != nil {
		t.Fatalf("Error adding balance to acc[100]: %v", err)
	}
	for j := byte(0); j < 101; j++ {
		err = ls.SetState(ctx, addr, common.Hash{j}, common.Hash{100, j})
		if err != nil {
			t.Fatalf("Error setting storage of acc[100]: %v", err)
		}
	}
	err = ls.SetNonce(ctx, addr, 100)
	if err != nil {
		t.Fatalf("Error setting nonce for acc[100]: %v", err)
	}

	for i := byte(0); i < 101; i++ {
		addr := common.Address{i}

		bal, err := ls.GetBalance(ctx, addr)
		if err != nil {
			t.Fatalf("Error getting balance of acc[%d]: %v", i, err)
		}
		if bal.Int64() != int64(i)+1000 {
			t.Fatalf("Incorrect balance at acc[%d]: expected %v, got %v", i, int64(i)+1000, bal.Int64())
		}

		nonce, err := ls.GetNonce(ctx, addr)
		if err != nil {
			t.Fatalf("Error getting nonce of acc[%d]: %v", i, err)
		}
		if nonce != 100 {
			t.Fatalf("Incorrect nonce at acc[%d]: expected %v, got %v", i, 100, nonce)
		}

		code, err := ls.GetCode(ctx, addr)
		exp := []byte{i, i, i}
		if err != nil {
			t.Fatalf("Error getting code of acc[%d]: %v", i, err)
		}
		if !bytes.Equal(code, exp) {
			t.Fatalf("Incorrect code at acc[%d]: expected %v, got %v", i, exp, code)
		}

		for j := byte(0); j < 101; j++ {
			exp := common.Hash{i, j}
			val, err := ls.GetState(ctx, addr, common.Hash{j})
			if err != nil {
				t.Fatalf("Error retrieving acc[%d].storage[%d]: %v", i, j, err)
			}
			if val != exp {
				t.Fatalf("Retrieved wrong value from acc[%d].storage[%d]: expected %04x, got %04x", i, j, exp, val)
			}
		}
	}
}

func TestLightStateSetCopy(t *testing.T) {
	root, sdb := makeTestState()
	ldb, _ := ethdb.NewMemDatabase()
	odr := &testOdr{sdb: sdb, ldb: ldb}
	ls := NewLightState(root, odr)
	ctx := context.Background()
	trie.ClearGlobalCache()

	for i := byte(0); i < 100; i++ {
		addr := common.Address{i}
		err := ls.AddBalance(ctx, addr, big.NewInt(1000))
		if err != nil {
			t.Fatalf("Error adding balance to acc[%d]: %v", i, err)
		}
		err = ls.SetState(ctx, addr, common.Hash{100}, common.Hash{i, 100})
		if err != nil {
			t.Fatalf("Error setting storage of acc[%d]: %v", i, err)
		}
	}

	ls2 := ls.Copy()

	for i := byte(0); i < 100; i++ {
		addr := common.Address{i}
		err := ls2.AddBalance(ctx, addr, big.NewInt(1000))
		if err != nil {
			t.Fatalf("Error adding balance to acc[%d]: %v", i, err)
		}
		err = ls2.SetState(ctx, addr, common.Hash{100}, common.Hash{i, 200})
		if err != nil {
			t.Fatalf("Error setting storage of acc[%d]: %v", i, err)
		}
	}

	lsx := ls.Copy()
	ls.Set(ls2)
	ls2.Set(lsx)

	for i := byte(0); i < 100; i++ {
		addr := common.Address{i}
		// check balance in ls
		bal, err := ls.GetBalance(ctx, addr)
		if err != nil {
			t.Fatalf("Error getting balance to acc[%d]: %v", i, err)
		}
		if bal.Int64() != int64(i)+2000 {
			t.Fatalf("Incorrect balance at ls.acc[%d]: expected %v, got %v", i, int64(i)+1000, bal.Int64())
		}
		// check balance in ls2
		bal, err = ls2.GetBalance(ctx, addr)
		if err != nil {
			t.Fatalf("Error getting balance to acc[%d]: %v", i, err)
		}
		if bal.Int64() != int64(i)+1000 {
			t.Fatalf("Incorrect balance at ls.acc[%d]: expected %v, got %v", i, int64(i)+1000, bal.Int64())
		}
		// check storage in ls
		exp := common.Hash{i, 200}
		val, err := ls.GetState(ctx, addr, common.Hash{100})
		if err != nil {
			t.Fatalf("Error retrieving acc[%d].storage[100]: %v", i, err)
		}
		if val != exp {
			t.Fatalf("Retrieved wrong value from acc[%d].storage[100]: expected %04x, got %04x", i, exp, val)
		}
		// check storage in ls2
		exp = common.Hash{i, 100}
		val, err = ls2.GetState(ctx, addr, common.Hash{100})
		if err != nil {
			t.Fatalf("Error retrieving acc[%d].storage[100]: %v", i, err)
		}
		if val != exp {
			t.Fatalf("Retrieved wrong value from acc[%d].storage[100]: expected %04x, got %04x", i, exp, val)
		}
	}
}

func TestLightStateDelete(t *testing.T) {
	root, sdb := makeTestState()
	ldb, _ := ethdb.NewMemDatabase()
	odr := &testOdr{sdb: sdb, ldb: ldb}
	ls := NewLightState(root, odr)
	ctx := context.Background()
	trie.ClearGlobalCache()

	addr := common.Address{42}

	b, err := ls.HasAccount(ctx, addr)
	if err != nil {
		t.Fatalf("HasAccount error: %v", err)
	}
	if !b {
		t.Fatalf("HasAccount returned false, expected true")
	}

	b, err = ls.IsDeleted(ctx, addr)
	if err != nil {
		t.Fatalf("IsDeleted error: %v", err)
	}
	if b {
		t.Fatalf("IsDeleted returned true, expected false")
	}

	ls.Delete(ctx, addr)

	b, err = ls.IsDeleted(ctx, addr)
	if err != nil {
		t.Fatalf("IsDeleted error: %v", err)
	}
	if !b {
		t.Fatalf("IsDeleted returned false, expected true")
	}
}
