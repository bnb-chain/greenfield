package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetAutoResumeRecord set a specific autoResumeRecord in the store from its index
func (k Keeper) SetAutoResumeRecord(ctx sdk.Context, autoResumeRecord *types.AutoResumeRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)
	b := []byte{0x00}
	store.Set(types.AutoResumeRecordKey(
		autoResumeRecord.Timestamp,
		sdk.MustAccAddressFromHex(autoResumeRecord.Addr),
	), b)
}

// GetAutoResumeRecord returns a autoResumeRecord from its index
func (k Keeper) GetAutoResumeRecord(
	ctx sdk.Context,
	timestamp int64,
	addr sdk.AccAddress,
) (val *types.AutoResumeRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)

	b := store.Get(types.AutoResumeRecordKey(
		timestamp,
		addr,
	))
	if b == nil {
		return val, false
	}

	val = &types.AutoResumeRecord{
		Timestamp: timestamp,
		Addr:      addr.String(),
	}
	return val, true
}

// RemoveAutoResumeRecord removes a autoResumeRecord from the store
func (k Keeper) RemoveAutoResumeRecord(
	ctx sdk.Context,
	timestamp int64,
	addr sdk.AccAddress,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)
	store.Delete(types.AutoResumeRecordKey(
		timestamp,
		addr,
	))
}

// GetAllAutoResumeRecord returns all autoResumeRecord
func (k Keeper) GetAllAutoResumeRecord(ctx sdk.Context) (list []types.AutoResumeRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.ParseAutoResumeRecordKey(iterator.Key())
		list = append(list, val)
	}

	return
}

func (k Keeper) UpdateAutoResumeRecord(ctx sdk.Context, addr sdk.AccAddress, oldTime, newTime int64) {
	if oldTime == newTime {
		return
	}
	if oldTime != 0 {
		k.RemoveAutoResumeRecord(ctx, oldTime, addr)
	}
	if newTime != 0 {
		k.SetAutoResumeRecord(ctx, &types.AutoResumeRecord{
			Timestamp: newTime,
			Addr:      addr.String(),
		})
	}
}
