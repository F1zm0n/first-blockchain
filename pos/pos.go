package pos

import (
	"context"
	"sync"

	"github.com/F1zm0n/blocker/proto"
)

type Pos struct {
	proto.UnimplementedPosConsensusServer
	mu         sync.RWMutex
	stakes     map[string]int64
	validators map[string]*proto.Version
}

func (n *Pos) Stake(ctx context.Context, req *proto.StakeRequest) (*proto.Ack, error) {
	panic("a")
}

func (n *Pos) ChooseValidator(ctx context.Context, req *proto.ValidatorRequest) (*proto.Ack, error) {
	panic("")
}
