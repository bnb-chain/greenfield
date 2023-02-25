package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// SaveSlash set a specific slash in the store
func (k Keeper) SaveSlash(ctx sdk.Context, slash types.Slash) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SlashKeyPrefix))
	bz := k.cdc.MustMarshal(&slash)

	store.Set(getSlashKeyBytes(slash.SpOperatorAddress, slash.ObjectId), bz)
}

// SaveChallenge saves challenge to the store
func (k Keeper) RemoveSlashUntil(ctx sdk.Context, height uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SlashKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var slash types.Slash
		k.cdc.MustUnmarshal(iterator.Value(), &slash)
		if slash.Height <= height {
			store.Delete(getSlashKeyBytes(slash.SpOperatorAddress, slash.ObjectId))
		}
	}
}

// ExistsSlash check whether there exists recent slash for a pair of sp and object info or not
func (k Keeper) ExistsSlash(ctx sdk.Context, spOperatorAddress string, objectId uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.SlashKeyPrefix))

	bz := store.Get(getSlashKeyBytes(spOperatorAddress, objectId))
	return bz != nil
}

// getSlashKeyBytes returns the byte representation of Slash key
func getSlashKeyBytes(spOperatorAddress string, objectId uint64) []byte {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, objectId)

	return append(sdk.Keccak256([]byte(spOperatorAddress)), idBytes...)
}
