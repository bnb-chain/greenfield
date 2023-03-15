package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) MaxSegmentSize(ctx sdk.Context) (res uint64) {
	k.paramStore.Get(ctx, types.KeyMaxSegmentSize, &res)
	return
}

func (k Keeper) RedundantDataChunkNum(ctx sdk.Context) (res uint32) {
	k.paramStore.Get(ctx, types.KeyRedundantDataChunkNum, &res)
	return
}

func (k Keeper) RedundantParityChunkNum(ctx sdk.Context) (res uint32) {
	k.paramStore.Get(ctx, types.KeyRedundantParityChunkNum, &res)
	return
}

func (k Keeper) GetExpectSecondarySPNumForECObject(ctx sdk.Context) (res uint32) {
	return k.RedundantDataChunkNum(ctx) + k.RedundantParityChunkNum(ctx)
}

func (k Keeper) MaxPayloadSize(ctx sdk.Context) (res uint64) {
	k.paramStore.Get(ctx, types.KeyMaxPayloadSize, &res)
	return
}

func (k Keeper) MinChargeSize(ctx sdk.Context) (res uint64) {
	k.paramStore.Get(ctx, types.KeyMinChargeSize, &res)
	return
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaxSegmentSize(ctx),
		k.RedundantDataChunkNum(ctx),
		k.RedundantParityChunkNum(ctx),
		k.MaxPayloadSize(ctx),
		k.MinChargeSize(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
