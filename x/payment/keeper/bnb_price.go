package keeper

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetBnbPrice set bnbPrice in the store
func (k Keeper) SetBnbPrice(ctx sdk.Context, bnbPrice types.BnbPrice) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKey))
	b := k.cdc.MustMarshal(&bnbPrice)
	store.Set([]byte{0}, b)
}

// GetBnbPrice returns bnbPrice
func (k Keeper) GetBnbPrice(ctx sdk.Context) (val types.BnbPrice, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKey))

	b := store.Get([]byte{0})
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveBnbPrice removes bnbPrice from the store
func (k Keeper) RemoveBnbPrice(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.BnbPriceKey))
	store.Delete([]byte{0})
}

func (k Keeper) SubmitBNBPrice(ctx sdk.Context, time int64, price uint64) {
	bnbPrice, _ := k.GetBnbPrice(ctx)
	bnbPrice.Prices = append(bnbPrice.Prices, &types.SingleBnbPrice{
		Time:  time,
		Price: price,
	})
	k.SetBnbPrice(ctx, bnbPrice)
}

// GetBNBPrice return the price of BNB at priceTime
// price = num / precision
// e.g. num = 27740000000, precision = 100000000, price = 27740000000 / 100000000 = 277.4
func (k Keeper) GetBNBPriceByTime(ctx sdk.Context, priceTime int64) (num, precision sdkmath.Int, err error) {
	//return sdkmath.NewInt(27740000000), sdkmath.NewInt(100000000)
	prices, _ := k.GetBnbPrice(ctx)
	length := len(prices.Prices)
	if length == 0 {
		err = fmt.Errorf("no bnb price found")
		return
	}
	precision = sdkmath.NewInt(100000000)
	for i := length - 1; i >= 0; i-- {
		if prices.Prices[i].Time <= priceTime {
			num = sdkmath.NewIntFromUint64(prices.Prices[i].Price)
			return
		}
	}
	num = sdkmath.NewIntFromUint64(prices.Prices[0].Price)
	return
}
