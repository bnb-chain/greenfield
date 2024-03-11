package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (p types.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.ParamsKey)
	if bz == nil {
		return p
	}
	k.cdc.MustUnmarshal(bz, &p)
	return p
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&params)
	store.Set(types.ParamsKey, bz)

	// store versioned params
	err := k.SetVersionedParamsWithTs(ctx, params.VersionedParams)
	if err != nil {
		return err
	}

	return nil
}

func (k Keeper) SetVersionedParamsWithTs(ctx sdk.Context, verParams types.VersionedParams) error {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VersionedParamsKeyPrefix)
	key := types.VersionedParamsKey(ctx.BlockTime().Unix())

	b := k.cdc.MustMarshal(&verParams)
	store.Set(key, b)

	return nil
}

// GetVersionedParamsWithTs find the latest params before and equal than the specific timestamp
func (k Keeper) GetVersionedParamsWithTs(ctx sdk.Context, ts int64) (verParams types.VersionedParams, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.VersionedParamsKeyPrefix)

	// params are updated in the endblock, so we do not need to make the ts to be included
	// for example, if the params is updated in 100 timestamp: the txs that are executed in 100 timestamp
	// will use the old parameter, after 100 timestamp, when we passing 100 to query, we should still get
	// the old parameter.
	startKey := types.VersionedParamsKey(ts)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return verParams, fmt.Errorf("no versioned params found")
	}

	k.cdc.MustUnmarshal(iterator.Value(), &verParams)

	return verParams, nil
}
