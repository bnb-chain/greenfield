package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/permission/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaximumStatementsNum(ctx),
		k.MaximumPolicyGroupSize(ctx))
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) MaximumStatementsNum(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxStatementsNum, &res)
	return
}

func (k Keeper) MaximumPolicyGroupSize(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxPolicyGroupSIze, &res)
	return
}
