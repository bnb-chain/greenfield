package keeper

import (
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetPaymentAccount set a specific paymentAccount in the store from its index
func (k Keeper) SetPaymentAccount(ctx sdk.Context, paymentAccount types.PaymentAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountKeyPrefix))
	b := k.cdc.MustMarshal(&paymentAccount)
	store.Set(types.PaymentAccountKey(
		paymentAccount.Addr,
	), b)
}

// GetPaymentAccount returns a paymentAccount from its index
func (k Keeper) GetPaymentAccount(
	ctx sdk.Context,
	addr string,

) (val types.PaymentAccount, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountKeyPrefix))

	b := store.Get(types.PaymentAccountKey(
		addr,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemovePaymentAccount removes a paymentAccount from the store
func (k Keeper) RemovePaymentAccount(
	ctx sdk.Context,
	addr string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountKeyPrefix))
	store.Delete(types.PaymentAccountKey(
		addr,
	))
}

// GetAllPaymentAccount returns all paymentAccount
func (k Keeper) GetAllPaymentAccount(ctx sdk.Context) (list []types.PaymentAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.PaymentAccountKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PaymentAccount
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) IsPaymentAccountOwner(ctx sdk.Context, addr string, owner string) bool {
	if addr == owner {
		return true
	}
	paymentAccount, _ := k.GetPaymentAccount(ctx, addr)
	return paymentAccount.Owner == owner
}
