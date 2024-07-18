package node

import (
	"testing"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/util"
	"github.com/dgraph-io/badger/v4"
	"github.com/stretchr/testify/require"
)

// func Test_NodeServer_HandleTransaction(t *testing.T) {
// 	client, err := grpc.NewClient(
// 		":3000",
// 		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}...)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer client.Close()

// 	newCli := proto.NewNodeClient(client)
// 	privKey := crypto.NewPrivateKey()
// 	for {
// 		time.Sleep(2 * time.Second)
// 		tx := &proto.Transaction{
// 			Version: 0,
// 			Inputs: []*proto.TxInput{
// 				{PrevTxHash: util.RandomHash(), PrevOutIndex: 0, PublicKey: privKey.Bytes()},
// 			},
// 			Outputs: []*proto.TxOutput{
// 				{
// 					Amount:  99,
// 					Address: privKey.Public().Address().Bytes(),
// 				},
// 			},
// 		}
// 		_, err = newCli.HandleTransaction(context.TODO(), tx)
// 		require.NoError(t, err)
// 	}
// }

func TestBadgerMemPoolPut(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/test/mempool")
	opts.Logger = nil
	var (
		pk = crypto.NewPrivateKey()
		r  = require.New(t)
		tx = &proto.Transaction{
			Version: 1,
			Inputs: []*proto.TxInput{{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 1,
				PublicKey:    pk.Public().Bytes(),
			}},
			Outputs: []*proto.TxOutput{
				{
					Amount:  100,
					Address: pk.Public().Address().Bytes(),
				},
			},
		}
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})
	pool := NewBadgerMemPool(db)
	r.NoError(pool.Put(tx))
}

func TestBadgerMemPoolHas(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/test/mempool")
	opts.Logger = nil
	var (
		pk = crypto.NewPrivateKey()
		r  = require.New(t)
		tx = &proto.Transaction{
			Version: 1,
			Inputs: []*proto.TxInput{{
				PrevTxHash:   util.RandomHash(),
				PrevOutIndex: 1,
				PublicKey:    pk.Public().Bytes(),
			}},
			Outputs: []*proto.TxOutput{
				{
					Amount:  100,
					Address: pk.Public().Address().Bytes(),
				},
			},
		}
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})
	pool := NewBadgerMemPool(db)
	r.NoError(pool.Put(tx))

	r.NoError(pool.Has(tx))
}

func TestBadgerMemPoolClear(t *testing.T) {
	opts := badger.DefaultOptions("/tmp/test/mempool")
	opts.Logger = nil
	var (
		pk  = crypto.NewPrivateKey()
		r   = require.New(t)
		txs = []*proto.Transaction{
			{
				Version: 1,
				Inputs: []*proto.TxInput{{
					PrevTxHash:   util.RandomHash(),
					PrevOutIndex: 1,
					PublicKey:    pk.Public().Bytes(),
				}},
				Outputs: []*proto.TxOutput{
					{
						Amount:  100,
						Address: pk.Public().Address().Bytes(),
					},
				},
			},
			{
				Version: 2,
				Inputs: []*proto.TxInput{{
					PrevTxHash:   util.RandomHash(),
					PrevOutIndex: 2,
					PublicKey:    pk.Public().Bytes(),
				}},
				Outputs: []*proto.TxOutput{
					{
						Amount:  300,
						Address: pk.Public().Address().Bytes(),
					},
				},
			},
			{
				Version: 3,
				Inputs: []*proto.TxInput{{
					PrevTxHash:   util.RandomHash(),
					PrevOutIndex: 3,
					PublicKey:    pk.Public().Bytes(),
				}},
				Outputs: []*proto.TxOutput{
					{
						Amount:  400,
						Address: pk.Public().Address().Bytes(),
					},
				},
			},
		}
		db, err = badger.Open(opts)
	)
	r.NoError(err)
	t.Cleanup(func() {
		db.DropAll()
		db.Close()
	})
	pool := NewBadgerMemPool(db)
	for i, tx := range txs {
		r.NoError(pool.Put(tx))
		r.Equal(pool.Len(), i+1)
		r.NoError(pool.Has(tx))
	}

	fetchedTxs, err := pool.Clear()
	r.NoError(err)

	r.Len(fetchedTxs, len(txs))
}
