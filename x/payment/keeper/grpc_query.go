package keeper

import (
	"github.com/bnb-chain/greenfield/x/payment/types"
)

var _ types.QueryServer = Keeper{}
