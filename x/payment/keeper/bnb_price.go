package keeper

import (
	sdkmath "cosmossdk.io/math"
	"encoding/binary"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBnbPrice set a specific BnbPrice in the store from its index
func (k Keeper) SetBnbPrice(ctx sdk.Context, BnbPrice types.BnbPrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKeyPrefix))
	b := k.cdc.MustMarshal(&BnbPrice)
	store.Set(types.BnbPriceKey(
		BnbPrice.Time,
	), b)
}

// GetBnbPrice returns a BnbPrice from its index
func (k Keeper) GetBnbPrice(
	ctx sdk.Context,
	time int64,

) (val types.BnbPrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKeyPrefix))

	b := store.Get(types.BnbPriceKey(
		time,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBnbPrice removes a BnbPrice from the store
func (k Keeper) RemoveBnbPrice(
	ctx sdk.Context,
	time int64,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKeyPrefix))
	store.Delete(types.BnbPriceKey(
		time,
	))
}

// GetAllBnbPrice returns all BnbPrice
func (k Keeper) GetAllBnbPrice(ctx sdk.Context) (list []types.BnbPrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.BnbPrice
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		fmt.Printf("key: %x, value: %+v\n", iterator.Key(), val)
		list = append(list, val)
	}

	return
}

func (k Keeper) SubmitBNBPrice(ctx sdk.Context, time int64, price uint64) {
	k.SetBnbPrice(ctx, types.BnbPrice{Time: time, Price: price})
}

// GetBNBPrice return the price of BNB at priceTime
// price = num / precision
// e.g. num = 27740000000, precision = 100000000, price = 27740000000 / 100000000 = 277.4
func (k Keeper) GetBNBPriceByTime(ctx sdk.Context, priceTime int64) (bnbPrice types.BNBPrice, err error) {
	//return sdkmath.NewInt(27740000000), sdkmath.NewInt(100000000)
	bnbPrice = types.BNBPrice{
		Precision: sdkmath.NewInt(100000000),
	}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKeyPrefix))
	timeBytes := make([]byte, 8)
	// ReverseIterator end is not included, so we need to add 1 to the priceTime
	binary.BigEndian.PutUint64(timeBytes, uint64(priceTime+1))
	iterator := store.ReverseIterator(nil, timeBytes)
	defer iterator.Close()
	if !iterator.Valid() {
		return bnbPrice, fmt.Errorf("no price found")
	}

	var val types.BnbPrice
	k.cdc.MustUnmarshal(iterator.Value(), &val)
	bnbPrice.Num = sdkmath.NewIntFromUint64(val.Price)
	return
}

func (k Keeper) GetCurrentBNBPrice(ctx sdk.Context) (bnbPrice types.BNBPrice, err error) {
	return k.GetBNBPriceByTime(ctx, ctx.BlockTime().Unix())
}
