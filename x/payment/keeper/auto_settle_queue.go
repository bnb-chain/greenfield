package keeper

import (
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetAutoSettleQueue set a specific autoSettleQueue in the store from its index
func (k Keeper) SetAutoSettleQueue(ctx sdk.Context, autoSettleQueue types.AutoSettleQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AutoSettleQueueKeyPrefix))
	b := k.cdc.MustMarshal(&autoSettleQueue)
	store.Set(types.AutoSettleQueueKey(
		autoSettleQueue.Timestamp,
		autoSettleQueue.User,
	), b)
}

// GetAutoSettleQueue returns a autoSettleQueue from its index
func (k Keeper) GetAutoSettleQueue(
	ctx sdk.Context,
	timestamp int64,
	user string,

) (val types.AutoSettleQueue, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AutoSettleQueueKeyPrefix))

	b := store.Get(types.AutoSettleQueueKey(
		timestamp,
		user,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveAutoSettleQueue removes a autoSettleQueue from the store
func (k Keeper) RemoveAutoSettleQueue(
	ctx sdk.Context,
	timestamp int64,
	user string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AutoSettleQueueKeyPrefix))
	store.Delete(types.AutoSettleQueueKey(
		timestamp,
		user,
	))
}

// GetAllAutoSettleQueue returns all autoSettleQueue
func (k Keeper) GetAllAutoSettleQueue(ctx sdk.Context) (list []types.AutoSettleQueue) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.AutoSettleQueueKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.AutoSettleQueue
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

func (k Keeper) UpdateAutoSettleQueue(ctx sdk.Context, user string, oldTime, newTime int64) {
	if oldTime != 0 {
		k.RemoveAutoSettleQueue(ctx, oldTime, user)
	}
	if newTime != 0 {
		k.SetAutoSettleQueue(ctx, types.AutoSettleQueue{
			Timestamp: newTime,
			User:      user,
		})
	}
}
