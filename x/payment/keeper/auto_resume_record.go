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
) (*types.AutoResumeRecord, bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)

	b := store.Get(types.AutoResumeRecordKey(
		timestamp,
		addr,
	))
	if b == nil {
		return nil, false
	}

	return &types.AutoResumeRecord{
		Timestamp: timestamp,
		Addr:      addr.String(),
	}, true
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

// ExistsAutoResumeRecord checks whether there exists a autoResumeRecord
func (k Keeper) ExistsAutoResumeRecord(
	ctx sdk.Context,
	timestamp int64,
	addr sdk.AccAddress,
) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoResumeRecordKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	exists := false
	for ; iterator.Valid(); iterator.Next() {
		record := types.ParseAutoResumeRecordKey(iterator.Key())
		if record.Timestamp > timestamp {
			break
		}
		if sdk.MustAccAddressFromHex(record.Addr).Equals(addr) {
			exists = true
			break
		}
	}

	return exists
}
