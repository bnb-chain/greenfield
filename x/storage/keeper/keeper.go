package keeper

import (
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type (
	Keeper struct {
		cdc           codec.BinaryCodec
		storeKey      storetypes.StoreKey
		memKey        storetypes.StoreKey
		paramStore    paramtypes.Subspace
		spKeeper      types.SpKeeper
		paymentKeeper types.PaymentKeeper

		// sequence
		bucketSeq Sequence
		objectSeq Sequence
		groupSeq  Sequence
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey,
	memKey storetypes.StoreKey,
	ps paramtypes.Subspace,
	spKeeper types.SpKeeper,
	paymentKeeper types.PaymentKeeper,

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := Keeper{
		cdc:           cdc,
		storeKey:      storeKey,
		memKey:        memKey,
		paramStore:    ps,
		spKeeper:      spKeeper,
		paymentKeeper: paymentKeeper,
	}

	k.bucketSeq = NewSequence(types.BucketPrefix)
	k.objectSeq = NewSequence(types.ObjectPrefix)
	k.groupSeq = NewSequence(types.GroupPrefix)
	return &k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) CreateBucket(ctx sdk.Context, bucketInfo types.BucketInfo) error {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bucketKey := types.GetBucketKey(bucketInfo.BucketName)
	if bucketStore.Has(bucketKey) {
		return types.ErrBucketAlreadyExists
	}

	bz := k.cdc.MustMarshal(&bucketInfo)
	bucketStore.Set(bucketKey, bz)
	return nil
}

func (k Keeper) DeleteBucket(ctx sdk.Context, bucketName string) error {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bucketKey := types.GetBucketKey(bucketName)

	// check if the bucket empty
	if k.isEmptyBucket(ctx, bucketKey) {
		return types.ErrBucketNotEmpty
	}
	bucketStore.Delete(bucketKey)
	return nil
}

func (k Keeper) SetBucket(ctx sdk.Context, bucketInfo types.BucketInfo) {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bucketKey := types.GetBucketKey(bucketInfo.BucketName)
	bz := k.cdc.MustMarshal(&bucketInfo)
	bucketStore.Set(bucketKey, bz)
}

func (k Keeper) MustGetBucket(ctx sdk.Context, bucketName string) (bucketInfo types.BucketInfo) {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bz := bucketStore.Get(types.GetBucketKey(bucketName))
	if bz == nil {
		panic(fmt.Sprintf("bucket not found for bucketName: %X\n", bucketName))
	}

	k.cdc.MustUnmarshal(bz, &bucketInfo)
	return bucketInfo
}

func (k Keeper) GetBucket(ctx sdk.Context, bucketName string) (bucketInfo types.BucketInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	bucketKey := types.GetBucketKey(bucketName)
	bz := bucketStore.Get(bucketKey)
	if bz == nil {
		return bucketInfo, false
	}

	k.cdc.MustUnmarshal(bz, &bucketInfo)

	return bucketInfo, true
}

func (k Keeper) GetBucketId(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.bucketSeq.NextVal(store)
	return seq
}

func (k Keeper) GetObjectID(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.objectSeq.NextVal(store)
	return seq
}

func (k Keeper) GetGroupId(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.groupSeq.NextVal(store)
	return seq
}

func (k Keeper) isEmptyBucket(ctx sdk.Context, bucketKey []byte) bool {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	iter := objectStore.Iterator(bucketKey, nil)
	return iter.Valid()
}

func (k Keeper) CreateObject(ctx sdk.Context, objectInfo types.ObjectInfo) error {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	objectKey := types.GetObjectKey(objectInfo.BucketName, objectInfo.ObjectName)
	if objectStore.Has(objectKey) {
		return types.ErrObjectAlreadyExists
	}

	bz := k.cdc.MustMarshal(&objectInfo)
	objectStore.Set(objectKey, bz)
	return nil
}

func (k Keeper) GetObject(ctx sdk.Context, bucketName string, objectName string) (objectInfo types.ObjectInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	objectKey := types.GetObjectKey(bucketName, objectName)
	bz := objectStore.Get(objectKey)
	if bz == nil {
		return objectInfo, false
	}

	k.cdc.MustUnmarshal(bz, &objectInfo)

	return objectInfo, true
}

func (k Keeper) MustGetObject(ctx sdk.Context, bucketName string, objectName string) (objectInfo types.ObjectInfo) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	objectKey := types.GetObjectKey(bucketName, objectName)
	bz := objectStore.Get(objectKey)
	if bz == nil {
		panic(fmt.Sprintf("object not found for bucketName: %X\n", objectName))
	}

	k.cdc.MustUnmarshal(bz, &objectInfo)

	return objectInfo
}

func (k Keeper) SetObject(ctx sdk.Context, objectInfo types.ObjectInfo) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	objectKey := types.GetObjectKey(objectInfo.BucketName, objectInfo.ObjectName)
	bz := k.cdc.MustMarshal(&objectInfo)
	objectStore.Set(objectKey, bz)
}

func (k Keeper) DeleteObject(ctx sdk.Context, bucketName string, objectName string) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)
	objectKey := types.GetObjectKey(bucketName, objectName)

	objectStore.Delete(objectKey)
}

func (k Keeper) CreateGroup(ctx sdk.Context, groupInfo types.GroupInfo) error {
	store := ctx.KVStore(k.storeKey)
	groupStore := prefix.NewStore(store, types.GroupPrefix)

	groupKey := types.GetGroupKey(groupInfo.Owner, groupInfo.GroupName)
	if groupStore.Has(groupKey) {
		return types.ErrGroupAlreadyExists
	}
	bz := k.cdc.MustMarshal(&groupInfo)
	groupStore.Set(groupKey, bz)
	return nil
}

func (k Keeper) GetGroup(ctx sdk.Context, ownerAddr string, groupName string) (groupInfo types.GroupInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	groupStore := prefix.NewStore(store, types.GroupPrefix)

	groupKey := types.GetGroupKey(ownerAddr, groupName)
	bz := groupStore.Get(groupKey)
	if bz == nil {
		return groupInfo, false
	}

	k.cdc.MustUnmarshal(bz, &groupInfo)
	return groupInfo, true
}

func (k Keeper) DeleteGroup(ctx sdk.Context, ownerAddr string, groupName string) error {
	store := ctx.KVStore(k.storeKey)
	groupStore := prefix.NewStore(store, types.GroupPrefix)

	groupKey := types.GetGroupKey(ownerAddr, groupName)
	groupStore.Delete(groupKey)
	return nil
}

func (k Keeper) AddGroupMember(ctx sdk.Context, groupMemberInfo types.GroupMemberInfo) error {
	store := ctx.KVStore(k.storeKey)
	groupMemberStore := prefix.NewStore(store, types.GroupMemberPrefix)

	groupMemberKey := types.GetGroupMemberKey(groupMemberInfo.Id, groupMemberInfo.Member)
	if groupMemberStore.Has(groupMemberKey) {
		return types.ErrGroupMemberAlreadyExists
	}
	bz := k.cdc.MustMarshal(&groupMemberInfo)
	groupMemberStore.Set(groupMemberKey, bz)
	return nil
}

func (k Keeper) RemoveGroupMember(ctx sdk.Context, groupId math.Uint, member string) error {
	store := ctx.KVStore(k.storeKey)
	groupMemberStore := prefix.NewStore(store, types.GroupMemberPrefix)
	memberKey := types.GetGroupMemberKey(groupId, member)
	if groupMemberStore.Has(memberKey) {
		return types.ErrNoSuchGroupMember
	}
	groupMemberStore.Delete(memberKey)
	return nil
}

func (k Keeper) HasGroupMember(ctx sdk.Context, groupMemberKey []byte) bool {
	store := ctx.KVStore(k.storeKey)
	groupMemberStore := prefix.NewStore(store, types.GroupMemberPrefix)

	return groupMemberStore.Has(groupMemberKey)
}

func (k Keeper) VerifySPAndSignature(ctx sdk.Context, spAddr string, sigData []byte, signature []byte) error {
	spAcc, err := sdk.AccAddressFromHexUnsafe(spAddr)
	if err != nil {
		return err
	}
	sp, found := k.spKeeper.GetStorageProvider(ctx, spAcc)
	if !found {
		return types.ErrNoSuchStorageProvider
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return types.ErrStorageProviderNotInService
	}

	approvalAcc, err := sdk.AccAddressFromHexUnsafe(sp.ApprovalAddress)
	if err != nil {
		return err
	}

	err = types.VerifySignature(approvalAcc, sigData, signature)
	if err != nil {
		return err
	}
	return nil
}
