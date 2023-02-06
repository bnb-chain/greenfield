package keeper

import (
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.EventCountPerBlock(ctx),
		k.ChallengeExpirePeriod(ctx),
		k.SlashCoolingOffPeriod(ctx),
		k.SlashDenom(ctx),
		k.SlashAmountPerKb(ctx),
		k.SlashAmountMin(ctx),
		k.SlashAmountMax(ctx),
		k.RewardValidatorRatio(ctx),
		k.RewardChallengerRatio(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// EventCountPerBlock returns the EventCountPerBlock param
func (k Keeper) EventCountPerBlock(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyEventCountPerBlock, &res)
	return
}

// ChallengeExpirePeriod returns the ChallengeExpirePeriod param
func (k Keeper) ChallengeExpirePeriod(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyChallengeExpirePeriod, &res)
	return
}

// SlashDenom returns the SlashDenom param
func (k Keeper) SlashDenom(ctx sdk.Context) (res string) {
	k.paramstore.Get(ctx, types.KeySlashDenom, &res)
	return
}

// SlashCoolingOffPeriod returns the SlashCoolingOffPeriod param
func (k Keeper) SlashCoolingOffPeriod(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeySlashCoolingOffPeriod, &res)
	return
}

// SlashAmountPerKb returns the SlashAmountPerKb param
func (k Keeper) SlashAmountPerKb(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeySlashAmountPerKb, &res)
	return
}

// SlashAmountMin returns the SlashAmountMin param
func (k Keeper) SlashAmountMin(ctx sdk.Context) (res math.Int) {
	k.paramstore.Get(ctx, types.KeySlashAmountMin, &res)
	return
}

// SlashAmountMax returns the SlashAmountMax param
func (k Keeper) SlashAmountMax(ctx sdk.Context) (res math.Int) {
	k.paramstore.Get(ctx, types.KeySlashAmountMax, &res)
	return
}

// RewardValidatorRatio returns the RewardValidatorRatio param
func (k Keeper) RewardValidatorRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyRewardValidatorRatio, &res)
	return
}

// RewardChallengerRatio returns the RewardChallengerRatio param
func (k Keeper) RewardChallengerRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyRewardChallengerRatio, &res)
	return
}
