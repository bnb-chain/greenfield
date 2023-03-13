package keeper

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

// SetSpStoragePrice set a specific SpStoragePrice in the store from its index
func (k Keeper) SetSpStoragePrice(ctx sdk.Context, SpStoragePrice types.SpStoragePrice) {
	event := &types.EventSpStoragePriceUpdate{
		SpAddress:     SpStoragePrice.SpAddress,
		UpdateTime:    SpStoragePrice.UpdateTime,
		ReadPrice:     SpStoragePrice.ReadPrice,
		StorePrice:    SpStoragePrice.StorePrice,
		FreeReadQuota: SpStoragePrice.FreeReadQuota,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	key := types.SpStoragePriceKey(
		SpStoragePrice.SpAddress,
		SpStoragePrice.UpdateTime,
	)
	SpStoragePrice.UpdateTime = 0
	SpStoragePrice.SpAddress = ""
	b := k.cdc.MustMarshal(&SpStoragePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
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

func (k Keeper) SetSecondarySpStorePrice(ctx sdk.Context, secondarySpStorePrice types.SecondarySpStorePrice) {
	event := &types.EventSecondarySpStorePriceUpdate{
		UpdateTime: secondarySpStorePrice.UpdateTime,
		StorePrice: secondarySpStorePrice.StorePrice,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SecondarySpStorePriceKeyPrefix)
	key := types.SecondarySpStorePriceKey(
		secondarySpStorePrice.UpdateTime,
	)
	secondarySpStorePrice.UpdateTime = 0
	b := k.cdc.MustMarshal(&secondarySpStorePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// UpdateSecondarySpStorePrice calculate the price of secondary store by the average price of all sp store price
func (k Keeper) UpdateSecondarySpStorePrice(ctx sdk.Context) error {
	sps := k.GetAllStorageProviders(ctx)
	total := sdk.ZeroDec()
	current := ctx.BlockTime().Unix()
	var spNumInService int64
	for _, sp := range sps {
		if sp.Status != types.STATUS_IN_SERVICE {
			continue
		}
		price, err := k.GetSpStoragePriceByTime(ctx, sp.OperatorAddress, current)
		if err != nil {
			return err
		}
		spNumInService++
		total = total.Add(price.StorePrice)
	}
	if spNumInService == 0 {
		return nil
	}
	price := types.SecondarySpStorePriceRatio.Mul(total).QuoInt64(spNumInService)
	secondarySpStorePrice := types.SecondarySpStorePrice{
		StorePrice: price,
		UpdateTime: current,
	}
	k.SetSecondarySpStorePrice(ctx, secondarySpStorePrice)
	return nil
}

func (k Keeper) GetSecondarySpStorePriceByTime(ctx sdk.Context, time int64) (val types.SecondarySpStorePrice, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SecondarySpStorePriceKeyPrefix)

	startKey := types.SecondarySpStorePriceKey(
		time + 1,
	)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return val, fmt.Errorf("no price found")
	}

	k.cdc.MustUnmarshal(iterator.Value(), &val)
	_, updateTime := types.ParseSpStoragePriceKey(iterator.Key())
	val.UpdateTime = updateTime
	return val, nil
}
