package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/bridge/types"
)

// GetParams returns the current bridge module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the params of bridge module
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}

// GetTransferOutRelayerFee gets the transfer out relayer fee params
func (k Keeper) GetTransferOutRelayerFee(ctx sdk.Context) (sdkmath.Int, sdkmath.Int) {
	params := k.GetParams(ctx)

	return params.TransferOutRelayerFee, params.TransferOutAckRelayerFee
}
