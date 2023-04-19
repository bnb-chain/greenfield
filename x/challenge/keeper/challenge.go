package keeper

import (
	"encoding/binary"

	"cosmossdk.io/store/prefix"
	storetypes "cosmossdk.io/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// GetChallengeId gets the challenge id
func (k Keeper) GetChallengeId(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := store.Get(types.ChallengeIdKey)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// setChallengeId sets the new challenge id to the store
func (k Keeper) setChallengeId(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(types.ChallengeIdKey, bz)
}

// SaveChallenge set a specific challenge in the store
func (k Keeper) SaveChallenge(ctx sdk.Context, challenge types.Challenge) {
	k.setChallengeId(ctx, challenge.Id)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ChallengeKeyPrefix)

	heightBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(heightBytes, challenge.ExpiredHeight)

	store.Set(getChallengeKeyBytes(challenge.Id), heightBytes)
}

// RemoveChallengeUntil removes challenges which are expired
func (k Keeper) RemoveChallengeUntil(ctx sdk.Context, height uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ChallengeKeyPrefix)
	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		expiredHeight := binary.BigEndian.Uint64(iterator.Value())
		if expiredHeight <= height {
			store.Delete(iterator.Key())
		}
	}
}

// ExistsChallenge check whether there exists ongoing challenge for an id
func (k Keeper) ExistsChallenge(ctx sdk.Context, challengeId uint64) bool {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.ChallengeKeyPrefix)

	return store.Has(getChallengeKeyBytes(challengeId))
}

// getChallengeKeyBytes returns the byte representation of challenge key
func getChallengeKeyBytes(challengeId uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	return bz
}

// GetAttestChallengeId gets the challenge id of the latest attestation challenge
func (k Keeper) GetAttestChallengeId(ctx sdk.Context) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := store.Get(types.AttestChallengeIdKey)

	if bz == nil {
		return 0
	}

	return binary.BigEndian.Uint64(bz)
}

// SetAttestChallengeId sets the new id of challenge to the store
func (k Keeper) SetAttestChallengeId(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(types.AttestChallengeIdKey, bz)
}

// GetChallengeCountCurrentBlock gets the count of challenges
func (k Keeper) GetChallengeCountCurrentBlock(ctx sdk.Context) uint64 {
	store := ctx.TransientStore(k.tKey)
	bz := store.Get(types.CurrentBlockChallengeCountKey)
	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

// setGetChallengeCountCurrentBlock sets the new count of challenge to the store
func (k Keeper) setGetChallengeCountCurrentBlock(ctx sdk.Context, challengeId uint64) {
	store := ctx.TransientStore(k.tKey)
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(types.CurrentBlockChallengeCountKey, bz)
}

// IncrChallengeCountCurrentBlock increases the count of challenge by one
func (k Keeper) IncrChallengeCountCurrentBlock(ctx sdk.Context) {
	k.setGetChallengeCountCurrentBlock(ctx, k.GetChallengeCountCurrentBlock(ctx)+1)
}
