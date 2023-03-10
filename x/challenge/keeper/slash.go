package keeper

import (
	"encoding/binary"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// SaveSlash set a specific slash in the store
func (k Keeper) SaveSlash(ctx sdk.Context, slash types.Slash) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SlashKeyPrefix)

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, slash.Height)

	store.Set(getSlashKeyBytes(slash.SpOperatorAddress, slash.ObjectId), heightBytes)
}

// RemoveSlashUntil removes slashes which are created earlier
func (k Keeper) RemoveSlashUntil(ctx sdk.Context, height uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SlashKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		slashHeight := binary.BigEndian.Uint64(iterator.Value())
		if slashHeight <= height {
			store.Delete(iterator.Key())
		}
	}
}

// ExistsSlash check whether there exists recent slash for a pair of sp and object info or not
func (k Keeper) ExistsSlash(ctx sdk.Context, spOperatorAddress sdk.AccAddress, objectId sdkmath.Uint) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.SlashKeyPrefix)

	bz := store.Get(getSlashKeyBytes(spOperatorAddress, objectId))
	return bz != nil
}

// getSlashKeyBytes returns the byte representation of Slash key
func getSlashKeyBytes(spOperatorAddress sdk.AccAddress, objectId sdkmath.Uint) []byte {
	allBytes := append(spOperatorAddress.Bytes(), objectId.Bytes()...)
	return sdk.Keccak256(allBytes)
}
