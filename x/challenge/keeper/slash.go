package keeper

import (
	"bytes"
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// GetRecentSlashCount get the total number of recentSlash
func (k Keeper) GetRecentSlashCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.RecentSlashCountKey)
	bz := store.Get(byteKey)

	// Count doesn't exist: no element
	if bz == nil {
		return 0
	}

	// Parse bytes
	return binary.BigEndian.Uint64(bz)
}

// SetRecentSlashCount set the total number of recentSlash
func (k Keeper) SetRecentSlashCount(ctx sdk.Context, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.RecentSlashCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, count)
	store.Set(byteKey, bz)
}

// AppendRecentSlash appends a recentSlash in the store with a new id and update the count
func (k Keeper) AppendRecentSlash(
	ctx sdk.Context,
	recentSlash types.Slash,
) uint64 {
	// Create the recentSlash
	count := k.GetRecentSlashCount(ctx)

	// Set the ID of the appended value
	recentSlash.Id = count

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecentSlashKey))
	appendedValue := k.cdc.MustMarshal(&recentSlash)
	store.Set(GetRecentSlashIDBytes(recentSlash.Id), appendedValue)

	// Update recentSlash count
	k.SetRecentSlashCount(ctx, count+1)

	return count
}

// SetRecentSlash set a specific recentSlash in the store
func (k Keeper) SetRecentSlash(ctx sdk.Context, recentSlash types.Slash) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecentSlashKey))
	b := k.cdc.MustMarshal(&recentSlash)
	store.Set(GetRecentSlashIDBytes(recentSlash.Id), b)
}

// GetRecentSlash returns a recentSlash from its id
func (k Keeper) GetRecentSlash(ctx sdk.Context, id uint64) (val types.Slash, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecentSlashKey))
	b := store.Get(GetRecentSlashIDBytes(id))
	if b == nil {
		return val, false
	}
	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveRecentSlash removes a recentSlash from the store
func (k Keeper) RemoveRecentSlash(ctx sdk.Context, id uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecentSlashKey))
	store.Delete(GetRecentSlashIDBytes(id))
}

// GetAllRecentSlash returns all recentSlash
func (k Keeper) GetAllRecentSlash(ctx sdk.Context) (list []types.Slash) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecentSlashKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Slash
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// ExistsSlash check whether there exists recent slash for a pair of sp and object info or not
func (k Keeper) ExistsSlash(ctx sdk.Context, spAddress string, objectKey []byte) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.RecentSlashKey))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Slash
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		if val.SpOperatorAddress == spAddress && bytes.Equal(val.ObjectKey, objectKey) {
			return true
		}
	}
	return false
}

// GetRecentSlashIDBytes returns the byte representation of the ID
func GetRecentSlashIDBytes(id uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, id)
	return bz
}

// GetRecentSlashIDFromBytes returns ID in uint64 format from a byte array
func GetRecentSlashIDFromBytes(bz []byte) uint64 {
	return binary.BigEndian.Uint64(bz)
}
