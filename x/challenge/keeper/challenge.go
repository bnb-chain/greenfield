package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// SaveChallenge saves challenge to the store
func (k Keeper) SaveChallenge(ctx sdk.Context, challenge types.Challenge) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ChallengeKeyPrefix))
	bz := k.cdc.MustMarshal(&challenge)

	store.Set(getChallengeKeyBytes(challenge.Id), bz)
}

// GetChallenge gets a challenge by id
func (k Keeper) GetChallenge(ctx sdk.Context, challengeId uint64) (challenge types.Challenge, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ChallengeKeyPrefix))
	bz := store.Get(getChallengeKeyBytes(challengeId))
	if bz == nil {
		return challenge, false
	}
	k.cdc.MustUnmarshal(bz, &challenge)
	return challenge, true
}

// RemoveChallengeUntil removes challenges which are created earlier
func (k Keeper) RemoveChallengeUntil(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.ChallengeKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var challenge types.Challenge
		k.cdc.MustUnmarshal(iterator.Value(), &challenge)
		if challenge.Id <= challengeId {
			store.Delete(getChallengeKeyBytes(challenge.Id))
		}
	}
}

// getChallengeKeyBytes returns the byte representation of Challenge key
func getChallengeKeyBytes(challengeId uint64) []byte {
	idBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(idBytes, challengeId)

	return idBytes
}

// GetOngoingChallengeId gets the highest challenge id
func (k Keeper) GetOngoingChallengeId(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.OngoingChallengeIdKey)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// SetOngoingChallengeId sets the new challenge id to the store
func (k Keeper) SetOngoingChallengeId(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.OngoingChallengeIdKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(byteKey, bz)
}

// GetAttestChallengeId gets the challenge id of the latest attestation challenge
func (k Keeper) GetAttestChallengeId(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.AttestChallengeIdKey)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// SetAttestChallengeId sets the new id of challenge to the store
func (k Keeper) SetAttestChallengeId(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.AttestChallengeIdKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(byteKey, bz)
}

// GetHeartbeatChallengeId gets the challenge id of the latest heartbeat challenge
func (k Keeper) GetHeartbeatChallengeId(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.HeartbeatChallengeIdKey)
	bz := store.Get(byteKey)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// SetHeartbeatChallengeId sets the new id of challenge to the store
func (k Keeper) SetHeartbeatChallengeId(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	byteKey := types.KeyPrefix(types.HeartbeatChallengeIdKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(byteKey, bz)
}

// GetChallengeCountCurrentBlock gets the count of challenges
func (k Keeper) GetChallengeCountCurrentBlock(ctx sdk.Context) uint64 {
	store := ctx.TransientStore(k.tKey)
	byteKey := types.KeyPrefix(types.CurrentBlockChallengeCountKey)
	bz := store.Get(byteKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// setGetChallengeCountCurrentBlock sets the new count of challenge to the store
func (k Keeper) setGetChallengeCountCurrentBlock(ctx sdk.Context, challengeId uint64) {
	store := ctx.TransientStore(k.tKey)
	byteKey := types.KeyPrefix(types.CurrentBlockChallengeCountKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(byteKey, bz)
}

// IncrChallengeCountCurrentBlock increases the count of challenge by one
func (k Keeper) IncrChallengeCountCurrentBlock(ctx sdk.Context) {
	k.setGetChallengeCountCurrentBlock(ctx, k.GetChallengeCountCurrentBlock(ctx)+1)
}
