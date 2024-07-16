package types

import (
	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/proto"
	"github.com/F1zm0n/blocker/util"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestHashTransaction(t *testing.T) {
	fromPrivateKey := crypto.NewPrivateKey()
	fromAddress := fromPrivateKey.Public().Address().Bytes()

	toPrivateKey := crypto.NewPrivateKey()
	toAddress := toPrivateKey.Public().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPrivateKey.Public().Bytes(),
	}
	outputSend := &proto.TxOutput{
		Amount:  5,
		Address: toAddress,
	}
	outputBack := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress,
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{outputSend, outputBack},
	}
	sig := SignTransaction(fromPrivateKey, tx)
	input.Signature = sig.Bytes()

	require.True(t, VerifyTransaction(tx))

}
