package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store"
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

func (k Keeper) encodeUint64(data uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, data)
	return bz
}

// GetAttestChallengeIds gets the challenge id of the latest attestation challenge
func (k Keeper) GetAttestChallengeIds(ctx sdk.Context) []uint64 {
	store := ctx.KVStore(k.storeKey)
	sizeBz := store.Get(types.AttestChallengeIdsSizeKey)

	if sizeBz == nil {
		return []uint64{}
	}

	size := binary.BigEndian.Uint64(sizeBz)
	cursor := binary.BigEndian.Uint64(store.Get(types.AttestChallengeIdsCursorKey))

	result := []uint64{}
	current := cursor
	idsStore := prefix.NewStore(store, types.AttestChallengeIdsPrefix)
	for {
		current = (current + 1) % size
		idBz := idsStore.Get(k.encodeUint64(current))
		if idBz != nil {
			result = append(result, binary.BigEndian.Uint64(idBz))
		}
		if current == cursor {
			break
		}
	}
	return result
}

// AppendAttestChallengeId sets the new id of challenge to the store
func (k Keeper) AppendAttestChallengeId(ctx sdk.Context, challengeId uint64) {
	toKeep := k.KeyAttestationKeptCount(ctx)

	store := ctx.KVStore(k.storeKey)
	sizeBz := store.Get(types.AttestChallengeIdsSizeKey)

	idsStore := prefix.NewStore(store, types.AttestChallengeIdsPrefix)
	if sizeBz == nil { // the first time to append
		store.Set(types.AttestChallengeIdsSizeKey, k.encodeUint64(toKeep))
		k.enqueueAttestChallengeId(store, idsStore, challengeId)
		return
	}

	size := binary.BigEndian.Uint64(sizeBz)
	if size != toKeep { // the parameter changes, which is not frequent
		currentIds := k.GetAttestChallengeIds(ctx)

		iterator := sdk.KVStorePrefixIterator(idsStore, []byte{})
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			idsStore.Delete(iterator.Key())
		}

		store.Set(types.AttestChallengeIdsSizeKey, k.encodeUint64(toKeep))
		store.Delete(types.AttestChallengeIdsCursorKey)

		for _, id := range currentIds {
			k.enqueueAttestChallengeId(store, idsStore, id)
		}
	}
	k.enqueueAttestChallengeId(store, idsStore, challengeId)
}

func (k Keeper) enqueueAttestChallengeId(store, idsStore store.KVStore, challengeId uint64) {
	size := binary.BigEndian.Uint64(store.Get(types.AttestChallengeIdsSizeKey))
	cursorBz := store.Get(types.AttestChallengeIdsCursorKey)
	cursor := uint64(0)
	if cursorBz != nil {
		cursor = binary.BigEndian.Uint64(cursorBz)
		cursor = (cursor + 1) % size
	}

	cursorBz = k.encodeUint64(cursor)
	store.Set(types.AttestChallengeIdsCursorKey, cursorBz)

	idsStore.Set(cursorBz, k.encodeUint64(challengeId))
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
