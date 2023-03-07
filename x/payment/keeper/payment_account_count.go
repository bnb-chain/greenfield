package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetPaymentAccountCount set a specific paymentAccountCount in the store from its index
func (k Keeper) SetPaymentAccountCount(ctx sdk.Context, paymentAccountCount *types.PaymentAccountCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)
	key := types.PaymentAccountCountKey(paymentAccountCount.Owner)
	paymentAccountCount.Owner = ""
	b := k.cdc.MustMarshal(paymentAccountCount)
	store.Set(key, b)
}

// GetPaymentAccountCount returns a paymentAccountCount from its index
func (k Keeper) GetPaymentAccountCount(
	ctx sdk.Context,
	owner string,
) (val *types.PaymentAccountCount, found bool) {
	val = &types.PaymentAccountCount{}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)

	b := store.Get(types.PaymentAccountCountKey(
		owner,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, val)
	val.Owner = owner
	return val, true
}

// RemovePaymentAccountCount removes a paymentAccountCount from the store
func (k Keeper) RemovePaymentAccountCount(
	ctx sdk.Context,
	owner string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)
	store.Delete(types.PaymentAccountCountKey(
		owner,
	))
}

// GetAllPaymentAccountCount returns all paymentAccountCount
func (k Keeper) GetAllPaymentAccountCount(ctx sdk.Context) (list []types.PaymentAccountCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PaymentAccountCount
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		val.Owner = string(iterator.Key())
		list = append(list, val)
	}

	return
}
