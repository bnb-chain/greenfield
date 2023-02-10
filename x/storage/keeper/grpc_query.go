package keeper

import (
	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ types.QueryServer = Keeper{}
