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

// Contains the Whisper protocol Envelope element. For formal details please see
// the specs at https://github.com/ethereum/wiki/wiki/Whisper-PoC-1-Protocol-Spec#envelopes.

package whisper

import (
	"crypto/ecdsa"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/rlp"
)

// Envelope represents a clear-text data packet to transmit through the Whisper
// network. Its contents may or may not be encrypted and signed.
type Envelope struct {
	Expiry uint32 // Whisper protocol specifies int32, really should be int64
	TTL    uint32 // ^^^^^^
	Topics []Topic
	Data   []byte
	Nonce  uint32

	hash common.Hash // Cached hash of the envelope to avoid rehashing every time
}

// NewEnvelope wraps a Whisper message with expiration and destination data
// included into an envelope for network forwarding.
func NewEnvelope(ttl time.Duration, topics []Topic, msg *Message) *Envelope {
	return &Envelope{
		Expiry: uint32(time.Now().Add(ttl).Unix()),
		TTL:    uint32(ttl.Seconds()),
		Topics: topics,
		Data:   msg.bytes(),
		Nonce:  0,
	}
}

// Seal closes the envelope by spending the requested amount of time as a proof
// of work on hashing the data.
func (self *Envelope) Seal(pow time.Duration) {
	d := make([]byte, 64)
	copy(d[:32], self.rlpWithoutNonce())

	finish, bestBit := time.Now().Add(pow).UnixNano(), 0
	for nonce := uint32(0); time.Now().UnixNano() < finish; {
		for i := 0; i < 1024; i++ {
			binary.BigEndian.PutUint32(d[60:], nonce)

			firstBit := common.FirstBitSet(common.BigD(crypto.Keccak256(d)))
			if firstBit > bestBit {
				self.Nonce, bestBit = nonce, firstBit
			}
			nonce++
		}
	}
}

// rlpWithoutNonce returns the RLP encoded envelope contents, except the nonce.
func (self *Envelope) rlpWithoutNonce() []byte {
	enc, _ := rlp.EncodeToBytes([]interface{}{self.Expiry, self.TTL, self.Topics, self.Data})
	return enc
}

// Open extracts the message contained within a potentially encrypted envelope.
func (self *Envelope) Open(key *ecdsa.PrivateKey) (msg *Message, err error) {
	// Split open the payload into a message construct
	data := self.Data

	message := &Message{
		Flags: data[0],
		Sent:  time.Unix(int64(self.Expiry-self.TTL), 0),
		TTL:   time.Duration(self.TTL) * time.Second,
		Hash:  self.Hash(),
	}
	data = data[1:]

	if message.Flags&signatureFlag == signatureFlag {
		if len(data) < signatureLength {
			return nil, fmt.Errorf("unable to open envelope. First bit set but len(data) < len(signature)")
		}
		message.Signature, data = data[:signatureLength], data[signatureLength:]
	}
	message.Payload = data

	// Decrypt the message, if requested
	if key == nil {
		return message, nil
	}
	err = message.decrypt(key)
	switch err {
	case nil:
		return message, nil

	case ecies.ErrInvalidPublicKey: // Payload isn't encrypted
		return message, err

	default:
		return nil, fmt.Errorf("unable to open envelope, decrypt failed: %v", err)
	}
}

// Hash returns the SHA3 hash of the envelope, calculating it if not yet done.
func (self *Envelope) Hash() common.Hash {
	if (self.hash == common.Hash{}) {
		enc, _ := rlp.EncodeToBytes(self)
		self.hash = crypto.Keccak256Hash(enc)
	}
	return self.hash
}

// DecodeRLP decodes an Envelope from an RLP data stream.
func (self *Envelope) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	// The decoding of Envelope uses the struct fields but also needs
	// to compute the hash of the whole RLP-encoded envelope. This
	// type has the same structure as Envelope but is not an
	// rlp.Decoder so we can reuse the Envelope struct definition.
	type rlpenv Envelope
	if err := rlp.DecodeBytes(raw, (*rlpenv)(self)); err != nil {
		return err
	}
	self.hash = crypto.Keccak256Hash(raw)
	return nil
}
