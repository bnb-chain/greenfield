package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) DepositDenomForSP(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeyDepostDenom, &res)
	return
}

func (k Keeper) MinDeposit(ctx sdk.Context) (res math.Int) {
	k.paramstore.Get(ctx, types.KeyMinDeposit, &res)
	return
}

func (k Keeper) SecondarySpStorePriceRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeySecondarySpStorePriceRatio, &res)
	return
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.DepositDenomForSP(ctx),
		k.MinDeposit(ctx),
		k.SecondarySpStorePriceRatio(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
