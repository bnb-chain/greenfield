package keeper

import (
	"fmt"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetSpStoragePrice set a specific SpStoragePrice in the store from its index
func (k Keeper) SetSpStoragePrice(ctx sdk.Context, SpStoragePrice types.SpStoragePrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	key := types.SpStoragePriceKey(
		SpStoragePrice.SpAddress,
		SpStoragePrice.UpdateTime,
	)
	SpStoragePrice.UpdateTime = 0
	SpStoragePrice.SpAddress = ""
	b := k.cdc.MustMarshal(&SpStoragePrice)
	store.Set(key, b)
}

// GetSpStoragePrice returns a SpStoragePrice from its index
func (k Keeper) GetSpStoragePrice(
	ctx sdk.Context,
	spAddr string,
	updateTime int64,
) (val types.SpStoragePrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	b := store.Get(types.SpStoragePriceKey(
		spAddr,
		updateTime,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	val.SpAddress = spAddr
	val.UpdateTime = updateTime
	return val, true
}

// GetAllSpStoragePrice returns all SpStoragePrice
func (k Keeper) GetAllSpStoragePrice(ctx sdk.Context) (list []types.SpStoragePrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.SpStoragePrice
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		spAddr, updateTime := types.ParseSpStoragePriceKey(iterator.Key())
		val.SpAddress = spAddr
		val.UpdateTime = updateTime
		list = append(list, val)
	}

	return
}

// GetSpStoragePriceByTime find the latest price before the given time
func (k Keeper) GetSpStoragePriceByTime(
	ctx sdk.Context,
	spAddr string,
	time int64,
) (val types.SpStoragePrice, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	startKey := types.SpStoragePriceKey(
		spAddr,
		time+1,
	)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return val, fmt.Errorf("no price found")
	}

	k.cdc.MustUnmarshal(iterator.Value(), &val)
	_, updateTime := types.ParseSpStoragePriceKey(iterator.Key())
	val.SpAddress = spAddr
	val.UpdateTime = updateTime

	return val, nil
}
