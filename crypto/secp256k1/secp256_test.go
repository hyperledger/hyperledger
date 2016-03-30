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

package secp256k1

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto/randentropy"
)

const TestCount = 1000

func TestPrivkeyGenerate(t *testing.T) {
	_, seckey := GenerateKeyPair()
	if err := VerifySeckeyValidity(seckey); err != nil {
		t.Errorf("seckey not valid: %s", err)
	}
}

func TestSignatureValidity(t *testing.T) {
	pubkey, seckey := GenerateKeyPair()
	msg := randentropy.GetEntropyCSPRNG(32)
	sig, err := Sign(msg, seckey)
	if err != nil {
		t.Errorf("signature error: %s", err)
	}
	compactSigCheck(t, sig)
	if len(pubkey) != 65 {
		t.Errorf("pubkey length mismatch: want: 65 have: %d", len(pubkey))
	}
	if len(seckey) != 32 {
		t.Errorf("seckey length mismatch: want: 32 have: %d", len(seckey))
	}
	if len(sig) != 65 {
		t.Errorf("sig length mismatch: want: 65 have: %d", len(sig))
	}
	recid := int(sig[64])
	if recid > 4 || recid < 0 {
		t.Errorf("sig recid mismatch: want: within 0 to 4 have: %d", int(sig[64]))
	}
}

func TestInvalidRecoveryID(t *testing.T) {
	_, seckey := GenerateKeyPair()
	msg := randentropy.GetEntropyCSPRNG(32)
	sig, _ := Sign(msg, seckey)
	sig[64] = 99
	_, err := RecoverPubkey(msg, sig)
	if err != ErrInvalidRecoveryID {
		t.Fatalf("got %q, want %q", err, ErrInvalidRecoveryID)
	}
}

func TestSignAndRecover(t *testing.T) {
	pubkey1, seckey := GenerateKeyPair()
	msg := randentropy.GetEntropyCSPRNG(32)
	sig, err := Sign(msg, seckey)
	if err != nil {
		t.Errorf("signature error: %s", err)
	}
	pubkey2, err := RecoverPubkey(msg, sig)
	if err != nil {
		t.Errorf("recover error: %s", err)
	}
	if !bytes.Equal(pubkey1, pubkey2) {
		t.Errorf("pubkey mismatch: want: %x have: %x", pubkey1, pubkey2)
	}
}

func TestRandomMessagesWithSameKey(t *testing.T) {
	pubkey, seckey := GenerateKeyPair()
	keys := func() ([]byte, []byte) {
		return pubkey, seckey
	}
	signAndRecoverWithRandomMessages(t, keys)
}

func TestRandomMessagesWithRandomKeys(t *testing.T) {
	keys := func() ([]byte, []byte) {
		pubkey, seckey := GenerateKeyPair()
		return pubkey, seckey
	}
	signAndRecoverWithRandomMessages(t, keys)
}

func signAndRecoverWithRandomMessages(t *testing.T, keys func() ([]byte, []byte)) {
	for i := 0; i < TestCount; i++ {
		pubkey1, seckey := keys()
		msg := randentropy.GetEntropyCSPRNG(32)
		sig, err := Sign(msg, seckey)
		if err != nil {
			t.Fatalf("signature error: %s", err)
		}
		if sig == nil {
			t.Fatal("signature is nil")
		}
		compactSigCheck(t, sig)

		// TODO: why do we flip around the recovery id?
		sig[len(sig)-1] %= 4

		pubkey2, err := RecoverPubkey(msg, sig)
		if err != nil {
			t.Fatalf("recover error: %s", err)
		}
		if pubkey2 == nil {
			t.Error("pubkey is nil")
		}
		if !bytes.Equal(pubkey1, pubkey2) {
			t.Fatalf("pubkey mismatch: want: %x have: %x", pubkey1, pubkey2)
		}
	}
}

func TestRecoveryOfRandomSignature(t *testing.T) {
	pubkey1, seckey := GenerateKeyPair()
	msg := randentropy.GetEntropyCSPRNG(32)
	sig, err := Sign(msg, seckey)
	if err != nil {
		t.Errorf("signature error: %s", err)
	}

	for i := 0; i < TestCount; i++ {
		sig = randSig()
		pubkey2, _ := RecoverPubkey(msg, sig)
		// recovery can sometimes work, but if so should always give wrong pubkey
		if bytes.Equal(pubkey1, pubkey2) {
			t.Fatalf("iteration: %d: pubkey mismatch: do NOT want %x: ", i, pubkey2)
		}
	}
}

func randSig() []byte {
	sig := randentropy.GetEntropyCSPRNG(65)
	sig[32] &= 0x70
	sig[64] %= 4
	return sig
}

func TestRandomMessagesAgainstValidSig(t *testing.T) {
	pubkey1, seckey := GenerateKeyPair()
	msg := randentropy.GetEntropyCSPRNG(32)
	sig, _ := Sign(msg, seckey)

	for i := 0; i < TestCount; i++ {
		msg = randentropy.GetEntropyCSPRNG(32)
		pubkey2, _ := RecoverPubkey(msg, sig)
		// recovery can sometimes work, but if so should always give wrong pubkey
		if bytes.Equal(pubkey1, pubkey2) {
			t.Fatalf("iteration: %d: pubkey mismatch: do NOT want %x: ", i, pubkey2)
		}
	}
}

func TestZeroPrivkey(t *testing.T) {
	zeroedBytes := make([]byte, 32)
	err := VerifySeckeyValidity(zeroedBytes)
	if err == nil {
		t.Errorf("zeroed bytes should have returned error")
	}
}

// Useful when the underlying libsecp256k1 API changes to quickly
// check only recover function without use of signature function
func TestRecoverSanity(t *testing.T) {
	msg, _ := hex.DecodeString("ce0677bb30baa8cf067c88db9811f4333d131bf8bcf12fe7065d211dce971008")
	sig, _ := hex.DecodeString("90f27b8b488db00b00606796d2987f6a5f59ae62ea05effe84fef5b8b0e549984a691139ad57a3f0b906637673aa2f63d1f55cb1a69199d4009eea23ceaddc9301")
	pubkey1, _ := hex.DecodeString("04e32df42865e97135acfb65f3bae71bdc86f4d49150ad6a440b6f15878109880a0a2b2667f7e725ceea70c673093bf67663e0312623c8e091b13cf2c0f11ef652")
	pubkey2, err := RecoverPubkey(msg, sig)
	if err != nil {
		t.Fatalf("recover error: %s", err)
	}
	if !bytes.Equal(pubkey1, pubkey2) {
		t.Errorf("pubkey mismatch: want: %x have: %x", pubkey1, pubkey2)
	}
}

// tests for malleability
// highest bit of signature ECDSA s value must be 0, in the 33th byte
func compactSigCheck(t *testing.T, sig []byte) {
	var b int = int(sig[32])
	if b < 0 {
		t.Errorf("highest bit is negative: %d", b)
	}
	if ((b >> 7) == 1) != ((b & 0x80) == 0x80) {
		t.Errorf("highest bit: %d bit >> 7: %d", b, b>>7)
	}
	if (b & 0x80) == 0x80 {
		t.Errorf("highest bit: %d bit & 0x80: %d", b, b&0x80)
	}
}

// godep go test -v -run=XXX -bench=BenchmarkSign
// add -benchtime=10s to benchmark longer for more accurate average

// to avoid compiler optimizing the benchmarked function call
var err error

func BenchmarkSign(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, seckey := GenerateKeyPair()
		msg := randentropy.GetEntropyCSPRNG(32)
		b.StartTimer()
		_, e := Sign(msg, seckey)
		err = e
		b.StopTimer()
	}
}

//godep go test -v -run=XXX -bench=BenchmarkECRec
func BenchmarkRecover(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, seckey := GenerateKeyPair()
		msg := randentropy.GetEntropyCSPRNG(32)
		sig, _ := Sign(msg, seckey)
		b.StartTimer()
		_, e := RecoverPubkey(msg, sig)
		err = e
		b.StopTimer()
	}
}
