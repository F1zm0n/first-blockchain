package node

import (
	"testing"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/types"
	"github.com/F1zm0n/blocker/util"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

func randomBlock(t *testing.T, chain Chain) *proto.Block {
	var (
		privKey        = crypto.NewPrivateKey()
		b              = util.RandomBlock()
		prevBlock, err = chain.GetBlockByHeight(chain.Height())
	)
	require.NoError(t, err)
	b.Header.PreviousHash = types.HashBlock(prevBlock)

	types.SignBlock(privKey, b)
	return b
}

func TestNewChain(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		r     = require.New(t)
	)
	r.Equal(0, chain.Height())
	block, err := chain.GetBlockByHeight(0)
	r.NoError(err)
	r.NotNil(block)
}

func TestChainHeight(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		count = 99
	)

	for range count {
		b := randomBlock(t, chain)

		err := chain.AddBlock(b)
		require.NoError(t, err)

	}
	require.Equal(t, count, chain.Height())
	require.Equal(t, count+1, chain.headers.Len())
}

func TestAddBlock(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		block = randomBlock(t, chain)
		err   = chain.AddBlock(block)
	)
	require.NoError(t, err)
}

func TestGetBlockByHeight(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		block = randomBlock(t, chain)
		err   = chain.AddBlock(block)
	)
	require.NoError(t, err)

	b, err := chain.GetBlockByHeight(1)
	require.NoError(t, err)
	require.Equal(t, block, b)
}

func TestGetBlock(t *testing.T) {
	var (
		chain = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		block = randomBlock(t, chain)
		err   = chain.AddBlock(block)
	)
	require.NoError(t, err)

	b, err := chain.GetBlockByHash(types.HashBlock(block))
	require.NoError(t, err)
	require.Equal(t, block, b)
}

func TestAddBlockWithTX_InsufficientFunds(t *testing.T) {
	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromHex(baseSeed)
		recipient = crypto.NewPrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("26e6f153537363b3ad45aba45bc2b815c7e8226ea513b63351bd2061aa76c400")
	require.NoError(t, err)

	var (
		inputs = []*proto.TxInput{
			{
				PrevTxHash:   types.HashTransaction(prevTx),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		}
		outputs = []*proto.TxOutput{
			{
				Amount:  1001,
				Address: recipient,
			},
		}
		tx = &proto.Transaction{
			Version: 1,
			Inputs:  inputs,
			Outputs: outputs,
		}
	)
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)

	err = chain.AddBlock(block)
	require.Error(t, err)
	require.EqualError(t, err, "insufficient funds")
}

func TestAddBlockWithTx(t *testing.T) {

	var (
		chain     = NewChain(NewMemoryBlockStore(), NewMemoryHeaderStore(), NewMemoryTxStore(), NewMemoryUTXOStore(), NewMemoryOutputStorage())
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromHex(baseSeed)
		recipient = crypto.NewPrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("26e6f153537363b3ad45aba45bc2b815c7e8226ea513b63351bd2061aa76c400")
	require.NoError(t, err)

	var (
		inputs = []*proto.TxInput{
			{
				PrevTxHash:   types.HashTransaction(prevTx),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		}
		outputs = []*proto.TxOutput{
			{
				Amount:  100,
				Address: recipient,
			},
			{
				Amount:  900,
				Address: privKey.Public().Address().Bytes(),
			},
		}
		tx = &proto.Transaction{
			Version: 1,
			Inputs:  inputs,
			Outputs: outputs,
		}
	)
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)

	require.NoError(t, chain.AddBlock(block))
}

func TestAddBlockWithTxBadger(t *testing.T) {
	var (
		r          = require.New(t)
		blockOpts  = badger.DefaultOptions("/tmp/block/")
		headerOpts = badger.DefaultOptions("/tmp/header/")
		txOpts     = badger.DefaultOptions("/tmp/tx/")
		utxoOpts   = badger.DefaultOptions("/tmp/utxo/")
		outputOpts = badger.DefaultOptions("/tmp/output/")
	)
	blockOpts.Logger = nil
	headerOpts.Logger = nil
	txOpts.Logger = nil
	utxoOpts.Logger = nil
	outputOpts.Logger = nil

	blockDb, err := badger.Open(blockOpts)
	r.NoError(err)
	headerDb, err := badger.Open(headerOpts)
	r.NoError(err)
	txDb, err := badger.Open(txOpts)
	r.NoError(err)
	utxoDb, err := badger.Open(utxoOpts)
	r.NoError(err)
	outputDb, err := badger.Open(outputOpts)
	r.NoError(err)

	t.Cleanup(func() {
		blockDb.DropAll()
		blockDb.Close()

		headerDb.DropAll()
		headerDb.Close()

		txDb.DropAll()
		txDb.Close()

		utxoDb.DropAll()
		utxoDb.Close()

		outputDb.DropAll()
		outputDb.Close()
	})

	var (
		chain = NewChain(
			NewBadgerBlockStore(blockDb),
			NewBadgerHeaderStore(headerDb),
			NewBadgerTxStore(txDb),
			NewBadgerUTXOStore(utxoDb),
			NewBadgerOutputStorage(outputDb),
		)
		block     = randomBlock(t, chain)
		privKey   = crypto.NewPrivateKeyFromHex(baseSeed)
		recipient = crypto.NewPrivateKey().Public().Address().Bytes()
	)

	prevTx, err := chain.txStore.Get("26e6f153537363b3ad45aba45bc2b815c7e8226ea513b63351bd2061aa76c400")
	require.NoError(t, err)

	var (
		inputs = []*proto.TxInput{
			{
				PrevTxHash:   types.HashTransaction(prevTx),
				PrevOutIndex: 0,
				PublicKey:    privKey.Public().Bytes(),
			},
		}
		outputs = []*proto.TxOutput{
			{
				Amount:  100,
				Address: recipient,
			},
			{
				Amount:  900,
				Address: privKey.Public().Address().Bytes(),
			},
		}
		tx = &proto.Transaction{
			Version: 1,
			Inputs:  inputs,
			Outputs: outputs,
		}
	)
	sig := types.SignTransaction(privKey, tx)
	tx.Inputs[0].Signature = sig.Bytes()

	block.Transactions = append(block.Transactions, tx)
	types.SignBlock(privKey, block)

	require.NoError(t, chain.AddBlock(block))
}
