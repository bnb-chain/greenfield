package keeper

import (
	sdkmath "cosmossdk.io/math"
	"encoding/binary"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBnbPricePrice set a specific bnbPricePrice in the store from its index
func (k Keeper) SetBnbPricePrice(ctx sdk.Context, bnbPricePrice types.BnbPricePrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPricePriceKeyPrefix))
	b := k.cdc.MustMarshal(&bnbPricePrice)
	store.Set(types.BnbPricePriceKey(
		bnbPricePrice.Time,
	), b)
}

// GetBnbPricePrice returns a bnbPricePrice from its index
func (k Keeper) GetBnbPricePrice(
	ctx sdk.Context,
	time int64,

) (val types.BnbPricePrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPricePriceKeyPrefix))

	b := store.Get(types.BnbPricePriceKey(
		time,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBnbPricePrice removes a bnbPricePrice from the store
func (k Keeper) RemoveBnbPricePrice(
	ctx sdk.Context,
	time int64,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPricePriceKeyPrefix))
	store.Delete(types.BnbPricePriceKey(
		time,
	))
}

// GetAllBnbPricePrice returns all bnbPricePrice
func (k Keeper) GetAllBnbPricePrice(ctx sdk.Context) (list []types.BnbPricePrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPricePriceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.BnbPricePrice
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		fmt.Printf("key: %x, value: %+v\n", iterator.Key(), val)
		list = append(list, val)
	}

	return
}

func (k Keeper) SubmitBNBPrice(ctx sdk.Context, time int64, price uint64) {
	k.SetBnbPricePrice(ctx, types.BnbPricePrice{Time: time, Price: price})
}

// GetBNBPrice return the price of BNB at priceTime
// price = num / precision
// e.g. num = 27740000000, precision = 100000000, price = 27740000000 / 100000000 = 277.4
func (k Keeper) GetBNBPriceByTime(ctx sdk.Context, priceTime int64) (bnbPrice types.BNBPrice, err error) {
	//return sdkmath.NewInt(27740000000), sdkmath.NewInt(100000000)
	bnbPrice = types.BNBPrice{
		Precision: sdkmath.NewInt(100000000),
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPricePriceKeyPrefix))
	timeBytes := make([]byte, 8)
	// ReverseIterator end is not included, so we need to add 1 to the priceTime
	binary.BigEndian.PutUint64(timeBytes, uint64(priceTime+1))
	iterator := store.ReverseIterator(nil, timeBytes)
	defer iterator.Close()
	if !iterator.Valid() {
		return bnbPrice, fmt.Errorf("no price found")
	}
	var val types.BnbPricePrice
	k.cdc.MustUnmarshal(iterator.Value(), &val)
	bnbPrice.Num = sdkmath.NewIntFromUint64(val.Price)
	return
}

func (k Keeper) GetCurrentBNBPrice(ctx sdk.Context) (bnbPrice types.BNBPrice, err error) {
	return k.GetBNBPriceByTime(ctx, ctx.BlockTime().Unix())
}
