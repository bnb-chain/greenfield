package keeper

import (
	"fmt"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetStreamRecord set a specific streamRecord in the store from its index
func (k Keeper) SetStreamRecord(ctx sdk.Context, streamRecord types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))
	b := k.cdc.MustMarshal(&streamRecord)
	store.Set(types.StreamRecordKey(
		streamRecord.Account,
	), b)
}

// GetStreamRecord returns a streamRecord from its index
func (k Keeper) GetStreamRecord(
	ctx sdk.Context,
	account string,

) (val types.StreamRecord, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))

	b := store.Get(types.StreamRecordKey(
		account,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveStreamRecord removes a streamRecord from the store
func (k Keeper) RemoveStreamRecord(
	ctx sdk.Context,
	account string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))
	store.Delete(types.StreamRecordKey(
		account,
	))
}

// GetAllStreamRecord returns all streamRecord
func (k Keeper) GetAllStreamRecord(ctx sdk.Context) (list []types.StreamRecord) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.StreamRecordKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.StreamRecord
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) UpdateStreamRecord(ctx sdk.Context, streamRecord *types.StreamRecord) {
	currentTimestamp := ctx.BlockTime().Unix()
	timestamp := streamRecord.CrudTimestamp
	if currentTimestamp == timestamp {
		return
	}

	flowDelta := (currentTimestamp - timestamp) * streamRecord.NetflowRate
	streamRecord.StaticBalance += flowDelta
	streamRecord.CrudTimestamp = currentTimestamp
}

func (k Keeper) UpdateStreamRecordByRate(ctx sdk.Context, streamRecord *types.StreamRecord, rate int64) error {
	k.UpdateStreamRecord(ctx, streamRecord)
	streamRecord.NetflowRate += rate
	if rate < 0 {
		reserveTime := k.GetParams(ctx).ReserveTime
		addtionalReserveBalance := -rate * int64(reserveTime)
		if addtionalReserveBalance >= streamRecord.StaticBalance {
			return fmt.Errorf("static balance is not enough, have: %d, need: %d", streamRecord.StaticBalance, addtionalReserveBalance)
		}
		streamRecord.StaticBalance -= addtionalReserveBalance
		streamRecord.BufferBalance += addtionalReserveBalance
	}
	return nil
}
