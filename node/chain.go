package node

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/types"
)

const baseSeed = "e5f05fafb2b46ede8984a1b3c1ae585203b524c4a1abc40a077655d09893bd2b"

type UTXO struct {
	Hash     string
	OutIndex int
	Amount   int64
	Spent    bool
}

type Chain struct {
	blockStore    BlockStore
	headers       HeaderStore
	utxoStore     UTXOStore
	txStore       TxStore
	outputStorage OutputStorage
}

func NewChain(bs BlockStore, hs HeaderStore, txx TxStore, us UTXOStore, os OutputStorage) Chain {
	chain := Chain{
		blockStore:    bs,
		headers:       hs,
		utxoStore:     us,
		txStore:       txx,
		outputStorage: os,
	}
	chain.addBlock(chain.createGensisBlock())
	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(b *proto.Block) error {
	if err := c.ValidateBlock(b); err != nil {
		return err
	}
	return c.addBlock(b)
}

func (c *Chain) addBlock(b *proto.Block) error {
	if err := c.headers.Put(b.Header); err != nil {
		return err
	}
	for _, tx := range b.Transactions {
		if err := c.txStore.Put(tx); err != nil {
			return err
		}
		hash := hex.EncodeToString(types.HashTransaction(tx))
		for i, output := range tx.Outputs {
			utxo := &UTXO{
				Hash:     hash,
				OutIndex: i,
				Amount:   output.Amount,
				Spent:    false,
			}
			if err := c.utxoStore.Put(utxo); err != nil {
				return err
			}

		}
		for _, input := range tx.Inputs {
			key := fmt.Sprintf("%s_%d", hex.EncodeToString(input.PrevTxHash), input.PrevOutIndex)
			utxo, err := c.utxoStore.Get(key)
			if err != nil {
				return err
			}
			utxo.Spent = true

			if err = c.utxoStore.Put(utxo); err != nil {
				return err
			}
		}
	}
	return c.blockStore.Put(b)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashStr := hex.EncodeToString(hash)
	return c.blockStore.Get(hashStr)
}

func (c *Chain) GetBlockByHeight(h int) (*proto.Block, error) {
	if c.Height() < h {
		return nil, errors.New("block doesn't exists")
	}

	header, err := c.headers.GetByIdx(h)
	if err != nil {
		return nil, err
	}

	hash := types.HashHeader(header)

	block, err := c.GetBlockByHash(hash)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (c *Chain) ValidateBlock(b *proto.Block) error {
	if err := types.VerifyBlock(b); err != nil {
		return err
	}

	currBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}

	currHash := types.HashBlock(currBlock)
	if !bytes.Equal(currHash, b.Header.PreviousHash) {
		return errors.New("invalid previous block hash")
	}

	for _, tx := range b.Transactions {
		if err := c.ValidateTransaction(tx); err != nil {
			return err
		}
	}

	return nil
}

func (c *Chain) ValidateTransaction(tx *proto.Transaction) error {
	// verify the signature
	if !types.VerifyTransaction(tx) {
		return errors.New("invalid transaction signature")
	}
	// verify is outputs are spent
	sumInputs := 0
	for i := range len(tx.Inputs) {
		key := fmt.Sprintf("%s_%d", hex.EncodeToString(tx.Inputs[i].PrevTxHash), i)
		utxo, err := c.utxoStore.Get(key)
		if err != nil {
			return err
		}
		if utxo.Spent {
			return errors.New("output of transaction is already spent")
		}
		sumInputs += int(utxo.Amount)
	}
	sumOutput := 0
	for _, out := range tx.Outputs {
		sumOutput += int(out.Amount)
	}

	if sumInputs < sumOutput {
		return errors.New("insufficient funds")
	}

	return nil
}

func (c *Chain) createGensisBlock() *proto.Block {
	privKey := crypto.NewPrivateKeyFromHex(baseSeed)
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{},
		Outputs: []*proto.TxOutput{
			{
				Amount:  1000,
				Address: privKey.Public().Address().Bytes(),
			},
		},
	}

	block.Transactions = append(block.Transactions, tx)

	types.SignBlock(privKey, block)
	return block
}
