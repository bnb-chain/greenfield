package keeper

import (
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
