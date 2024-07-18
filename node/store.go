package node

import (
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/types"
	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
)

type TxStore interface {
	Put(tx *proto.Transaction) error
	Get(hash string) (*proto.Transaction, error)
}

type BadgerTxStore struct {
	db *badger.DB
}

func NewBadgerTxStore(db *badger.DB) TxStore {
	return &BadgerTxStore{
		db: db,
	}
}

func (b *BadgerTxStore) Get(hash string) (*proto.Transaction, error) {
	var valNot []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hash))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			valNot = append(valNot, val...)
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var tx *proto.Transaction = &proto.Transaction{}
	err = sonic.Unmarshal(valNot, tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (b *BadgerTxStore) Put(tx *proto.Transaction) error {
	return b.db.Update(func(txn *badger.Txn) error {
		bys, err := sonic.Marshal(tx)
		if err != nil {
			return err
		}
		return txn.Set([]byte(hex.EncodeToString(types.HashTransaction(tx))), bys)
	})
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

type BadgerBlockStore struct {
	db *badger.DB
}

func NewBadgerBlockStore(db *badger.DB) BlockStore {
	return &BadgerBlockStore{
		db: db,
	}
}

func (b *BadgerBlockStore) Get(hash string) (*proto.Block, error) {
	var buf []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(hash))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			buf = append(buf, val...)
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var block *proto.Block = &proto.Block{}
	if err := sonic.Unmarshal(buf, block); err != nil {
		return nil, err
	}
	return block, nil
}

func (b *BadgerBlockStore) Put(block *proto.Block) error {
	return b.db.Update(func(txn *badger.Txn) error {
		bys, err := sonic.Marshal(block)
		if err != nil {
			return err
		}
		return txn.Set([]byte(hex.EncodeToString(types.HashBlock(block))), bys)
	})
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

type BadgerUTXOStore struct {
	db *badger.DB
}

func NewBadgerUTXOStore(db *badger.DB) UTXOStore {
	return &BadgerUTXOStore{
		db: db,
	}
}

func (b *BadgerUTXOStore) Get(key string) (*UTXO, error) {
	var buf []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			buf = append(buf, val...)
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var utxo *UTXO = &UTXO{}
	if err := sonic.Unmarshal(buf, utxo); err != nil {
		return nil, err
	}
	return utxo, nil
}

func (b *BadgerUTXOStore) Put(utxo *UTXO) error {
	return b.db.Update(func(txn *badger.Txn) error {
		bys, err := sonic.Marshal(utxo)
		if err != nil {
			return err
		}
		return txn.Set([]byte(fmt.Sprintf("%s_%d", utxo.Hash, utxo.OutIndex)), bys)
	})
}

type HeaderStore interface {
	Put(header *proto.Header) error
	Len() int
	Height() int
	GetByIdx(idx int) (*proto.Header, error)
}

type BadgerHeaderStore struct {
	db  *badger.DB
	len int
}

func calculateLength(db *badger.DB) (int, error) {
	count := 0
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			count += 1
		}
		return nil
	})
	return count, err
}

func NewBadgerHeaderStore(db *badger.DB) HeaderStore {
	len, err := calculateLength(db)
	if err != nil {
		panic(err)
	}
	return &BadgerHeaderStore{
		db:  db,
		len: len,
	}
}

func (b *BadgerHeaderStore) Put(header *proto.Header) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		byt, err := sonic.Marshal(header)
		if err != nil {
			return err
		}
		return txn.Set([]byte(strconv.Itoa(b.len)), byt)
	})
	if err != nil {
		return err
	}
	b.len += 1
	return nil
}

func (b *BadgerHeaderStore) GetByIdx(idx int) (*proto.Header, error) {
	if b.len-1 < idx {
		return nil, errors.New("index out of range")
	}

	var buf []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(strconv.Itoa(idx)))
		if err != nil {
			return err
		}
		if err := item.Value(func(val []byte) error {
			buf = append(buf, val...)
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	var header *proto.Header = &proto.Header{}
	if err := sonic.Unmarshal(buf, header); err != nil {
		return nil, err
	}

	return header, nil
}

func (b *BadgerHeaderStore) Height() int {
	return b.len - 1
}

func (b *BadgerHeaderStore) Len() int {
	return b.len
}

type MemoryHeaderStore struct {
	headers []*proto.Header
}

func NewMemoryHeaderStore() HeaderStore {
	return &MemoryHeaderStore{
		headers: []*proto.Header{},
	}
}

func (h *MemoryHeaderStore) Put(header *proto.Header) error {
	h.headers = append(h.headers, header)
	return nil
}

func (h *MemoryHeaderStore) Len() int {
	return len(h.headers)
}

func (h *MemoryHeaderStore) Height() int {
	return len(h.headers) - 1
}

func (h *MemoryHeaderStore) GetByIdx(idx int) (*proto.Header, error) {
	if idx > h.Height() {
		panic("index out of range")
	}
	return h.headers[idx], nil
}

type OutputStorage interface {
	Put(output *proto.TxOutput) error
	Get(address string) ([]*proto.TxOutput, error)
}

type BadgerOutputStorage struct {
	db *badger.DB
}

func NewBadgerOutputStorage(db *badger.DB) OutputStorage {
	return &BadgerOutputStorage{
		db: db,
	}
}

func (b *BadgerOutputStorage) Get(address string) ([]*proto.TxOutput, error) {
	var buf []byte
	err := b.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(address))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			buf = append(buf, val...)
			return nil
		})
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return []*proto.TxOutput{}, err
	}
	var outputs []*proto.TxOutput = make([]*proto.TxOutput, 0)
	if err := sonic.Unmarshal(buf, &outputs); err != nil {
		return []*proto.TxOutput{}, err
	}

	return outputs, nil
}

func (b *BadgerOutputStorage) Put(output *proto.TxOutput) error {
	addr := hex.EncodeToString(output.Address)
	outputs, err := b.Get(addr)
	if err != nil && !errors.Is(err, badger.ErrKeyNotFound) {
		return err
	}

	err = b.db.Update(func(txn *badger.Txn) error {
		outputs = append(outputs, output)

		bytesOutputs, err := sonic.Marshal(outputs)
		if err != nil {
			return err
		}

		if err := txn.Set([]byte(addr), bytesOutputs); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

type MemoryOutputStorage struct {
	store map[string][]*proto.TxOutput
	mu    sync.RWMutex
}

func NewMemoryOutputStorage() OutputStorage {
	return &MemoryOutputStorage{
		store: map[string][]*proto.TxOutput{},
		mu:    sync.RWMutex{},
	}
}

func (m *MemoryOutputStorage) Get(address string) ([]*proto.TxOutput, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.store[address], nil
}

func (m *MemoryOutputStorage) Put(output *proto.TxOutput) error {
	hash := hex.EncodeToString(output.Address)
	outputs, err := m.Get(hash)
	if err != nil {
		return err
	}
	outputs = append(outputs, output)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[hash] = outputs
	return nil
}
