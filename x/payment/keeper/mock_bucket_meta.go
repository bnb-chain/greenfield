package keeper

import (
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// SetMockBucketMeta set a specific mockBucketMeta in the store from its index
func (k Keeper) SetMockBucketMeta(ctx sdk.Context, mockBucketMeta types.MockBucketMeta) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MockBucketMetaKeyPrefix)
	b := k.cdc.MustMarshal(&mockBucketMeta)
	store.Set(types.MockBucketMetaKey(
		mockBucketMeta.BucketName,
	), b)
}

// GetMockBucketMeta returns a mockBucketMeta from its index
func (k Keeper) GetMockBucketMeta(
	ctx sdk.Context,
	bucketName string,

) (val types.MockBucketMeta, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MockBucketMetaKeyPrefix)

	b := store.Get(types.MockBucketMetaKey(
		bucketName,
	))
	if b == nil {
		return val, false
	}

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMockBucketMeta removes a mockBucketMeta from the store
func (k Keeper) RemoveMockBucketMeta(
	ctx sdk.Context,
	bucketName string,

) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MockBucketMetaKeyPrefix)
	store.Delete(types.MockBucketMetaKey(
		bucketName,
	))
}

// GetAllMockBucketMeta returns all mockBucketMeta
func (k Keeper) GetAllMockBucketMeta(ctx sdk.Context) (list []types.MockBucketMeta) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.MockBucketMetaKeyPrefix)
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.MockBucketMeta
		k.cdc.MustUnmarshal(iterator.Value(), &val)
		list = append(list, val)
	}

	return
}
