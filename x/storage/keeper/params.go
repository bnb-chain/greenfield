package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) MaxSegmentSize(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.MaxSegmentSize
}

func (k Keeper) RedundantDataChunkNum(ctx sdk.Context) (res uint32) {
	params := k.GetParams(ctx)
	return params.RedundantDataChunkNum
}

func (k Keeper) RedundantParityChunkNum(ctx sdk.Context) (res uint32) {
	params := k.GetParams(ctx)
	return params.RedundantParityChunkNum
}

func (k Keeper) MaxBucketsPerAccount(ctx sdk.Context) (res uint32) {
	params := k.GetParams(ctx)
	return params.MaxBucketsPerAccount
}

func (k Keeper) GetExpectSecondarySPNumForECObject(ctx sdk.Context) (res uint32) {
	return k.RedundantDataChunkNum(ctx) + k.RedundantParityChunkNum(ctx)
}

func (k Keeper) MaxPayloadSize(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.MaxPayloadSize
}

func (k Keeper) MinChargeSize(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.MinChargeSize
}

func (k Keeper) MirrorBucketRelayerFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	relayerFeeParam := params.MirrorBucketRelayerFee
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorBucketAckRelayerFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	relayerFeeParam := params.MirrorBucketAckRelayerFee

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorObjectRelayerFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	relayerFeeParam := params.MirrorObjectRelayerFee

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorObjectAckRelayerFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	relayerFeeParam := params.MirrorObjectAckRelayerFee

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorGroupRelayerFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	relayerFeeParam := params.MirrorGroupRelayerFee

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorGroupAckRelayerFee(ctx sdk.Context) *big.Int {
	params := k.GetParams(ctx)
	relayerFeeParam := params.MirrorGroupAckRelayerFee

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) DiscontinueCountingWindow(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.DiscontinueCountingWindow
}

func (k Keeper) DiscontinueObjectMax(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.DiscontinueObjectMax
}

func (k Keeper) DiscontinueBucketMax(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.DiscontinueBucketMax
}

func (k Keeper) DiscontinueConfirmPeriod(ctx sdk.Context) (res int64) {
	params := k.GetParams(ctx)
	return params.DiscontinueConfirmPeriod
}

func (k Keeper) DiscontinueDeletionMax(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.DiscontinueDeletionMax
}

// GetParams returns the current storage module parameters.
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}

	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams sets the params of storage module
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	return nil
}
