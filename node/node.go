package node

import (
	"context"
	"encoding/hex"
	"errors"
	"log"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/types"
	"github.com/bytedance/sonic"
	"github.com/dgraph-io/badger/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/peer"
)

const blockDur = time.Second * 5

type MemPool interface {
	Clear() ([]*proto.Transaction, error)
	Len() int
	Has(tx *proto.Transaction) error
	Put(tx *proto.Transaction) error
}

type BadgerMemPool struct {
	len int
	db  *badger.DB
}

func NewBadgerMemPool(db *badger.DB) MemPool {
	len, err := calculateLength(db)
	if err != nil {
		panic(err)
	}
	return &BadgerMemPool{
		len: len,
		db:  db,
	}
}

func (b *BadgerMemPool) Clear() ([]*proto.Transaction, error) {
	var txs []*proto.Transaction
	counter := 0
	err := b.db.Update(func(txn *badger.Txn) error {

		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			err := item.Value(func(val []byte) error {
				var tx *proto.Transaction = &proto.Transaction{}
				if err := sonic.Unmarshal(val, tx); err != nil {
					return err
				}
				txs = append(txs, tx)
				return nil
			})
			if err != nil {
				return err
			}

			err = txn.Delete(it.Item().Key())
			if err != nil {
				return err
			}
			counter += 1
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	b.len -= counter
	return txs, nil
}

func (b *BadgerMemPool) Has(tx *proto.Transaction) error {
	err := b.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(hex.EncodeToString(types.HashTransaction(tx))))
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (b *BadgerMemPool) Len() int {
	return b.len
}

func (b *BadgerMemPool) Put(tx *proto.Transaction) error {
	err := b.db.Update(func(txn *badger.Txn) error {
		byts, err := sonic.Marshal(tx)
		if err != nil {
			return err
		}
		return txn.Set([]byte(hex.EncodeToString(types.HashTransaction(tx))), byts)
	})
	if err != nil {
		return err
	}
	b.len += 1
	return nil
}

type MemoryMemPool struct {
	pool map[string]*proto.Transaction
	mu   sync.RWMutex
}

func NewMemoryMemPool() MemPool {
	return &MemoryMemPool{
		pool: map[string]*proto.Transaction{},
		mu:   sync.RWMutex{},
	}
}

func (m *MemoryMemPool) Clear() ([]*proto.Transaction, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	p := make([]*proto.Transaction, 0, len(m.pool))
	for k, v := range m.pool {
		delete(m.pool, k)
		p = append(p, v)
	}

	return p, nil
}

func (m *MemoryMemPool) Len() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.pool)
}

func (m *MemoryMemPool) Has(tx *proto.Transaction) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.pool[hex.EncodeToString(types.HashTransaction(tx))]
	if !ok {
		return errors.New("transaction doesn't exist")
	}
	return nil
}

func (m *MemoryMemPool) Put(tx *proto.Transaction) error {
	err := m.Has(tx)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.pool[hex.EncodeToString(types.HashTransaction(tx))] = tx
	return nil
}

type ServerConfig struct {
	Version    string
	ListenAddr string
	PrivateKey *crypto.PrivateKey
}

type Node struct {
	ServerConfig
	mu      sync.RWMutex
	peers   map[proto.NodeClient]*proto.Version
	memPool MemPool
	sl      *slog.Logger
	proto.UnimplementedNodeServer
}

func New(sl *slog.Logger, memPool MemPool, cfg ServerConfig) Node {
	return Node{
		ServerConfig:            cfg,
		mu:                      sync.RWMutex{},
		peers:                   make(map[proto.NodeClient]*proto.Version),
		memPool:                 memPool,
		sl:                      sl.With("we", cfg.ListenAddr),
		UnimplementedNodeServer: proto.UnimplementedNodeServer{},
	}
}

func (n *Node) HandleTransaction(ctx context.Context, req *proto.Transaction) (*proto.Ack, error) {
	p, _ := peer.FromContext(ctx)

	err := n.memPool.Put(req)
	if err != nil {
		return nil, err
	}
	n.sl.Info("received tx", slog.String("remote_node", p.Addr.String()))
	go func() {
		if err := n.broadcast(context.Background(), req); err != nil {
			n.sl.Error("broadcast error", slog.String("error", err.Error()))
		}
	}()

	return &proto.Ack{}, nil
}

func (n *Node) Handshake(ctx context.Context, req *proto.Version) (*proto.Version, error) {
	c, err := makeNodeClient(req.ListenAddr)
	if err != nil {
		return nil, err
	}
	n.addPeer(c, req)
	return n.getVersion(), nil
}

func (n *Node) validatorLoop() {

	l := slog.With(
		slog.Any("public_key",
			n.PrivateKey.Public().Address().String()),
		slog.Duration("block_time", blockDur),
	)
	l.Info(
		"starting validator loop",
	)
	ticker := time.NewTicker(blockDur)
	for {
		<-ticker.C
		txx, err := n.memPool.Clear()
		if err != nil {
			l.Error("error clearing memory pool", slog.String("error", err.Error()))
			continue
		}
		l.Warn("time to create a new block", slog.Int("len_tx", len(txx)))
	}
}

func (n *Node) broadcast(ctx context.Context, msg any) error {
	for peer := range n.peers {
		switch v := msg.(type) {
		case *proto.Transaction:
			if _, err := peer.HandleTransaction(ctx, v); err != nil {
				return err
			}
		}
	}
	return nil
}

func makeNodeClient(listenAddr string) (proto.NodeClient, error) {
	client, err := grpc.NewClient(
		listenAddr,
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}...)
	if err != nil {
		return nil, err
	}

	newCli := proto.NewNodeClient(client)

	return newCli, err
}

func (n *Node) Start(bootstrapNodes []string) error {
	ln, err := net.Listen("tcp", n.ListenAddr)
	if err != nil {
		log.Fatal(err)
	}
	grpcServ := grpc.NewServer([]grpc.ServerOption{}...)

	proto.RegisterNodeServer(grpcServ, n)

	if len(bootstrapNodes) > 0 {
		go func() {
			if err := n.bootstrapNetwork(bootstrapNodes); err != nil {
				log.Fatal(err)
			}
		}()
	}
	if n.PrivateKey != nil {
		go n.validatorLoop()
	}

	n.sl.Info("server is starting")
	return grpcServ.Serve(ln)
}

func (n *Node) addPeer(c proto.NodeClient, ver *proto.Version) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.peers[c] = ver
	if len(ver.PeerList) > 0 {
		go n.bootstrapNetwork(ver.PeerList)
	}
	n.sl.Info(
		"new peer successfully connected",
		slog.Int("height", int(ver.Height)),
		slog.String("remote_node", ver.ListenAddr),
	)
}

func (n *Node) bootstrapNetwork(addrs []string) error {
	for _, addr := range addrs {
		if !n.canConnectWith(addr) {
			continue
		}

		n.sl.Debug("dialing remote node", slog.String("remote_node", addr))

		c, v, err := n.dialNode(addr)
		if err != nil {
			return err
		}

		n.addPeer(c, v)
	}
	return nil
}

func (n *Node) deletePeer(c proto.NodeClient) {
	n.mu.Lock()
	defer n.mu.Unlock()
	delete(n.peers, c)
}

func (n *Node) dialNode(addr string) (proto.NodeClient, *proto.Version, error) {
	c, err := makeNodeClient(addr)
	if err != nil {
		return nil, nil, err
	}

	v, err := c.Handshake(context.Background(), n.getVersion())
	if err != nil {
		return nil, nil, err
	}

	return c, v, nil
}

func (n *Node) getVersion() *proto.Version {
	return &proto.Version{
		Version:    n.Version,
		Height:     0,
		ListenAddr: n.ListenAddr,
		PeerList:   n.getPeerList(),
	}
}

func (n *Node) canConnectWith(addr string) bool {
	if n.ListenAddr == addr {
		return false
	}

	for _, connectedAddr := range n.getPeerList() {
		if addr == connectedAddr {
			return false
		}
	}
	return true
}

func (n *Node) getPeerList() []string {
	sli := make([]string, 0, len(n.peers))
	n.mu.RLock()
	defer n.mu.RUnlock()
	for _, v := range n.peers {
		sli = append(sli, v.ListenAddr)
	}
	return sli
}
