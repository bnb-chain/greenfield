package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
)

// SetMockObjectInfo set a specific mockObjectInfo in the store from its index
func (k Keeper) SetMockObjectInfo(ctx sdk.Context, mockObjectInfo types.MockObjectInfo) {
	store :=  prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MockObjectInfoKeyPrefix))
	b := k.cdc.MustMarshal(&mockObjectInfo)
	store.Set(types.MockObjectInfoKey(
        mockObjectInfo.BucketName,
    mockObjectInfo.ObjectName,
    ), b)
}

// GetMockObjectInfo returns a mockObjectInfo from its index
func (k Keeper) GetMockObjectInfo(
    ctx sdk.Context,
    bucketName string,
    objectName string,
    
) (val types.MockObjectInfo, found bool) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MockObjectInfoKeyPrefix))

	b := store.Get(types.MockObjectInfoKey(
        bucketName,
    objectName,
    ))
    if b == nil {
        return val, false
    }

	k.cdc.MustUnmarshal(b, &val)
	return val, true
}

// RemoveMockObjectInfo removes a mockObjectInfo from the store
func (k Keeper) RemoveMockObjectInfo(
    ctx sdk.Context,
    bucketName string,
    objectName string,
    
) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MockObjectInfoKeyPrefix))
	store.Delete(types.MockObjectInfoKey(
	    bucketName,
    objectName,
    ))
}

// GetAllMockObjectInfo returns all mockObjectInfo
func (k Keeper) GetAllMockObjectInfo(ctx sdk.Context) (list []types.MockObjectInfo) {
    store := prefix.NewStore(ctx.KVStore(k.storeKey), types.KeyPrefix(types.MockObjectInfoKeyPrefix))
	iterator := sdk.KVStorePrefixIterator(store, []byte{})

	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		var val types.MockObjectInfo
		k.cdc.MustUnmarshal(iterator.Value(), &val)
        list = append(list, val)
	}

    return
}
