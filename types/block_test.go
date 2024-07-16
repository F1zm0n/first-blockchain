package types

import (
	"testing"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/util"
	"github.com/stretchr/testify/require"
)

func TestCalculateRootHash(t *testing.T) {
	var (
		privateKey = crypto.NewPrivateKey()
		block      = util.RandomBlock()
		tx         = &proto.Transaction{
			Version: 1,
		}
	)
	block.Transactions = append(block.Transactions, tx)

	SignBlock(privateKey, block)

	require.Len(t, block.Header.RootHash, 32)
	require.NoError(t, VerifyRootHash(block))
}

func TestSignBlock(t *testing.T) {
	var (
		r          = require.New(t)
		block      = util.RandomBlock()
		privateKey = crypto.NewPrivateKey()
		pubKey     = privateKey.Public()
	)

	sig := SignBlock(privateKey, block)
	r.Equal(64, len(sig.Bytes()))
	r.True(sig.Verify(pubKey, HashBlock(block)))

	r.Equal(block.PublicKey, pubKey.Bytes())
	r.Equal(block.Signature, sig.Bytes())

}

func TestVerifyBlock(t *testing.T) {
	var (
		r          = require.New(t)
		block      = util.RandomBlock()
		privateKey = crypto.NewPrivateKey()
	)

	SignBlock(privateKey, block)

	r.NoError(VerifyBlock(block))

	invalidPrivKey := crypto.NewPrivateKey()

	block.PublicKey = invalidPrivKey.Public().Bytes()

	r.Error(VerifyBlock(block))
}

func TestHashBlock(t *testing.T) {
	r := require.New(t)
	block := util.RandomBlock()
	hashBlock := HashBlock(block)
	r.Equal(32, len(hashBlock))
}
