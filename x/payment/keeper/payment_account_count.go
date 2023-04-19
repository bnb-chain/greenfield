package keeper

import (
	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetPaymentAccountCount set a specific paymentAccountCount in the store from its index
func (k Keeper) SetPaymentAccountCount(ctx sdk.Context, paymentAccountCount *types.PaymentAccountCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)
	key := types.PaymentAccountCountKey(sdk.MustAccAddressFromHex(paymentAccountCount.Owner))
	paymentAccountCount.Owner = ""
	b := k.cdc.MustMarshal(paymentAccountCount)
	store.Set(key, b)
}

// GetPaymentAccountCount returns a paymentAccountCount from its index
func (k Keeper) GetPaymentAccountCount(
	ctx sdk.Context,
	owner sdk.AccAddress,
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
	val.Owner = owner.String()
	return val, true
}

// RemovePaymentAccountCount removes a paymentAccountCount from the store
func (k Keeper) RemovePaymentAccountCount(
	ctx sdk.Context,
	owner sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)
	store.Delete(types.PaymentAccountCountKey(
		owner,
	))
}

// GetAllPaymentAccountCount returns all paymentAccountCount
func (k Keeper) GetAllPaymentAccountCount(ctx sdk.Context) (list []types.PaymentAccountCount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountCountKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PaymentAccountCount
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		val.Owner = sdk.AccAddress(iterator.Key()).String()
		list = append(list, val)
	}

	return
}
