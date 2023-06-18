package keeper

import (
	"fmt"
	"sort"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

// SetSpStoragePrice set a specific SpStoragePrice in the store from its index
func (k Keeper) SetSpStoragePrice(ctx sdk.Context, SpStoragePrice types.SpStoragePrice) {
	event := &types.EventSpStoragePriceUpdate{
		SpAddress:     SpStoragePrice.SpAddress,
		UpdateTimeSec: SpStoragePrice.UpdateTimeSec,
		ReadPrice:     SpStoragePrice.ReadPrice,
		StorePrice:    SpStoragePrice.StorePrice,
		FreeReadQuota: SpStoragePrice.FreeReadQuota,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	key := types.SpStoragePriceKey(
		SpStoragePrice.GetSpAccAddress(),
		SpStoragePrice.UpdateTimeSec,
	)
	SpStoragePrice.UpdateTimeSec = 0
	SpStoragePrice.SpAddress = ""
	b := k.cdc.MustMarshal(&SpStoragePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// GetSpStoragePrice returns a SpStoragePrice from its index
func (k Keeper) GetSpStoragePrice(
	ctx sdk.Context,
	spAddr sdk.AccAddress,
	UpdateTimeSec int64,
) (val types.SpStoragePrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	b := store.Get(types.SpStoragePriceKey(
		spAddr,
		UpdateTimeSec,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	val.SpAddress = spAddr.String()
	val.UpdateTimeSec = UpdateTimeSec
	return val, true
}

// GetAllSpStoragePrice returns all SpStoragePrice
func (k Keeper) GetAllSpStoragePrice(ctx sdk.Context) (list []types.SpStoragePrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.SpStoragePrice
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		spAddr, UpdateTimeSec := types.ParseSpStoragePriceKey(iterator.Key())
		val.SpAddress = spAddr.String()
		val.UpdateTimeSec = UpdateTimeSec
		list = append(list, val)
	}

	return
}

// GetSpStoragePriceByTime find the latest price before the given time
func (k Keeper) GetSpStoragePriceByTime(
	ctx sdk.Context,
	spAddr sdk.AccAddress,
	time int64,
) (val types.SpStoragePrice, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	startKey := types.SpStoragePriceKey(
		spAddr,
		time,
	)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return val, fmt.Errorf("no price found")
	}

	spAddrRes, UpdateTimeSec := types.ParseSpStoragePriceKey(iterator.Key())
	if !spAddrRes.Equals(spAddr) {
		return val, fmt.Errorf("no price found")
	}
	k.cdc.MustUnmarshal(iterator.Value(), &val)
	val.SpAddress = spAddr.String()
	val.UpdateTimeSec = UpdateTimeSec

	return val, nil
}

func (k Keeper) SetSecondarySpStorePrice(ctx sdk.Context, secondarySpStorePrice types.SecondarySpStorePrice) {
	event := &types.EventSecondarySpStorePriceUpdate{
		UpdateTimeSec: secondarySpStorePrice.UpdateTimeSec,
		StorePrice:    secondarySpStorePrice.StorePrice,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SecondarySpStorePriceKeyPrefix)
	key := types.SecondarySpStorePriceKey(
		secondarySpStorePrice.UpdateTimeSec,
	)
	secondarySpStorePrice.UpdateTimeSec = 0
	b := k.cdc.MustMarshal(&secondarySpStorePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// UpdateSecondarySpStorePrice calculate the price of secondary store by the average price of all sp store price
func (k Keeper) UpdateSecondarySpStorePrice(ctx sdk.Context) error {
	sps := k.GetAllStorageProviders(ctx)
	current := ctx.BlockTime().Unix()
	prices := make([]sdk.Dec, 0)
	for _, sp := range sps {
		if sp.Status != types.STATUS_IN_SERVICE {
			continue
		}
		price, err := k.GetSpStoragePriceByTime(ctx, sp.GetOperatorAccAddress(), current+1)
		if err != nil {
			return err
		}
		prices = append(prices, price.StorePrice)
	}
	l := len(prices)
	if l == 0 {
		return nil
	}

	sort.Slice(prices, func(i, j int) bool { return prices[i].LT(prices[j]) })
	var median sdk.Dec
	if l%2 == 0 {
		median = prices[l/2-1].Add(prices[l/2]).QuoInt64(2)
	} else {
		median = prices[l/2]
	}
	price := k.SecondarySpStorePriceRatio(ctx).Mul(median)
	secondarySpStorePrice := types.SecondarySpStorePrice{
		StorePrice:    price,
		UpdateTimeSec: current,
	}
	k.SetSecondarySpStorePrice(ctx, secondarySpStorePrice)
	return nil
}

func (k Keeper) GetSecondarySpStorePriceByTime(ctx sdk.Context, time int64) (val types.SecondarySpStorePrice, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SecondarySpStorePriceKeyPrefix)

	startKey := types.SecondarySpStorePriceKey(
		time,
	)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return val, fmt.Errorf("no price found")
	}

	k.cdc.MustUnmarshal(iterator.Value(), &val)
	_, UpdateTimeSec := types.ParseSpStoragePriceKey(iterator.Key())
	val.UpdateTimeSec = UpdateTimeSec
	return val, nil
}
