package keeper

import (
	"fmt"

	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"
)

type (
	Keeper struct {
		cdc        codec.BinaryCodec
		storeKey   storetypes.StoreKey
		memKey     storetypes.StoreKey
		paramstore paramtypes.Subspace
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	return &Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
	}
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) GetBucket(ctx sdk.Context, bucketKey []byte) (bucketInfo types.BucketInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bz := bucketStore.Get(bucketKey)
	if bz == nil {
		return bucketInfo, false
	}

	err := k.cdc.Unmarshal(bz, &bucketInfo)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal bucketInfo value %v", err))
	}

	return bucketInfo, true
}

func (k Keeper) HasBucket(ctx sdk.Context, bucketKey []byte) (found bool) {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	return bucketStore.Has(bucketKey)
}

func (k Keeper) SetBucket(ctx sdk.Context, bucketKey []byte, bucketInfo types.BucketInfo) {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bz := k.cdc.MustMarshal(&bucketInfo)
	bucketStore.Set(bucketKey, bz)
}

func (k Keeper) DeleteBucket(ctx sdk.Context, bucketKey []byte) {
  store := ctx.KVStore(k.storeKey)
  bucketStore := prefix.NewStore(store, types.BucketPrefix)

  bucketStore.Delete(bucketKey)
}

func (k Keeper) IsEmptyBucket(ctx sdk.Context, bucketKey []byte) bool {
  store := ctx.KVStore(k.storeKey)
  objectStore := prefix.NewStore(store, types.ObjectPrefix)

  iter := objectStore.Iterator(bucketKey, nil)
  if iter.Valid() {
    return true
  }
  return false
} 

func (k Keeper) GetObject(ctx sdk.Context, objectKey []byte) (objectInfo types.ObjectInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	bz := objectStore.Get(objectKey)
	if bz == nil {
		return objectInfo, false
	}

	err := k.cdc.Unmarshal(bz, &objectInfo)
	if err != nil {
		panic(fmt.Errorf("unable to unmarshal bucketInfo value %v", err))
	}

	return objectInfo, true
}

func (k Keeper) HasObject(ctx sdk.Context, objectKey []byte) (found bool) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	return objectStore.Has(objectKey)
}

func (k Keeper) SetObject(ctx sdk.Context, objectKey []byte, objectInfo types.ObjectInfo) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	bz := k.cdc.MustMarshal(&objectInfo)
	objectStore.Set(objectKey, bz)
}

func (k Keeper) DeleteObject(ctx sdk.Context, objectKey []byte) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

  objectStore.Delete(objectKey)
}

func (k Keeper) SetGroup(ctx sdk.Context, groupKey []byte, groupInfo types.GroupInfo) {
  store := ctx.KVStore(k.storeKey)
  groupStore := prefix.NewStore(store, types.GroupPrefix)
  
  bz := k.cdc.MustMarshal(&groupInfo)
  groupStore.Set(groupKey, bz);
}

func (k Keeper) GetGroup(ctx sdk.Context, groupKey []byte) (groupInfo types.GroupInfo, found bool) {
  store := ctx.KVStore(k.storeKey)
  groupStore := prefix.NewStore(store, types.GroupPrefix)
  
  bz := groupStore.Get(groupKey)
  if bz == nil {
    return groupInfo, false
  }

  err := k.cdc.Unmarshal(bz, &groupInfo)
  if err != nil {
		panic(fmt.Errorf("unable to unmarshal groupInfo value %v", err))
  }
  return groupInfo, true
}

func (k Keeper) HasGroup(ctx sdk.Context, groupKey []byte) (found bool) {
  store := ctx.KVStore(k.storeKey)
  groupStore := prefix.NewStore(store, types.GroupPrefix)
  
  return groupStore.Has(groupKey)
}

func (k Keeper) DeleteGroup(ctx sdk.Context, groupKey []byte) {
  store := ctx.KVStore(k.storeKey)
  groupStore := prefix.NewStore(store, types.GroupPrefix)
  
  groupStore.Delete(groupKey)
}

func (k Keeper) SetGroupMember(ctx sdk.Context, groupMemberKey []byte, groupMemberInfo types.GroupMemberInfo) {
  store := ctx.KVStore(k.storeKey)
  groupMemberStore := prefix.NewStore(store, types.GroupMemberPrefix)

  bz := k.cdc.MustMarshal(&groupMemberInfo)
  groupMemberStore.Set(groupMemberKey, bz)
}

func (k Keeper) HasGroupMember(ctx sdk.Context, groupMemberKey []byte) bool {
  store := ctx.KVStore(k.storeKey)
  groupMemberStore := prefix.NewStore(store, types.GroupMemberPrefix)

  return groupMemberStore.Has(groupMemberKey)
}

func (k Keeper) DeleteGroupMember(ctx sdk.Context, groupMemberKey []byte) {
  store := ctx.KVStore(k.storeKey)
  groupMemberStore := prefix.NewStore(store, types.GroupMemberPrefix)

  groupMemberStore.Delete(groupMemberKey)
}
