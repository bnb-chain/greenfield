package keeper

import (
	"github.com/bnb-chain/greenfield/x/bridge/types"
)

var _ types.QueryServer = Keeper{}
