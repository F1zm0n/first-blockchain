package node

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/util"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Test_NodeServer_HandleTransaction(t *testing.T) {
	client, err := grpc.NewClient(
		":3000",
		[]grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}...)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	newCli := proto.NewNodeClient(client)
	privKey := crypto.NewPrivateKey()
	for {
		time.Sleep(2 * time.Second)
		tx := &proto.Transaction{
			Version: 0,
			Inputs: []*proto.TxInput{
				{PrevTxHash: util.RandomHash(), PrevOutIndex: 0, PublicKey: privKey.Bytes()},
			},
			Outputs: []*proto.TxOutput{
				{
					Amount:  99,
					Address: privKey.Public().Address().Bytes(),
				},
			},
		}
		_, err = newCli.HandleTransaction(context.TODO(), tx)
		require.NoError(t, err)
	}
}
