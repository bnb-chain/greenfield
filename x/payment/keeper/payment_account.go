package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetPaymentAccount set a specific paymentAccount in the store from its index
func (k Keeper) SetPaymentAccount(ctx sdk.Context, paymentAccount types.PaymentAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)
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
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)

	b := store.Get(types.PaymentAccountKey(
		addr,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// GetAllPaymentAccount returns all paymentAccount
func (k Keeper) GetAllPaymentAccount(ctx sdk.Context) (list []types.PaymentAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)
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
