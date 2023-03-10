package keeper

import (
	"github.com/bnb-chain/greenfield/x/permission/types"
)

var _ types.QueryServer = Keeper{}
