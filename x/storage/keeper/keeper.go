package keeper

import (
	"cosmossdk.io/math"
	"fmt"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
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
		spKeeper   types.SpKeeper

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

) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := Keeper{
		cdc:        cdc,
		storeKey:   storeKey,
		memKey:     memKey,
		paramstore: ps,
		spKeeper:   spKeeper,
	}

	k.bucketSeq = NewSequence(types.BucketPrefix)
	k.objectSeq = NewSequence(types.ObjectPrefix)
	k.groupSeq = NewSequence(types.GroupPrefix)
	return &k
}

func (k Keeper) CheckSPAndSignature(ctx sdk.Context, spAddrs []string, sigData [][]byte, signature [][]byte) error {
	for i, spAddr := range spAddrs {
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

		err = types.VerifySignature(approvalAcc, sigData[i], signature[i])
		if err != nil {
			return err
		}
	}
	return nil
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
	return iter.Valid()
}

func (k Keeper) GetObject(ctx sdk.Context, objectKey []byte) (objectInfo types.ObjectInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	bz := objectStore.Get(objectKey)
	if bz == nil {
		return objectInfo, false
	}

	k.cdc.MustUnmarshal(bz, &objectInfo)

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
	groupStore.Set(groupKey, bz)
}

func (k Keeper) GetGroup(ctx sdk.Context, groupKey []byte) (groupInfo types.GroupInfo, found bool) {
	store := ctx.KVStore(k.storeKey)
	groupStore := prefix.NewStore(store, types.GroupPrefix)

	bz := groupStore.Get(groupKey)
	if bz == nil {
		return groupInfo, false
	}

	k.cdc.MustUnmarshal(bz, &groupInfo)
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
