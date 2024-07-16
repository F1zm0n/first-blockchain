package node

import (
	"encoding/hex"
	"errors"
	"fmt"
	"sync"

	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/types"
)

type TxStore interface {
	Put(tx *proto.Transaction) error
	Get(hash string) (*proto.Transaction, error)
}

type MemoryTxStore struct {
	txs map[string]*proto.Transaction
	mu  sync.RWMutex
}

func NewMemoryTxStore() TxStore {
	return &MemoryTxStore{
		txs: map[string]*proto.Transaction{},
		mu:  sync.RWMutex{},
	}
}

func (m *MemoryTxStore) Get(hash string) (*proto.Transaction, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	v, ok := m.txs[hash]
	if !ok {
		return nil, errors.New("transaction not found")
	}
	return v, nil
}

func (m *MemoryTxStore) Put(tx *proto.Transaction) error {
	hash := hex.EncodeToString(types.HashTransaction(tx))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.txs[hash] = tx
	return nil
}

type BlockStore interface {
	Put(block *proto.Block) error
	Get(hash string) (*proto.Block, error)
}

type MemoryBlockStore struct {
	blocks map[string]*proto.Block
	mu     sync.RWMutex
}

func (m *MemoryBlockStore) Get(hash string) (*proto.Block, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	block, ok := m.blocks[hash]
	if !ok {
		return nil, errors.New("block doesn't exist")
	}
	return block, nil
}

func (m *MemoryBlockStore) Put(block *proto.Block) error {
	hash := hex.EncodeToString(types.HashBlock(block))
	m.mu.Lock()
	defer m.mu.Unlock()
	m.blocks[hash] = block
	return nil
}

func NewMemoryBlockStore() BlockStore {
	return &MemoryBlockStore{
		blocks: map[string]*proto.Block{},
		mu:     sync.RWMutex{},
	}
}

type UTXOStore interface {
	Put(utxo *UTXO) error
	Get(key string) (*UTXO, error)
}

type MemoryUTXOStore struct {
	mu    sync.RWMutex
	store map[string]*UTXO
}

func NewMemoryUTXOStore() UTXOStore {
	return &MemoryUTXOStore{
		mu:    sync.RWMutex{},
		store: map[string]*UTXO{},
	}
}

func (m *MemoryUTXOStore) Get(key string) (*UTXO, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dat, ok := m.store[key]
	if !ok {
		return nil, errors.New("utxo doesn't exist")
	}
	return dat, nil
}

func (m *MemoryUTXOStore) Put(utxo *UTXO) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex)
	m.store[key] = utxo
	return nil
}
