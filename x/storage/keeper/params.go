package keeper

import (
	"fmt"
	"math/big"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) MaxBucketsPerAccount(ctx sdk.Context) (res uint32) {
	params := k.GetParams(ctx)
	return params.MaxBucketsPerAccount
}

func (k Keeper) GetExpectSecondarySPNumForECObject(ctx sdk.Context, createTime int64) (res uint32) {
	versionParams, err := k.GetVersionedParamsWithTs(ctx, createTime)
	if err != nil {
		panic(fmt.Sprintf("get expect secondary sp num error, msg: %s", err))
	}
	return versionParams.RedundantParityChunkNum + versionParams.RedundantDataChunkNum
}

func (k Keeper) MaxPayloadSize(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.MaxPayloadSize
}

func (k Keeper) MirrorBucketRelayerFee(ctx sdk.Context, destChainId sdk.ChainID) *big.Int {
	params := k.GetParams(ctx)

	var relayerFeeParam string
	if k.crossChainKeeper.GetDestBscChainID() == destChainId {
		relayerFeeParam = params.BscMirrorBucketRelayerFee
	} else if k.crossChainKeeper.GetDestOpChainID() == destChainId {
		relayerFeeParam = params.OpMirrorBucketRelayerFee
	} else {
		panic(fmt.Sprintf("chain id(%d) is not supported", destChainId))
	}

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorBucketAckRelayerFee(ctx sdk.Context, destChainId sdk.ChainID) *big.Int {
	params := k.GetParams(ctx)

	var relayerFeeParam string
	if k.crossChainKeeper.GetDestBscChainID() == destChainId {
		relayerFeeParam = params.BscMirrorBucketAckRelayerFee
	} else if k.crossChainKeeper.GetDestOpChainID() == destChainId {
		relayerFeeParam = params.OpMirrorBucketAckRelayerFee
	} else {
		panic(fmt.Sprintf("chain id(%d) is not supported", destChainId))
	}

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorObjectRelayerFee(ctx sdk.Context, destChainId sdk.ChainID) *big.Int {
	params := k.GetParams(ctx)

	var relayerFeeParam string
	if k.crossChainKeeper.GetDestBscChainID() == destChainId {
		relayerFeeParam = params.BscMirrorObjectRelayerFee
	} else if k.crossChainKeeper.GetDestOpChainID() == destChainId {
		relayerFeeParam = params.OpMirrorObjectRelayerFee
	} else {
		panic(fmt.Sprintf("chain id(%d) is not supported", destChainId))
	}

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorObjectAckRelayerFee(ctx sdk.Context, destChainId sdk.ChainID) *big.Int {
	params := k.GetParams(ctx)

	var relayerFeeParam string
	if k.crossChainKeeper.GetDestBscChainID() == destChainId {
		relayerFeeParam = params.BscMirrorObjectAckRelayerFee
	} else if k.crossChainKeeper.GetDestOpChainID() == destChainId {
		relayerFeeParam = params.OpMirrorObjectAckRelayerFee
	} else {
		panic(fmt.Sprintf("chain id(%d) is not supported", destChainId))
	}

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorGroupRelayerFee(ctx sdk.Context, destChainId sdk.ChainID) *big.Int {
	params := k.GetParams(ctx)
	var relayerFeeParam string
	if k.crossChainKeeper.GetDestBscChainID() == destChainId {
		relayerFeeParam = params.BscMirrorGroupRelayerFee
	} else if k.crossChainKeeper.GetDestOpChainID() == destChainId {
		relayerFeeParam = params.OpMirrorGroupRelayerFee
	} else {
		panic(fmt.Sprintf("chain id(%d) is not supported", destChainId))
	}

	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	return relayerFee
}

func (k Keeper) MirrorGroupAckRelayerFee(ctx sdk.Context, destChainId sdk.ChainID) *big.Int {
	params := k.GetParams(ctx)

	var relayerFeeParam string
	if k.crossChainKeeper.GetDestBscChainID() == destChainId {
		relayerFeeParam = params.BscMirrorGroupAckRelayerFee
	} else if k.crossChainKeeper.GetDestOpChainID() == destChainId {
		relayerFeeParam = params.OpMirrorGroupAckRelayerFee
	} else {
		panic(fmt.Sprintf("chain id(%d) is not supported", destChainId))
	}

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

func (k Keeper) MaxSegmentSize(ctx sdk.Context) (res uint64) {
	p := k.GetParams(ctx)
	params := p.GetVersionedParams()
	return params.MaxSegmentSize
}

func (k Keeper) RedundantDataChunkNum(ctx sdk.Context) (res uint32) {
	p := k.GetParams(ctx)
	params := p.GetVersionedParams()
	return params.RedundantDataChunkNum
}

func (k Keeper) RedundantParityChunkNum(ctx sdk.Context) (res uint32) {
	p := k.GetParams(ctx)
	params := p.GetVersionedParams()
	return params.RedundantParityChunkNum
}

func (k Keeper) MinChargeSize(ctx sdk.Context) (res uint64) {
	p := k.GetParams(ctx)
	params := p.GetVersionedParams()
	return params.MinChargeSize
}

func (k Keeper) StalePolicyCleanupMax(ctx sdk.Context) (res uint64) {
	params := k.GetParams(ctx)
	return params.StalePolicyCleanupMax
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

	// store another kv with timestamp
	err := k.SetVersionedParamsWithTs(ctx, params.VersionedParams)
	if err != nil {
		return err
	}

	return nil
}

// SetVersionedParamsWithTs set a specific params in the store from its index
func (k Keeper) SetVersionedParamsWithTs(ctx sdk.Context, verParams types.VersionedParams) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VersionedParamsKeyPrefix)
	key := types.GetParamsKeyWithTimestamp(ctx.BlockTime().Unix())

	b := k.cdc.MustMarshal(&verParams)
	store.Set(key, b)

	return nil
}

// GetVersionedParamsWithTs find the latest params before and equal than the specific timestamp
func (k Keeper) GetVersionedParamsWithTs(ctx sdk.Context, ts int64) (verParams types.VersionedParams, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VersionedParamsKeyPrefix)

	// ReverseIterator will exclusive end, so we increment ts by 1
	startKey := types.GetParamsKeyWithTimestamp(ts + 1)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return verParams, fmt.Errorf("no versioned params found, ts:%d", uint64(ts))
	}

	k.cdc.MustUnmarshal(iterator.Value(), &verParams)

	return verParams, nil
}

func (k Keeper) MaxLocalVirtualGroupNumPerBucket(ctx sdk.Context) (res uint32) {
	params := k.GetParams(ctx)
	return params.MaxLocalVirtualGroupNumPerBucket
}
