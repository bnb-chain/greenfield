package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
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
	iterator := sdk.KVStorePrefixIterator(store, []byte{})
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

// GetAttestChallengeIds gets the challenge id of the latest attestation challenge
func (k Keeper) GetAttestChallengeIds(ctx sdk.Context) []uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := store.Get(types.AttestChallengeIdKey)

	if bz == nil {
		return []uint64{}
	}

	attestedChallengeIds := new(types.AttestedChallengeIds)
	k.cdc.MustUnmarshal(bz, attestedChallengeIds)

	cq := circularQueue{
		size:  attestedChallengeIds.Size_,
		items: attestedChallengeIds.Ids,
		front: attestedChallengeIds.Cursor,
	}
	return cq.retrieveAll()
}

// AppendAttestChallengeId sets the new id of challenge to the store
func (k Keeper) AppendAttestChallengeId(ctx sdk.Context, challengeId uint64) {
	maxToKeep := k.KeyAttestationKeptCount(ctx)

	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := store.Get(types.AttestChallengeIdKey)

	attestedChallengeIds := new(types.AttestedChallengeIds)
	if bz == nil {
		attestedChallengeIds = &types.AttestedChallengeIds{
			Size_:  maxToKeep,
			Ids:    make([]uint64, maxToKeep),
			Cursor: -1,
		}
	} else {
		k.cdc.MustUnmarshal(bz, attestedChallengeIds)
	}

	cq := circularQueue{
		size:  attestedChallengeIds.Size_,
		items: attestedChallengeIds.Ids,
		front: attestedChallengeIds.Cursor,
	}
	if cq.size != maxToKeep { // happens when the parameter changes
		cq.resize(maxToKeep)
	}
	cq.enqueue(challengeId)
	attestedChallengeIds.Size_ = cq.size
	attestedChallengeIds.Ids = cq.items
	attestedChallengeIds.Cursor = cq.front

	store.Set(types.AttestChallengeIdKey, k.cdc.MustMarshal(attestedChallengeIds))
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
