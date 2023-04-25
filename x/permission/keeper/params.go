package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/permission/types"
)

func (k Keeper) MaximumStatementsNum(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.MaximumStatementsNum
}

func (k Keeper) MaximumPolicyGroupSize(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.MaximumGroupNum
}

// GetParams returns the current permission module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the params of permission module
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}
