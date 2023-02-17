package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetSpStoragePrice set a specific SpStoragePrice in the store from its index
func (k Keeper) SetSpStoragePrice(ctx sdk.Context, SpStoragePrice types.SpStoragePrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	b := k.cdc.MustMarshal(&SpStoragePrice)
	store.Set(types.SpStoragePriceKey(
		SpStoragePrice.SpAddress,
		SpStoragePrice.UpdateTime,
	), b)
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
		list = append(list, val)
	}

	return
}
