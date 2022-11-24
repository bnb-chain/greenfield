package keeper

import (
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPaymentAccountCount set a specific paymentAccountCount in the store from its index
func (k Keeper) SetPaymentAccountCount(ctx sdk.Context, paymentAccountCount types.PaymentAccountCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountCountKeyPrefix))
	b := k.cdc.MustMarshal(&paymentAccountCount)
	store.Set(types.PaymentAccountCountKey(
		paymentAccountCount.Owner,
	), b)
}

// GetPaymentAccountCount returns a paymentAccountCount from its index
func (k Keeper) GetPaymentAccountCount(
	ctx sdk.Context,
	owner string,

) (val types.PaymentAccountCount, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountCountKeyPrefix))

	b := store.Get(types.PaymentAccountCountKey(
		owner,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePaymentAccountCount removes a paymentAccountCount from the store
func (k Keeper) RemovePaymentAccountCount(
	ctx sdk.Context,
	owner string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountCountKeyPrefix))
	store.Delete(types.PaymentAccountCountKey(
		owner,
	))
}

// GetAllPaymentAccountCount returns all paymentAccountCount
func (k Keeper) GetAllPaymentAccountCount(ctx sdk.Context) (list []types.PaymentAccountCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountCountKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PaymentAccountCount
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
