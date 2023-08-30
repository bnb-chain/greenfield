package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetPaymentAccount set a specific paymentAccount in the store from its index
func (k Keeper) SetPaymentAccount(ctx sdk.Context, paymentAccount *types.PaymentAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)
	key := types.PaymentAccountKey(sdk.MustAccAddressFromHex(paymentAccount.Addr))
	addr := paymentAccount.Addr
	paymentAccount.Addr = ""
	b := k.cdc.MustMarshal(paymentAccount)
	store.Set(key, b)
	_ = ctx.EventManager().EmitTypedEvents(&types.EventPaymentAccountUpdate{
		Addr:       addr,
		Owner:      paymentAccount.Owner,
		Refundable: paymentAccount.Refundable,
	})
}

// GetPaymentAccount returns a paymentAccount from its index
func (k Keeper) GetPaymentAccount(
	ctx sdk.Context,
	addr sdk.AccAddress,
) (val *types.PaymentAccount, found bool) {
	val = &types.PaymentAccount{Addr: addr.String()}
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)

	b := store.Get(types.PaymentAccountKey(
		addr,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, val)
	val.Addr = addr.String()
	return val, true
}

// IsPaymentAccount returns is the account address a payment account
func (k Keeper) IsPaymentAccount(
	ctx sdk.Context,
	addr sdk.AccAddress,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)
	return store.Has(types.PaymentAccountKey(
		addr,
	))
}

// GetAllPaymentAccount returns all paymentAccount
func (k Keeper) GetAllPaymentAccount(ctx sdk.Context) (list []types.PaymentAccount) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.PaymentAccountKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.PaymentAccount
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		val.Addr = sdk.AccAddress(iterator.Key()).String()
		list = append(list, val)
	}

	return
}

func (k Keeper) IsPaymentAccountOwner(ctx sdk.Context, addr, owner sdk.AccAddress) bool {
	if addr.Equals(owner) {
		return true
	}
	paymentAccount, _ := k.GetPaymentAccount(ctx, addr)
	return paymentAccount.Owner == owner.String()
}

func (k Keeper) DerivePaymentAccountAddress(owner sdk.AccAddress, index uint64) sdk.AccAddress {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, index)
	return address.Derive(owner.Bytes(), b)[:sdk.EthAddressLength]
}
