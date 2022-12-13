package keeper

import (
	"github.com/bnb-chain/inscription/x/storage/types"
)

var _ types.QueryServer = Keeper{}
