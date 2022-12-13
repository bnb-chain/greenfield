package keeper

import (
	"github.com/bnb-chain/bfs/x/bridge/types"
)

var _ types.QueryServer = Keeper{}
