package keeper

import (
	"github.com/bnb-chain/bfs/x/bfs/types"
)

var _ types.QueryServer = Keeper{}
