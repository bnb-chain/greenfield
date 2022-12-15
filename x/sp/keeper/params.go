package keeper

import (
	"cosmossdk.io/math"
	"github.com/bnb-chain/bfs/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k Keeper) MaxStorageProviders(ctx sdk.Context) (res uint32) {
	k.paramstore.Get(ctx, types.KeyMaxStorageProviders, &res)
	return
}

func (k Keeper) DepositDenomForSP(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyDepostDenom, &res)
	return
}

func (k Keeper) MinDeposit(ctx sdk.Context) (res math.Int) {
	k.paramstore.Get(ctx, types.KeyMinDeposit, &res)
	return
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaxStorageProviders(ctx),
		k.DepositDenomForSP(ctx),
		k.MinDeposit(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
