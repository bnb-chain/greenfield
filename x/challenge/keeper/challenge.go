package keeper

import (
	"encoding/binary"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
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

func (k Keeper) encodeUint64(data uint64) []byte {
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, data)
	return bz
}

// GetAttestedChallenges gets the latest attested challenges
func (k Keeper) GetAttestedChallenges(ctx sdk.Context) []*types.AttestedChallenge {
	store := ctx.KVStore(k.storeKey)
	sizeBz := store.Get(types.AttestedChallengesSizeKey)

	if sizeBz == nil {
		return []*types.AttestedChallenge{}
	}

	size := binary.BigEndian.Uint64(sizeBz)
	cursor := binary.BigEndian.Uint64(store.Get(types.AttestedChallengesCursorKey))

	result := []*types.AttestedChallenge{}
	current := cursor
	challengeStore := prefix.NewStore(store, types.AttestedChallengesPrefix)
	for {
		current = (current + 1) % size
		challengeBz := challengeStore.Get(k.encodeUint64(current))
		if challengeBz != nil {
			var challenge types.AttestedChallenge
			k.cdc.MustUnmarshal(challengeBz, &challenge)
			result = append(result, &challenge)
		}
		if current == cursor {
			break
		}
	}
	return result
}

// AppendAttestedChallenge sets the new id of challenge to the store
func (k Keeper) AppendAttestedChallenge(ctx sdk.Context, challenge *types.AttestedChallenge) {
	toKeep := k.GetParams(ctx).AttestationKeptCount

	store := ctx.KVStore(k.storeKey)
	sizeBz := store.Get(types.AttestedChallengesSizeKey)

	challengeStore := prefix.NewStore(store, types.AttestedChallengesPrefix)
	if sizeBz == nil { // the first time to append
		store.Set(types.AttestedChallengesSizeKey, k.encodeUint64(toKeep))
		k.enqueueAttestedChallenge(store, challengeStore, challenge)
		return
	}

	size := binary.BigEndian.Uint64(sizeBz)
	if size != toKeep { // the parameter changes, which is not frequent
		currentChallenges := k.GetAttestedChallenges(ctx)

		iterator := storetypes.KVStorePrefixIterator(challengeStore, []byte{})
		defer iterator.Close()

		for ; iterator.Valid(); iterator.Next() {
			challengeStore.Delete(iterator.Key())
		}

		store.Set(types.AttestedChallengesSizeKey, k.encodeUint64(toKeep))
		store.Delete(types.AttestedChallengesCursorKey)

		for _, c := range currentChallenges {
			k.enqueueAttestedChallenge(store, challengeStore, c)
		}
	}
	k.enqueueAttestedChallenge(store, challengeStore, challenge)
}

func (k Keeper) enqueueAttestedChallenge(store, challengeStore storetypes.KVStore, challenge *types.AttestedChallenge) {
	size := binary.BigEndian.Uint64(store.Get(types.AttestedChallengesSizeKey))
	cursorBz := store.Get(types.AttestedChallengesCursorKey)
	cursor := uint64(0)
	if cursorBz != nil {
		cursor = binary.BigEndian.Uint64(cursorBz)
		cursor = (cursor + 1) % size
	}

	cursorBz = k.encodeUint64(cursor)
	store.Set(types.AttestedChallengesCursorKey, cursorBz)

	challengeStore.Set(cursorBz, k.cdc.MustMarshal(challenge))
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
