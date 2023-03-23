package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/bridge/types"
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

// GetTransferOutRelayerFee gets the transfer out relayer fee params
func (k Keeper) GetTransferOutRelayerFee(ctx sdk.Context) (sdkmath.Int, sdkmath.Int) {
	var relayerFeeParam, ackRelayerFeeParam sdkmath.Int

	k.paramstore.Get(ctx, types.KeyParamTransferOutRelayerFee, &relayerFeeParam)
	k.paramstore.Get(ctx, types.KeyParamTransferOutAckRelayerFee, &ackRelayerFeeParam)

	return relayerFeeParam, ackRelayerFeeParam
}
