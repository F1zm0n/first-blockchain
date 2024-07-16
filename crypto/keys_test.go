package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGeneratePrivateKey(t *testing.T) {
	r := require.New(t)
	privateKey := NewPrivateKey()
	r.Equal(len(privateKey.Bytes()), PrivateKeyLen)
	pubKey := privateKey.Public()
	r.Equal(len(pubKey.Bytes()), PubKeyLen)
}

func TestNewPrivateKeyFromHex(t *testing.T) {
	r := require.New(t)

	seed := make([]byte, seedLen)
	_, err := io.ReadFull(rand.Reader, seed)
	r.NoError(err)
	seedHex := hex.EncodeToString(seed)

	privateKey := NewPrivateKeyFromHex(seedHex)
	r.Equal(PrivateKeyLen, len(privateKey.Bytes()))
	address := privateKey.Public().Address()
	addressHex := address.String()

	r.Equal(addressHex, address.String())
}

func TestPrivateKeySign(t *testing.T) {

	r := require.New(t)
	privateKey := NewPrivateKey()
	pubKey := privateKey.Public()
	msg := []byte("foo bar baz")

	// test with correct data
	sig := privateKey.Sign(msg)
	r.True(sig.Verify(pubKey, msg))
	// test with invalid msg
	r.False(sig.Verify(pubKey, []byte("invalid value")))
	// test with invalid public value
	invalidPublicKey := NewPrivateKey().Public()
	r.False(sig.Verify(invalidPublicKey, msg))
}

func TestPublicKeyToAddress(t *testing.T) {
	r := require.New(t)
	privateKey := NewPrivateKey()
	pubKey := privateKey.Public()
	address := pubKey.Address()
	r.Equal(len(address.Bytes()), AddressLen)
}
