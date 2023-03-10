package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.ChallengeCountPerBlock(ctx),
		k.SlashCoolingOffPeriod(ctx),
		k.SlashAmountSizeRate(ctx),
		k.SlashAmountMin(ctx),
		k.SlashAmountMax(ctx),
		k.RewardValidatorRatio(ctx),
		k.RewardSubmitterRatio(ctx),
		k.RewardSubmitterThreshold(ctx),
		k.HeartbeatInterval(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// ChallengeCountPerBlock returns the ChallengeCountPerBlock param
func (k Keeper) ChallengeCountPerBlock(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyChallengeCountPerBlock, &res)
	return
}

// SlashCoolingOffPeriod returns the SlashCoolingOffPeriod param
func (k Keeper) SlashCoolingOffPeriod(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeySlashCoolingOffPeriod, &res)
	return
}

// SlashAmountSizeRate returns the SlashAmountSizeRate param
func (k Keeper) SlashAmountSizeRate(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeySlashAmountSizeRate, &res)
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

// RewardSubmitterRatio returns the RewardSubmitterRatio param
func (k Keeper) RewardSubmitterRatio(ctx sdk.Context) (res sdk.Dec) {
	k.paramstore.Get(ctx, types.KeyRewardSubmitterRatio, &res)
	return
}

// HeartbeatInterval returns the HeartbeatInterval param
func (k Keeper) HeartbeatInterval(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyHeartbeatInterval, &res)
	return
}

// RewardSubmitterThreshold returns the RewardSubmitterThreshold param
func (k Keeper) RewardSubmitterThreshold(ctx sdk.Context) (res math.Int) {
	k.paramstore.Get(ctx, types.KeyRewardSubmitterThreshold, &res)
	return
}
