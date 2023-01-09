package keeper

import (
	"fmt"
	"math/big"

	"github.com/bnb-chain/bfs/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetTransferOutRelayerFee gets the transfer out relayer fee param
func (k Keeper) GetTransferOutRelayerFee(ctx sdk.Context) (relayerFee *big.Int) {
	var relayerFeeParam string
	k.paramstore.Get(ctx, types.KeyParamTransferOutRelayerFee, &relayerFeeParam)
	relayerFee, valid := relayerFee.SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}
	return relayerFee
}
