package keeper

import (
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

// SetAutoSettleRecord set a specific autoSettleRecord in the store from its index
func (k Keeper) SetAutoSettleRecord(ctx sdk.Context, autoSettleRecord types.AutoSettleRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoSettleRecordKeyPrefix)
	b := []byte{0x00}
	store.Set(types.AutoSettleRecordKey(
		autoSettleRecord.Timestamp,
		autoSettleRecord.Addr,
	), b)
}

// GetAutoSettleRecord returns a autoSettleRecord from its index
func (k Keeper) GetAutoSettleRecord(
	ctx sdk.Context,
	timestamp int64,
	addr string,
) (val types.AutoSettleRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoSettleRecordKeyPrefix)

	b := store.Get(types.AutoSettleRecordKey(
		timestamp,
		addr,
	))
	if b == nil {
		return val, false
	}

	val.Timestamp = timestamp
	val.Addr = addr
	return val, true
}

// RemoveAutoSettleRecord removes a autoSettleRecord from the store
func (k Keeper) RemoveAutoSettleRecord(
	ctx sdk.Context,
	timestamp int64,
	addr string,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoSettleRecordKeyPrefix)
	store.Delete(types.AutoSettleRecordKey(
		timestamp,
		addr,
	))
}

// GetAllAutoSettleRecord returns all autoSettleRecord
func (k Keeper) GetAllAutoSettleRecord(ctx sdk.Context) (list []types.AutoSettleRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.AutoSettleRecordKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		val := types.ParseAutoSettleRecordKey(iterator.Key())
		list = append(list, val)
	}

	return
}

func (k Keeper) UpdateAutoSettleRecord(ctx sdk.Context, addr string, oldTime, newTime int64) {
	if oldTime == newTime {
		return
	}
	if oldTime != 0 {
		k.RemoveAutoSettleRecord(ctx, oldTime, addr)
	}
	if newTime != 0 {
		k.SetAutoSettleRecord(ctx, types.AutoSettleRecord{
			Timestamp: newTime,
			Addr:      addr,
		})
	}
}
