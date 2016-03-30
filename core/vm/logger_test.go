// Copyright 2016 The go-ethereum Authors
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

package vm

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
)

type dummyContractRef struct {
	calledForEach bool
}

func (dummyContractRef) ReturnGas(*big.Int, *big.Int) {}
func (dummyContractRef) Address() common.Address      { return common.Address{} }
func (dummyContractRef) Value() *big.Int              { return new(big.Int) }
func (dummyContractRef) SetCode([]byte)               {}
func (d *dummyContractRef) ForEachStorage(callback func(key, value common.Hash) bool) {
	d.calledForEach = true
}
func (d *dummyContractRef) SubBalance(amount *big.Int) {}
func (d *dummyContractRef) AddBalance(amount *big.Int) {}
func (d *dummyContractRef) SetBalance(*big.Int)        {}
func (d *dummyContractRef) SetNonce(uint64)            {}
func (d *dummyContractRef) Balance() *big.Int          { return new(big.Int) }

type dummyEnv struct {
	*Env
	ref *dummyContractRef
}

func newDummyEnv(ref *dummyContractRef) *dummyEnv {
	return &dummyEnv{
		Env: NewEnv(),
		ref: ref,
	}
}
func (d dummyEnv) GetAccount(common.Address) Account {
	return d.ref
}
func (d dummyEnv) AddStructLog(StructLog) {}

func TestStoreCapture(t *testing.T) {
	var (
		env      = NewEnv()
		logger   = newLogger(LogConfig{Collector: env}, env)
		mem      = NewMemory()
		stack    = newstack()
		contract = NewContract(&dummyContractRef{}, &dummyContractRef{}, new(big.Int), new(big.Int), new(big.Int))
	)
	stack.push(big.NewInt(1))
	stack.push(big.NewInt(0))

	var index common.Hash

	logger.captureState(0, SSTORE, new(big.Int), new(big.Int), mem, stack, contract, nil)
	if len(logger.changedValues[contract.Address()]) == 0 {
		t.Fatalf("expected exactly 1 changed value on address %x, got %d", contract.Address(), len(logger.changedValues[contract.Address()]))
	}

	exp := common.BigToHash(big.NewInt(1))
	if logger.changedValues[contract.Address()][index] != exp {
		t.Errorf("expected %x, got %x", exp, logger.changedValues[contract.Address()][index])
	}
}

func TestStorageCapture(t *testing.T) {
	t.Skip("implementing this function is difficult. it requires all sort of interfaces to be implemented which isn't trivial. The value (the actual test) isn't worth it")
	var (
		ref      = &dummyContractRef{}
		contract = NewContract(ref, ref, new(big.Int), new(big.Int), new(big.Int))
		env      = newDummyEnv(ref)
		logger   = newLogger(LogConfig{Collector: env}, env)
		mem      = NewMemory()
		stack    = newstack()
	)

	logger.captureState(0, STOP, new(big.Int), new(big.Int), mem, stack, contract, nil)
	if ref.calledForEach {
		t.Error("didn't expect for each to be called")
	}

	logger = newLogger(LogConfig{Collector: env, FullStorage: true}, env)
	logger.captureState(0, STOP, new(big.Int), new(big.Int), mem, stack, contract, nil)
	if !ref.calledForEach {
		t.Error("expected for each to be called")
	}
}
