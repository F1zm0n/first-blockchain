package main

import (
	"encoding/hex"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/F1zm0n/blocker/crypto"
	"github.com/F1zm0n/blocker/node"
	"github.com/F1zm0n/blocker/util"
)

func main() {
	log.Println(hex.EncodeToString(util.RandomHash()))
	sl := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource:   false,
		Level:       slog.LevelDebug,
		ReplaceAttr: nil,
	}))

	pk := crypto.NewPrivateKey()

	nodeServ := node.New(sl, node.NewMemoryMemPool(), node.ServerConfig{
		Version:    "blocker-0.1",
		ListenAddr: ":3000",
		PrivateKey: &pk,
	})

	nodeCli := node.New(sl, node.NewMemoryMemPool(), node.ServerConfig{
		Version:    "blocker-0.1",
		ListenAddr: ":4000",
	})

	nodeBack := node.New(sl, node.NewMemoryMemPool(), node.ServerConfig{
		Version:    "blocker-0.1",
		ListenAddr: ":5000",
	})

	makeNode(&nodeServ, []string{})
	time.Sleep(1 * time.Second)
	makeNode(&nodeCli, []string{":3000"})
	time.Sleep(1 * time.Second)
	makeNode(&nodeBack, []string{":4000"})

	block := make(chan struct{})
	<-block
}

func makeNode(n *node.Node, bootstrapNodes []string) *node.Node {
	go func() {
		log.Fatal(n.Start(bootstrapNodes))
	}()
	return n
}
