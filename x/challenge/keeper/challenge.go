package keeper

import (
	"encoding/binary"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

// SetOngoingChallenge set a specific ongoingChallenge in the store from its index
func (k Keeper) SetOngoingChallenge(ctx sdk.Context, ongoingChallenge types.Challenge) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OngoingChallengeKeyPrefix))
	b := k.cdc.MustMarshal(&ongoingChallenge)
	store.Set(types.OngoingChallengeKey(
		ongoingChallenge.Id,
	), b)
}

// GetOngoingChallenge returns a ongoingChallenge from its index
func (k Keeper) GetOngoingChallenge(
	ctx sdk.Context,
	id uint64,
) (val types.Challenge, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OngoingChallengeKeyPrefix))

	b := store.Get(types.OngoingChallengeKey(
		id,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveOngoingChallenge removes a ongoingChallenge from the store
func (k Keeper) RemoveOngoingChallenge(
	ctx sdk.Context,
	id uint64,
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OngoingChallengeKeyPrefix))
	store.Delete(types.OngoingChallengeKey(
		id,
	))
}

// GetAllOngoingChallenge returns all ongoingChallenge
func (k Keeper) GetAllOngoingChallenge(ctx sdk.Context) (list []types.Challenge) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OngoingChallengeKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.Challenge
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}

// GetChallengeID gets the highest challenge ID
func (k Keeper) GetChallengeID(ctx sdk.Context) (challengeId uint64, err error) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.ChallengeIdKey)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0, sdkerrors.Wrap(types.ErrInvalidGenesis, "initial challenge ID hasn't been set")
	}

	return binary.BigEndian.Uint64(bz), nil
}

// SetChallengeID sets the new challenge ID to the store
func (k Keeper) SetChallengeID(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.ChallengeIdKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(byteKey, bz)
}

// GetChallengeCount gets the highest challenge ID
func (k Keeper) GetChallengeCount(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.ChallengeCountKey)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// setGetChallengeCount sets the new count of challenge to the store
func (k Keeper) setGetChallengeCount(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.ChallengeCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(byteKey, bz)
}

// ResetChallengeCount sets the count of challenge to zero
func (k Keeper) ResetChallengeCount(ctx sdk.Context) {
	k.setGetChallengeCount(ctx, 0)
}

// IncrChallengeCount increases the count of challenge by one
func (k Keeper) IncrChallengeCount(ctx sdk.Context) {
	k.setGetChallengeCount(ctx, k.GetChallengeCount(ctx)+1)
}
