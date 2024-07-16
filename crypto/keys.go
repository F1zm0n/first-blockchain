package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
)

const (
	PrivateKeyLen = 64
	SignatureLen  = 64
	PubKeyLen     = 32
	seedLen       = 32
	AddressLen    = 20
)

type PrivateKey struct {
	value ed25519.PrivateKey
}

func NewPrivateKeyFromHex(s string) PrivateKey {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return NewPrivateKeyFromSeed(b)
}

func NewPrivateKeyFromSeed(seed []byte) PrivateKey {
	if len(seed) != seedLen {
		log.Panicf("invalid seed length, must be %d", seedLen)
	}
	return PrivateKey{value: ed25519.NewKeyFromSeed(seed)}
}

func NewPrivateKey() PrivateKey {
	seed := make([]byte, seedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	if err != nil {
		panic(err)
	}
	return PrivateKey{value: ed25519.NewKeyFromSeed(seed)}
}

func (p PrivateKey) Bytes() []byte {
	return p.value
}

func (p PrivateKey) Sign(msg []byte) Signature {
	return Signature{value: ed25519.Sign(p.value, msg)}
}

type PublicKey struct {
	value ed25519.PublicKey
}

func PublicKeyFromBytes(pk []byte) PublicKey {
	if len(pk) != PubKeyLen {
		log.Panicf("public key length is invalid %d, should be %d", len(pk), PubKeyLen)
	}

	return PublicKey{value: pk}
}

func (p PublicKey) Address() Address {
	return Address{
		value: p.value[len(p.value)-AddressLen:],
	}
}

func (p PrivateKey) Public() PublicKey {
	b := make([]byte, PubKeyLen)
	copy(b, p.value[32:])

	return PublicKey{
		value: b,
	}
}

func (p PublicKey) Bytes() []byte {
	return p.value
}

type Signature struct {
	value []byte
}

func SignatureFromBytes(sig []byte) Signature {
	if len(sig) != SignatureLen {
		log.Panicf("signature length is invalid %d, should be %d", len(sig), SignatureLen)
	}
	return Signature{value: sig}
}

func (s Signature) Bytes() []byte {
	return s.value
}

func (s Signature) Verify(publicKey PublicKey, msg []byte) bool {
	return ed25519.Verify(publicKey.value, msg, s.value)
}

type Address struct {
	value []byte
}

func NewAddressFromBytes(b []byte) Address {
	if len(b) != AddressLen {
		log.Panicf("address length is %d, should be %d", len(b), AddressLen)
	}
	return Address{
		value: b,
	}
}

func (a Address) String() string {
	return hex.EncodeToString(a.value)
}

func (a Address) Bytes() []byte {
	return a.value
}
