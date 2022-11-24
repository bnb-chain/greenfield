package keeper

import (
	"github.com/bnb-chain/bfs/x/payment/types"
)

var _ types.QueryServer = Keeper{}
