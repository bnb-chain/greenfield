package keeper

import (
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

var _ types.QueryServer = Keeper{}
