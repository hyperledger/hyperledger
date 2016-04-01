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

package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"os"

	"encoding/hex"
	"encoding/json"
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto/ecies"
	"github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/crypto/sha3"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/pborman/uuid"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/ripemd160"
)

func Keccak256(data ...[]byte) []byte {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	return d.Sum(nil)
}

func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Deprecated: For backward compatibility as other packages depend on these
func Sha3(data ...[]byte) []byte          { return Keccak256(data...) }
func Sha3Hash(data ...[]byte) common.Hash { return Keccak256Hash(data...) }

// Creates an ethereum address given the bytes and the nonce
func CreateAddress(b common.Address, nonce uint64) common.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{b, nonce})
	return common.BytesToAddress(Keccak256(data)[12:])
}

func Sha256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}

func Ripemd160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)

	return ripemd.Sum(nil)
}

func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

// New methods using proper ecdsa keys from the stdlib
func ToECDSA(prv []byte) *ecdsa.PrivateKey {
	if len(prv) == 0 {
		return nil
	}

	priv := new(ecdsa.PrivateKey)
	priv.PublicKey.Curve = secp256k1.S256()
	priv.D = common.BigD(prv)
	priv.PublicKey.X, priv.PublicKey.Y = secp256k1.S256().ScalarBaseMult(prv)
	return priv
}

func FromECDSA(prv *ecdsa.PrivateKey) []byte {
	if prv == nil {
		return nil
	}
	return prv.D.Bytes()
}

func ToECDSAPub(pub []byte) *ecdsa.PublicKey {
	if len(pub) == 0 {
		return nil
	}
	x, y := elliptic.Unmarshal(secp256k1.S256(), pub)
	return &ecdsa.PublicKey{secp256k1.S256(), x, y}
}

func FromECDSAPub(pub *ecdsa.PublicKey) []byte {
	if pub == nil || pub.X == nil || pub.Y == nil {
		return nil
	}
	return elliptic.Marshal(secp256k1.S256(), pub.X, pub.Y)
}

// HexToECDSA parses a secp256k1 private key.
func HexToECDSA(hexkey string) (*ecdsa.PrivateKey, error) {
	b, err := hex.DecodeString(hexkey)
	if err != nil {
		return nil, errors.New("invalid hex string")
	}
	if len(b) != 32 {
		return nil, errors.New("invalid length, need 256 bits")
	}
	return ToECDSA(b), nil
}

// LoadECDSA loads a secp256k1 private key from the given file.
// The key data is expected to be hex-encoded.
func LoadECDSA(file string) (*ecdsa.PrivateKey, error) {
	buf := make([]byte, 64)
	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	if _, err := io.ReadFull(fd, buf); err != nil {
		return nil, err
	}

	key, err := hex.DecodeString(string(buf))
	if err != nil {
		return nil, err
	}

	return ToECDSA(key), nil
}

// SaveECDSA saves a secp256k1 private key to the given file with
// restrictive permissions. The key data is saved hex-encoded.
func SaveECDSA(file string, key *ecdsa.PrivateKey) error {
	k := hex.EncodeToString(FromECDSA(key))
	return ioutil.WriteFile(file, []byte(k), 0600)
}

func GenerateKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(secp256k1.S256(), rand.Reader)
}

func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	vint := uint32(v)
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if homestead && s.Cmp(secp256k1.HalfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	if s.Cmp(secp256k1.N) >= 0 {
		return false
	}
	if r.Cmp(secp256k1.N) < 0 && (vint == 27 || vint == 28) {
		return true
	} else {
		return false
	}
}

func SigToPub(hash, sig []byte) (*ecdsa.PublicKey, error) {
	s, err := Ecrecover(hash, sig)
	if err != nil {
		return nil, err
	}

	x, y := elliptic.Unmarshal(secp256k1.S256(), s)
	return &ecdsa.PublicKey{secp256k1.S256(), x, y}, nil
}

func Sign(hash []byte, prv *ecdsa.PrivateKey) (sig []byte, err error) {
	if len(hash) != 32 {
		return nil, fmt.Errorf("hash is required to be exactly 32 bytes (%d)", len(hash))
	}

	seckey := common.LeftPadBytes(prv.D.Bytes(), prv.Params().BitSize/8)
	defer zeroBytes(seckey)
	sig, err = secp256k1.Sign(hash, seckey)
	return
}

func Encrypt(pub *ecdsa.PublicKey, message []byte) ([]byte, error) {
	return ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(pub), message, nil, nil)
}

func Decrypt(prv *ecdsa.PrivateKey, ct []byte) ([]byte, error) {
	key := ecies.ImportECDSA(prv)
	return key.Decrypt(rand.Reader, ct, nil, nil)
}

// Used only by block tests.
func ImportBlockTestKey(privKeyBytes []byte) error {
	ks := NewKeyStorePassphrase(common.DefaultDataDir()+"/keystore", LightScryptN, LightScryptP)
	ecKey := ToECDSA(privKeyBytes)
	key := &Key{
		Id:         uuid.NewRandom(),
		Address:    PubkeyToAddress(ecKey.PublicKey),
		PrivateKey: ecKey,
	}
	err := ks.StoreKey(key, "")
	return err
}

// creates a Key and stores that in the given KeyStore by decrypting a presale key JSON
func ImportPreSaleKey(keyStore KeyStore, keyJSON []byte, password string) (*Key, error) {
	key, err := decryptPreSaleKey(keyJSON, password)
	if err != nil {
		return nil, err
	}
	key.Id = uuid.NewRandom()
	err = keyStore.StoreKey(key, password)
	return key, err
}

func decryptPreSaleKey(fileContent []byte, password string) (key *Key, err error) {
	preSaleKeyStruct := struct {
		EncSeed string
		EthAddr string
		Email   string
		BtcAddr string
	}{}
	err = json.Unmarshal(fileContent, &preSaleKeyStruct)
	if err != nil {
		return nil, err
	}
	encSeedBytes, err := hex.DecodeString(preSaleKeyStruct.EncSeed)
	iv := encSeedBytes[:16]
	cipherText := encSeedBytes[16:]
	/*
		See https://github.com/ethereum/pyethsaletool

		pyethsaletool generates the encryption key from password by
		2000 rounds of PBKDF2 with HMAC-SHA-256 using password as salt (:().
		16 byte key length within PBKDF2 and resulting key is used as AES key
	*/
	passBytes := []byte(password)
	derivedKey := pbkdf2.Key(passBytes, passBytes, 2000, 16, sha256.New)
	plainText, err := aesCBCDecrypt(derivedKey, cipherText, iv)
	if err != nil {
		return nil, err
	}
	ethPriv := Keccak256(plainText)
	ecKey := ToECDSA(ethPriv)
	key = &Key{
		Id:         nil,
		Address:    PubkeyToAddress(ecKey.PublicKey),
		PrivateKey: ecKey,
	}
	derivedAddr := hex.EncodeToString(key.Address.Bytes()) // needed because .Hex() gives leading "0x"
	expectedAddr := preSaleKeyStruct.EthAddr
	if derivedAddr != expectedAddr {
		err = fmt.Errorf("decrypted addr '%s' not equal to expected addr '%s'", derivedAddr, expectedAddr)
	}
	return key, err
}

// AES-128 is selected due to size of encryptKey
func aesCTRXOR(key, inText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	stream := cipher.NewCTR(aesBlock, iv)
	outText := make([]byte, len(inText))
	stream.XORKeyStream(outText, inText)
	return outText, err
}

func aesCBCDecrypt(key, cipherText, iv []byte) ([]byte, error) {
	aesBlock, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	decrypter := cipher.NewCBCDecrypter(aesBlock, iv)
	paddedPlaintext := make([]byte, len(cipherText))
	decrypter.CryptBlocks(paddedPlaintext, cipherText)
	plaintext := PKCS7Unpad(paddedPlaintext)
	if plaintext == nil {
		err = errors.New("Decryption failed: PKCS7Unpad failed after AES decryption")
	}
	return plaintext, err
}

// From https://leanpub.com/gocrypto/read#leanpub-auto-block-cipher-modes
func PKCS7Unpad(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}

	padding := in[len(in)-1]
	if int(padding) > len(in) || padding > aes.BlockSize {
		return nil
	} else if padding == 0 {
		return nil
	}

	for i := len(in) - 1; i > len(in)-int(padding)-1; i-- {
		if in[i] != padding {
			return nil
		}
	}
	return in[:len(in)-int(padding)]
}

func PubkeyToAddress(p ecdsa.PublicKey) common.Address {
	pubBytes := FromECDSAPub(&p)
	return common.BytesToAddress(Keccak256(pubBytes[1:])[12:])
}

func zeroBytes(bytes []byte) {
	for i := range bytes {
		bytes[i] = 0
	}
}
