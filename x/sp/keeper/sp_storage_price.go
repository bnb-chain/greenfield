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
	key := types.SpStoragePriceKey(spStoragePrice.SpId)
	spStoragePrice.SpId = 0
	b := k.cdc.MustMarshal(&spStoragePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// GetSpStoragePrice returns a SpStoragePrice from its index
func (k Keeper) GetSpStoragePrice(
	ctx sdk.Context,
	spId uint32,
) (val types.SpStoragePrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SpStoragePriceKeyPrefix)

	b := store.Get(types.SpStoragePriceKey(spId))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	val.SpId = spId
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
		spId := types.ParseSpStoragePriceKey(iterator.Key())
		val.SpId = spId
		list = append(list, val)
	}

	return
}

func (k Keeper) SetGlobalSpStorePrice(ctx sdk.Context, globalSpStorePrice types.GlobalSpStorePrice) {
	event := &types.EventGlobalSpStorePriceUpdate{
		UpdateTimeSec:       globalSpStorePrice.UpdateTimeSec,
		PrimaryStorePrice:   globalSpStorePrice.PrimaryStorePrice,
		SecondaryStorePrice: globalSpStorePrice.SecondaryStorePrice,
		ReadPrice:           globalSpStorePrice.ReadPrice,
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GlobalSpStorePriceKeyPrefix)
	key := types.GlobalSpStorePriceKey(
		globalSpStorePrice.UpdateTimeSec,
	)
	globalSpStorePrice.UpdateTimeSec = 0
	b := k.cdc.MustMarshal(&globalSpStorePrice)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(event)
}

// UpdateGlobalSpStorePrice calculate the global prices by the median price of all sp store price
func (k Keeper) UpdateGlobalSpStorePrice(ctx sdk.Context) error {
	sps := k.GetAllStorageProviders(ctx)
	current := ctx.BlockTime().Unix()
	storePrices := make([]sdk.Dec, 0)
	readPrices := make([]sdk.Dec, 0)
	for _, sp := range sps {
		if sp.Status == types.STATUS_IN_SERVICE || sp.Status == types.STATUS_IN_MAINTENANCE {
			price, found := k.GetSpStoragePrice(ctx, sp.Id)
			if !found {
				return fmt.Errorf("cannot find price for storage provider %d", sp.Id)
			}
			storePrices = append(storePrices, price.StorePrice)
			readPrices = append(readPrices, price.ReadPrice)
		}
	}
	l := len(storePrices)
	if l == 0 {
		return nil
	}

	primaryStorePrice := k.calculateMedian(storePrices)
	secondaryStorePrice := k.SecondarySpStorePriceRatio(ctx).Mul(primaryStorePrice)
	readPrice := k.calculateMedian(readPrices)

	globalSpStorePrice := types.GlobalSpStorePrice{
		PrimaryStorePrice:   primaryStorePrice,
		SecondaryStorePrice: secondaryStorePrice,
		ReadPrice:           readPrice,
		UpdateTimeSec:       current,
	}
	k.SetGlobalSpStorePrice(ctx, globalSpStorePrice)
	return nil
}

func (k Keeper) calculateMedian(prices []sdk.Dec) sdk.Dec {
	l := len(prices)
	sort.Slice(prices, func(i, j int) bool { return prices[i].LT(prices[j]) })
	var median sdk.Dec
	if l%2 == 0 {
		median = prices[l/2-1].Add(prices[l/2]).QuoInt64(2)
	} else {
		median = prices[l/2]
	}
	return median
}

func (k Keeper) GetGlobalSpStorePriceByTime(ctx sdk.Context, time int64) (val types.GlobalSpStorePrice, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.GlobalSpStorePriceKeyPrefix)

	startKey := types.GlobalSpStorePriceKey(
		time,
	)
	iterator := store.ReverseIterator(nil, startKey)
	defer iterator.Close()
	if !iterator.Valid() {
		return val, fmt.Errorf("no price found")
	}

	k.cdc.MustUnmarshal(iterator.Value(), &val)
	updateTimeSec := types.ParseGlobalSpStorePriceKey(iterator.Key())
	val.UpdateTimeSec = updateTimeSec
	return val, nil
}
