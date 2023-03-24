package keeper

import (
	"fmt"
	"math/big"

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

func (k Keeper) MaxBucketsPerAccount(ctx sdk.Context) (res uint32) {
	k.paramStore.Get(ctx, types.KeyMaxBucketsPerAccount, &res)
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

func (k Keeper) MirrorBucketRelayerFee(ctx sdk.Context) *big.Int {
	var relayerFeeParam string

	k.paramStore.Get(ctx, types.KeyMirrorBucketRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorBucketAckRelayerFee(ctx sdk.Context) *big.Int {
	var relayerFeeParam string

	k.paramStore.Get(ctx, types.KeyMirrorBucketAckRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorObjectRelayerFee(ctx sdk.Context) *big.Int {
	var relayerFeeParam string

	k.paramStore.Get(ctx, types.KeyMirrorObjectRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorObjectAckRelayerFee(ctx sdk.Context) *big.Int {
	var relayerFeeParam string

	k.paramStore.Get(ctx, types.KeyMirrorObjectAckRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorGroupRelayerFee(ctx sdk.Context) *big.Int {
	var relayerFeeParam string

	k.paramStore.Get(ctx, types.KeyMirrorGroupRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorGroupAckRelayerFee(ctx sdk.Context) *big.Int {
	var relayerFeeParam string

	k.paramStore.Get(ctx, types.KeyMirrorGroupAckRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.MaxSegmentSize(ctx),
		k.RedundantDataChunkNum(ctx),
		k.RedundantParityChunkNum(ctx),
		k.MaxPayloadSize(ctx),
		k.MaxBucketsPerAccount(ctx),
		k.MinChargeSize(ctx),
		k.MirrorBucketRelayerFee(ctx).String(),
		k.MirrorBucketAckRelayerFee(ctx).String(),
		k.MirrorObjectRelayerFee(ctx).String(),
		k.MirrorObjectAckRelayerFee(ctx).String(),
		k.MirrorGroupRelayerFee(ctx).String(),
		k.MirrorGroupAckRelayerFee(ctx).String(),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}
