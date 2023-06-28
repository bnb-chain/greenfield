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
func (k Keeper) SetSpStoragePrice(ctx sdk.Context, spStoragePrice types.SpStoragePrice) {
	event := &types.EventSpStoragePriceUpdate{
		SpId:          spStoragePrice.SpId,
		UpdateTimeSec: spStoragePrice.UpdateTimeSec,
		ReadPrice:     spStoragePrice.ReadPrice,
		StorePrice:    spStoragePrice.StorePrice,
		FreeReadQuota: spStoragePrice.FreeReadQuota,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)
	key := types.SpStoragePriceKey(
		spStoragePrice.SpId,
		spStoragePrice.UpdateTimeSec,
	)
	spStoragePrice.UpdateTimeSec = 0
	spStoragePrice.SpId = 0
	b := k.cdc.MustMarshal(&spStoragePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// GetSpStoragePrice returns a SpStoragePrice from its index
func (k Keeper) GetSpStoragePrice(
	ctx sdk.Context,
	spId uint32,
	timestamp int64,
) (val types.SpStoragePrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	b := store.Get(types.SpStoragePriceKey(
		spId,
		timestamp,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	val.SpId = spId
	val.UpdateTimeSec = timestamp
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
		spId, timestamp := types.ParseSpStoragePriceKey(iterator.Key())
		val.SpId = spId
		val.UpdateTimeSec = timestamp
		list = append(list, val)
	}

	return
}

// GetSpStoragePriceByTime find the latest price before the given time
func (k Keeper) GetSpStoragePriceByTime(
	ctx sdk.Context,
	spId uint32,
	timestamp int64,
) (val types.SpStoragePrice, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	startKey := types.SpStoragePriceKey(
		spId,
		timestamp,
	)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return val, fmt.Errorf("no price found")
	}

	resSpId, resTimestamp := types.ParseSpStoragePriceKey(iterator.Key())
	if resSpId != spId {
		return val, fmt.Errorf("no price found")
	}
	k.cdc.MustUnmarshal(iterator.Value(), &val)
	val.SpId = spId
	val.UpdateTimeSec = resTimestamp

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
		price, err := k.GetSpStoragePriceByTime(ctx, sp.Id, current+1)
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
	updateTimeSec := types.ParseSecondarySpStorePriceKey(iterator.Key())
	val.UpdateTimeSec = updateTimeSec
	return val, nil
}
