package keeper

import (
	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetOngoingChallenge set a specific ongoingChallenge in the store from its index
func (k Keeper) SetOngoingChallenge(ctx sdk.Context, ongoingChallenge types.OngoingChallenge) {
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
) (val types.OngoingChallenge, found bool) {
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
func (k Keeper) GetAllOngoingChallenge(ctx sdk.Context) (list []types.OngoingChallenge) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.OngoingChallengeKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.OngoingChallenge
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
