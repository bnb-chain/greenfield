package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetDelayedWithdrawalRecord set a specific delayedWithdrawal in the store from its index
func (k Keeper) SetDelayedWithdrawalRecord(ctx sdk.Context, delayedWithdrawalRecord *types.DelayedWithdrawalRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedWithdrawalKeyPrefix)
	key := types.DelayedWithdrawalKey(
		sdk.MustAccAddressFromHex(delayedWithdrawalRecord.Addr),
	)

	addr := delayedWithdrawalRecord.Addr
	delayedWithdrawalRecord.Addr = ""
	store.Set(key, k.cdc.MustMarshal(delayedWithdrawalRecord))

	delayedWithdrawalRecord.Addr = addr
}

// GetDelayedWithdrawalRecord returns a delayedWithdrawal from its index
func (k Keeper) GetDelayedWithdrawalRecord(
	ctx sdk.Context,
	addr sdk.AccAddress,
) (*types.DelayedWithdrawalRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedWithdrawalKeyPrefix)

	delayedWithdrawal := &types.DelayedWithdrawalRecord{Addr: addr.String()}
	b := store.Get(types.DelayedWithdrawalKey(
		addr,
	))
	if b == nil {
		return delayedWithdrawal, false
	}

	k.cdc.MustUnmarshal(b, delayedWithdrawal)
	delayedWithdrawal.Addr = addr.String()
	return delayedWithdrawal, true
}

// RemoveDelayedWithdrawalRecord removes a delayedWithdrawal from the store
func (k Keeper) RemoveDelayedWithdrawalRecord(
	ctx sdk.Context,
	addr sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DelayedWithdrawalKeyPrefix)
	store.Delete(types.DelayedWithdrawalKey(
		addr,
	))
}
