package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) MaxSegmentSize(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxSegmentSize, &res)
	return
}

func (k Keeper) RedundantDataChunkNum(ctx sdk.Context) (res uint32) {
	k.paramstore.Get(ctx, types.KeyRedundantDataChunkNum, &res)
	return
}

func (k Keeper) RedundantParityChunkNum(ctx sdk.Context) (res uint32) {
	k.paramstore.Get(ctx, types.KeyRedundantParityChunkNum, &res)
	return
}

func (k Keeper) MaxPayloadSize(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyMaxPayloadSize, &res)
	return
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaxSegmentSize(ctx),
		k.RedundantDataChunkNum(ctx),
		k.RedundantParityChunkNum(ctx),
		k.MaxPayloadSize(ctx))
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}
