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
		return 1 // make challenge id starts from 1
	}

	return binary.BigEndian.Uint64(bz)
}

// SetChallengeId sets the new challenge id to the store
func (k Keeper) SetChallengeId(ctx sdk.Context, challengeId uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), []byte{})
	bz := make([]byte, 8)
	binary.BigEndian.PutUint64(bz, challengeId)
	store.Set(types.ChallengeIdKey, bz)
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
