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

package tests

import (
	"path/filepath"
	"testing"
)

func TestTransactions(t *testing.T) {
	err := RunTransactionTests(filepath.Join(transactionTestDir, "ttTransactionTest.json"), TransSkipTests)
	if err != nil {
		t.Fatal(err)
	}
}

func TestWrongRLPTransactions(t *testing.T) {
	err := RunTransactionTests(filepath.Join(transactionTestDir, "ttWrongRLPTransaction.json"), TransSkipTests)
	if err != nil {
		t.Fatal(err)
	}
}

func Test10MBTransactions(t *testing.T) {
	err := RunTransactionTests(filepath.Join(transactionTestDir, "tt10mbDataField.json"), TransSkipTests)
	if err != nil {
		t.Fatal(err)
	}
}

// homestead tests
func TestHomesteadTransactions(t *testing.T) {
	err := RunTransactionTests(filepath.Join(transactionTestDir, "Homestead", "ttTransactionTest.json"), TransSkipTests)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHomesteadWrongRLPTransactions(t *testing.T) {
	err := RunTransactionTests(filepath.Join(transactionTestDir, "Homestead", "ttWrongRLPTransaction.json"), TransSkipTests)
	if err != nil {
		t.Fatal(err)
	}
}

func TestHomestead10MBTransactions(t *testing.T) {
	err := RunTransactionTests(filepath.Join(transactionTestDir, "Homestead", "tt10mbDataField.json"), TransSkipTests)
	if err != nil {
		t.Fatal(err)
	}
}
