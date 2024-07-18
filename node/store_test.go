package node

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/types"
	"github.com/F1zm0n/blocker/util"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

func TestBadgerTxStorePut(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/tx")
	opts.Logger = nil
	var (
		r       = require.New(t)
		db, err = badger.Open(opts)
		pk      = crypto.NewPrivateKey()
	)
	r.NoError(err)

	t.Cleanup(func() {
		r.NoError(db.DropAll())
		db.Close()
	})

	var (
		store = NewBadgerTxStore(db)
		tx    = &proto.Transaction{
			Version: 1,
			Inputs:  []*proto.TxInput{},
			Outputs: []*proto.TxOutput{
				{
					Amount:  100,
					Address: pk.Public().Address().Bytes(),
				},
			},
		}
	)

	r.NoError(store.Put(tx))
}

func TestBadgerTxStoreGet(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/tx")
	opts.Logger = nil
	var (
		r       = require.New(t)
		db, err = badger.Open(opts)
		pk      = crypto.NewPrivateKey()
	)
	r.NoError(err)

	t.Cleanup(func() {
		r.NoError(db.DropAll())
		db.Close()
	})

	var (
		store = NewBadgerTxStore(db)
		tx    = &proto.Transaction{
			Version: 1,
			Outputs: []*proto.TxOutput{
				{
					Amount:  100,
					Address: pk.Public().Address().Bytes(),
				},
			},
		}
	)

	r.NoError(store.Put(tx))

	fetchedTx, err := store.Get(hex.EncodeToString(types.HashTransaction(tx)))
	r.NoError(err)
	for i, inp := range fetchedTx.Inputs {
		if inp == tx.Inputs[i] {
			r.Fail("inputs are not equal")
		}
	}
	for i, out := range fetchedTx.Outputs {
		if out == tx.Outputs[i] {
			r.Fail("outputs are not equal")
		}
	}
	r.Equal(fetchedTx.Version, tx.Version)
}

func TestBadgerBlockStorePut(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/block")
	opts.Logger = nil
	var (
		r       = require.New(t)
		block   = util.RandomBlock()
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(db.DropAll())
		db.Close()
	})

	store := NewBadgerBlockStore(db)
	r.NoError(store.Put(block))
}

func TestBadgerBlockStoreGet(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/block")
	opts.Logger = nil
	var (
		r       = require.New(t)
		block   = util.RandomBlock()
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(db.DropAll())
		db.Close()
	})

	store := NewBadgerBlockStore(db)
	r.NoError(store.Put(block))

	fetchedBlock, err := store.Get(hex.EncodeToString(types.HashBlock(block)))
	r.NoError(err)

	r.Equal(fetchedBlock.Header.Height, block.Header.Height)
	r.Equal(fetchedBlock.Header.PreviousHash, block.Header.PreviousHash)
	r.Equal(fetchedBlock.Header.RootHash, block.Header.RootHash)
	r.Equal(fetchedBlock.Header.Version, block.Header.Version)
	r.Equal(fetchedBlock.Header.Timestamp, block.Header.Timestamp)

	r.Equal(fetchedBlock.PublicKey, block.PublicKey)
	r.Equal(fetchedBlock.Signature, block.Signature)

	for i, fetchedTx := range fetchedBlock.Transactions {
		tx := block.Transactions[i]
		for i, inp := range fetchedTx.Inputs {
			if inp == tx.Inputs[i] {
				r.Fail("inputs are not equal")
			}
		}

		for i, out := range fetchedTx.Outputs {
			if out == tx.Outputs[i] {
				r.Fail("outputs are not equal")
			}
		}
		r.Equal(tx.Version, fetchedTx.Version)
	}
}

func TestBadgerUTXOStorePut(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/utxo")
	opts.Logger = nil
	var (
		r    = require.New(t)
		utxo = &UTXO{
			Hash:     hex.EncodeToString(util.RandomHash()),
			OutIndex: 1,
			Amount:   100,
			Spent:    false,
		}
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(db.DropAll())
		db.Close()
	})
	store := NewBadgerUTXOStore(db)

	r.NoError(store.Put(utxo))
}

func TestBadgerUTXOStoreGet(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/utxo")
	opts.Logger = nil
	var (
		r    = require.New(t)
		utxo = &UTXO{
			Hash:     hex.EncodeToString(util.RandomHash()),
			OutIndex: 1,
			Amount:   100,
			Spent:    false,
		}
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		r.NoError(db.DropAll())
		db.Close()
	})
	store := NewBadgerUTXOStore(db)

	r.NoError(store.Put(utxo))

	fetchedUTXO, err := store.Get(fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex))
	r.NoError(err)

	r.Equal(fetchedUTXO, utxo)
}

func TestBadgerHeaderStorePut(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/header")
	opts.Logger = nil
	var (
		r       = require.New(t)
		header  = util.RandomBlock().Header
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})

	store := NewBadgerHeaderStore(db)
	r.NoError(store.Put(header))
}

func TestBadgerHeaderStoreGet(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/header")
	opts.Logger = nil
	var (
		r       = require.New(t)
		header  = util.RandomBlock().Header
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})

	store := NewBadgerHeaderStore(db)
	r.NoError(store.Put(header))

	fetchedHeader, err := store.GetByIdx(0)
	r.NoError(err)

	r.Equal(header.Height, fetchedHeader.Height)
	r.Equal(header.PreviousHash, fetchedHeader.PreviousHash)
	r.Equal(header.RootHash, fetchedHeader.RootHash)
	r.Equal(header.Version, fetchedHeader.Version)
	r.Equal(header.Timestamp, fetchedHeader.Timestamp)
}

func TestBadgerHeaderStoreGetHeight_Len(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/header")
	opts.Logger = nil
	var (
		r       = require.New(t)
		header  = util.RandomBlock().Header
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})

	store := NewBadgerHeaderStore(db)
	r.NoError(store.Put(header))

	fetchedHeader, err := store.GetByIdx(0)
	r.NoError(err)

	r.Equal(header.Height, fetchedHeader.Height)
	r.Equal(header.PreviousHash, fetchedHeader.PreviousHash)
	r.Equal(header.RootHash, fetchedHeader.RootHash)
	r.Equal(header.Version, fetchedHeader.Version)
	r.Equal(header.Timestamp, fetchedHeader.Timestamp)

	height := store.Height()
	r.NoError(err)
	r.Equal(height, 0)

	len := store.Len()
	r.NoError(err)
	r.Equal(len, 1)
}

func TestBadgerHeaderStore_PutMultiple_And_GetHeight_Len(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/badger/test/header")
	opts.Logger = nil
	var (
		r       = require.New(t)
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})

	store := NewBadgerHeaderStore(db)
	len := 20

	for i := range len {
		header := util.RandomBlock().Header
		r.NoError(store.Put(header))

		fetchedHeader, err := store.GetByIdx(i)
		r.NoError(err)

		r.Equal(header.Height, fetchedHeader.Height)
		r.Equal(header.PreviousHash, fetchedHeader.PreviousHash)
		r.Equal(header.RootHash, fetchedHeader.RootHash)
		r.Equal(header.Version, fetchedHeader.Version)
		r.Equal(header.Timestamp, fetchedHeader.Timestamp)
	}

	height := store.Height()
	r.NoError(err)
	r.Equal(height, len-1)

	fetchedLen := store.Len()
	r.NoError(err)
	r.Equal(fetchedLen, len)
}

func TestOutputStoragePut(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/test/output")
	opts.Logger = nil
	var (
		privKey = crypto.NewPrivateKey()
		addr    = privKey.Public().Address()
		output  = &proto.TxOutput{
			Amount:  1000,
			Address: addr.Bytes(),
		}
		r       = require.New(t)
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})

	store := NewBadgerOutputStorage(db)

	r.NoError(store.Put(output))
}

func TestOutputStorageGet(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/test/output")
	opts.Logger = nil
	var (
		privKey = crypto.NewPrivateKey()
		addr    = privKey.Public().Address()
		output  = &proto.TxOutput{
			Amount:  1000,
			Address: addr.Bytes(),
		}
		outputSecond = &proto.TxOutput{
			Amount:  200,
			Address: addr.Bytes(),
		}
		r       = require.New(t)
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})

	store := NewBadgerOutputStorage(db)

	r.NoError(store.Put(output))
	r.NoError(store.Put(outputSecond))

	fetchedOutputs, err := store.Get(addr.String())
	r.NoError(err)

	r.Len(fetchedOutputs, 2)

	r.Equal(fetchedOutputs[0].Amount, output.Amount)
	r.Equal(fetchedOutputs[0].Address, output.Address)

	r.Equal(fetchedOutputs[1].Amount, outputSecond.Amount)
	r.Equal(fetchedOutputs[1].Address, outputSecond.Address)
}
