package keeper

import (
	"github.com/bnb-chain/greenfield/x/greenfield/types"
)

var _ types.QueryServer = Keeper{}
