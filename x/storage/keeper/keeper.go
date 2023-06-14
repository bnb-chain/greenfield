package keeper

import (
	"encoding/binary"
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/internal/sequence"
	"github.com/bnb-chain/greenfield/types/resource"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type (
	Keeper struct {
		cdc              codec.BinaryCodec
		storeKey         storetypes.StoreKey
		tStoreKey        storetypes.StoreKey
		spKeeper         types.SpKeeper
		paymentKeeper    types.PaymentKeeper
		accountKeeper    types.AccountKeeper
		permKeeper       types.PermissionKeeper
		crossChainKeeper types.CrossChainKeeper

		// sequence
		bucketSeq        sequence.U256
		objectSeq        sequence.U256
		groupSeq         sequence.U256
		executionTaskSeq sequence.U256

		authority string
	}
)

func NewKeeper(
	cdc codec.BinaryCodec,
	storeKey storetypes.StoreKey,
	tStoreKey storetypes.StoreKey,
	accountKeeper types.AccountKeeper,
	spKeeper types.SpKeeper,
	paymentKeeper types.PaymentKeeper,
	permKeeper types.PermissionKeeper,
	crossChainKeeper types.CrossChainKeeper,
	authority string,
) *Keeper {

	k := Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		tStoreKey:        tStoreKey,
		accountKeeper:    accountKeeper,
		spKeeper:         spKeeper,
		paymentKeeper:    paymentKeeper,
		permKeeper:       permKeeper,
		crossChainKeeper: crossChainKeeper,
		authority:        authority,
	}

	k.bucketSeq = sequence.NewSequence256(types.BucketSequencePrefix)
	k.objectSeq = sequence.NewSequence256(types.ObjectSequencePrefix)
	k.groupSeq = sequence.NewSequence256(types.GroupSequencePrefix)
	k.executionTaskSeq = sequence.NewSequence256(types.ExecutionTaskSequencePrefix)
	return &k
}

func (k Keeper) GetAuthority() string {
	return k.authority
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) CreateBucket(
	ctx sdk.Context, ownerAcc sdk.AccAddress, bucketName string,
	primarySpAcc sdk.AccAddress, opts *CreateBucketOptions) (sdkmath.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	// check if the bucket exist
	bucketKey := types.GetBucketKey(bucketName)
	if store.Has(bucketKey) {
		return sdkmath.ZeroUint(), types.ErrBucketAlreadyExists
	}

	// check payment account
	paymentAcc, err := k.VerifyPaymentAccount(ctx, opts.PaymentAddress, ownerAcc)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	// check primary sp approval
	if opts.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return sdkmath.ZeroUint(), errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}
	err = k.VerifySPAndSignature(ctx, primarySpAcc, opts.ApprovalMsgBytes, opts.PrimarySpApproval.Sig)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	bucketInfo := types.BucketInfo{
		Owner:            ownerAcc.String(),
		BucketName:       bucketName,
		Visibility:       opts.Visibility,
		CreateAt:         ctx.BlockTime().Unix(),
		SourceType:       opts.SourceType,
		BucketStatus:     types.BUCKET_STATUS_CREATED,
		ChargedReadQuota: opts.ChargedReadQuota,
		PaymentAddress:   paymentAcc.String(),
		PrimarySpAddress: primarySpAcc.String(),
	}

	// charge by read quota
	if opts.ChargedReadQuota != 0 {
		err = k.ChargeInitialReadFee(ctx, &bucketInfo)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
	}

	// Generate bucket Id
	bucketInfo.Id = k.GenNextBucketId(ctx)

	// store the bucket
	bz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(bucketKey, sequence.EncodeSequence(bucketInfo.Id))
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bz)

	// emit CreateBucket Event
	if err = ctx.EventManager().EmitTypedEvents(&types.EventCreateBucket{
		Owner:            bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		Visibility:       bucketInfo.Visibility,
		CreateAt:         bucketInfo.CreateAt,
		BucketId:         bucketInfo.Id,
		SourceType:       bucketInfo.SourceType,
		Status:           bucketInfo.BucketStatus,
		ChargedReadQuota: bucketInfo.ChargedReadQuota,
		PaymentAddress:   bucketInfo.PaymentAddress,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
	}); err != nil {
		return sdkmath.Uint{}, err
	}
	return bucketInfo.Id, nil
}

func (k Keeper) DeleteBucket(ctx sdk.Context, operator sdk.AccAddress, bucketName string, opts DeleteBucketOptions) error {
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	if bucketInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	// check permission
	effect := k.VerifyBucketPermission(ctx, bucketInfo, operator, permtypes.ACTION_DELETE_BUCKET, nil)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf("The operator(%s) has no DeleteBucket permission of the bucket(%s)",
			operator.String(), bucketName)
	}

	// check if the bucket empty
	if k.isNonEmptyBucket(ctx, bucketName) {
		return types.ErrBucketNotEmpty
	}

	// change the bill
	err := k.ChargeDeleteBucket(ctx, bucketInfo)
	if err != nil {
		return types.ErrChargeFailed.Wrapf("ChargeDeleteBucket error: %s", err)
	}

	return k.doDeleteBucket(ctx, operator, bucketInfo)
}

func (k Keeper) doDeleteBucket(ctx sdk.Context, operator sdk.AccAddress, bucketInfo *types.BucketInfo) error {
	store := ctx.KVStore(k.storeKey)
	bucketKey := types.GetBucketKey(bucketInfo.BucketName)
	store.Delete(bucketKey)
	store.Delete(types.GetBucketByIDKey(bucketInfo.Id))

	if err := k.appendResourceIdForGarbageCollection(ctx, resource.RESOURCE_TYPE_BUCKET, bucketInfo.Id); err != nil {
		return err
	}
	err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteBucket{
		Operator:         operator.String(),
		Owner:            bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		BucketId:         bucketInfo.Id,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
	})
	return err
}

// ForceDeleteBucket will delete bucket without permission check, it is used for discontinue request from sps.
// The cap parameter will limit the max objects can be deleted in the call.
// It will also return 1) whether the bucket is deleted, 2) the objects deleted, and 3) error if there is
func (k Keeper) ForceDeleteBucket(ctx sdk.Context, bucketId sdkmath.Uint, cap uint64) (bool, uint64, error) {
	bucketInfo, found := k.GetBucketInfoById(ctx, bucketId)
	if !found { // the bucket is already deleted
		return true, 0, nil
	}

	bucketDeleted := false
	sp := sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)

	store := ctx.KVStore(k.storeKey)
	objectPrefixStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(bucketInfo.BucketName))
	iter := objectPrefixStore.Iterator(nil, nil)
	defer iter.Close()

	deleted := uint64(0) // deleted object count
	for ; iter.Valid(); iter.Next() {
		if deleted >= cap {
			break
		}

		bz := store.Get(types.GetObjectByIDKey(types.DecodeSequence(iter.Value())))
		if bz == nil {
			panic("should not happen")
		}

		var objectInfo types.ObjectInfo
		k.cdc.MustUnmarshal(bz, &objectInfo)

		if objectInfo.ObjectStatus == types.OBJECT_STATUS_CREATED {
			if err := k.UnlockStoreFee(ctx, bucketInfo, &objectInfo); err != nil {
				ctx.Logger().Error("unlock store fee error", "err", err)
				return false, deleted, err
			}
		} else if objectInfo.ObjectStatus == types.OBJECT_STATUS_SEALED {
			if err := k.ChargeDeleteObject(ctx, bucketInfo, &objectInfo); err != nil {
				ctx.Logger().Error("charge delete object error", "err", err)
				return false, deleted, err
			}
		}
		if err := k.doDeleteObject(ctx, sp, bucketInfo, &objectInfo); err != nil {
			ctx.Logger().Error("do delete object err", "err", err)
			return false, deleted, err
		}
		deleted++
	}

	if !iter.Valid() {
		if err := k.ChargeDeleteBucket(ctx, bucketInfo); err != nil {
			ctx.Logger().Error("charge delete bucket error", "err", err)
			return false, deleted, err
		}

		if err := k.doDeleteBucket(ctx, sp, bucketInfo); err != nil {
			ctx.Logger().Error("do delete bucket error", "err", err)
			return false, deleted, err
		}
		bucketDeleted = true
	}

	return bucketDeleted, deleted, nil
}

func (k Keeper) UpdateBucketInfo(ctx sdk.Context, operator sdk.AccAddress, bucketName string, opts UpdateBucketOptions) error {
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	// check bucket source
	if bucketInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	// check permission
	effect := k.VerifyBucketPermission(ctx, bucketInfo, operator, permtypes.ACTION_UPDATE_BUCKET_INFO, nil)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf("The operator(%s) has no UpdateBucketInfo permission of the bucket(%s)",
			operator.String(), bucketName)
	}

	// handle fields not changed
	if opts.ChargedReadQuota == nil {
		opts.ChargedReadQuota = &bucketInfo.ChargedReadQuota
	}

	if opts.Visibility != types.VISIBILITY_TYPE_UNSPECIFIED {
		bucketInfo.Visibility = opts.Visibility
	}

	var paymentAcc sdk.AccAddress
	var err error
	if opts.PaymentAddress != "" {
		ownerAcc := sdk.MustAccAddressFromHex(bucketInfo.Owner)
		paymentAcc, err = k.VerifyPaymentAccount(ctx, opts.PaymentAddress, ownerAcc)
		if err != nil {
			return err
		}
	} else {
		paymentAcc = sdk.MustAccAddressFromHex(bucketInfo.PaymentAddress)
	}
	err = k.UpdateBucketInfoAndCharge(ctx, bucketInfo, paymentAcc.String(), *opts.ChargedReadQuota)
	if err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateBucketInfo{
		Operator:               operator.String(),
		BucketName:             bucketName,
		BucketId:               bucketInfo.Id,
		ChargedReadQuotaBefore: bucketInfo.ChargedReadQuota,
		ChargedReadQuotaAfter:  *opts.ChargedReadQuota,
		PaymentAddressBefore:   bucketInfo.PaymentAddress,
		PaymentAddressAfter:    paymentAcc.String(),
		Visibility:             bucketInfo.Visibility,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) DiscontinueBucket(ctx sdk.Context, operator sdk.AccAddress, bucketName, reason string) error {
	sp, found := k.spKeeper.GetStorageProviderByGcAddr(ctx, operator)
	if !found {
		return types.ErrNoSuchStorageProvider.Wrapf("SP operator address: %s", operator.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}

	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	if bucketInfo.BucketStatus == types.BUCKET_STATUS_DISCONTINUED {
		return types.ErrInvalidBucketStatus
	}

	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "only primary sp is allowed to do discontinue bucket")
	}

	count := k.getDiscontinueBucketCount(ctx, operator)
	max := k.DiscontinueBucketMax(ctx)
	if count+1 > max {
		return types.ErrNoMoreDiscontinue.Wrapf("no more buckets can be requested in this window")
	}

	bucketInfo.BucketStatus = types.BUCKET_STATUS_DISCONTINUED

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bz)

	deleteAt := ctx.BlockTime().Unix() + k.DiscontinueConfirmPeriod(ctx)

	k.appendDiscontinueBucketIds(ctx, deleteAt, []sdkmath.Uint{bucketInfo.Id})
	k.setDiscontinueBucketCount(ctx, operator, count+1)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDiscontinueBucket{
		BucketId:   bucketInfo.Id,
		BucketName: bucketInfo.BucketName,
		Reason:     reason,
		DeleteAt:   deleteAt,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) SetBucketInfo(ctx sdk.Context, bucketInfo *types.BucketInfo) {
	store := ctx.KVStore(k.storeKey)

	bz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bz)
}

func (k Keeper) GetBucketInfo(ctx sdk.Context, bucketName string) (*types.BucketInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	bucketKey := types.GetBucketKey(bucketName)
	bz := store.Get(bucketKey)
	if bz == nil {
		return nil, false
	}

	return k.GetBucketInfoById(ctx, sequence.DecodeSequence(bz))
}

func (k Keeper) GetBucketInfoById(ctx sdk.Context, bucketId sdkmath.Uint) (*types.BucketInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetBucketByIDKey(bucketId))
	if bz == nil {
		return nil, false
	}

	var bucketInfo types.BucketInfo
	k.cdc.MustUnmarshal(bz, &bucketInfo)

	return &bucketInfo, true
}

func (k Keeper) CreateObject(
	ctx sdk.Context, operator sdk.AccAddress, bucketName, objectName string,
	payloadSize uint64, opts CreateObjectOptions) (sdkmath.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	// check payload size
	if payloadSize > k.MaxPayloadSize(ctx) {
		return sdkmath.ZeroUint(), types.ErrTooLargeObject
	}

	// check bucket
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return sdkmath.ZeroUint(), types.ErrNoSuchBucket
	}
	if bucketInfo.BucketStatus == types.BUCKET_STATUS_DISCONTINUED {
		return sdkmath.ZeroUint(), types.ErrBucketDiscontinued
	}

	// verify permission
	verifyOpts := &permtypes.VerifyOptions{
		WantedSize: &payloadSize,
	}
	effect := k.VerifyBucketPermission(ctx, bucketInfo, operator, permtypes.ACTION_CREATE_OBJECT, verifyOpts)
	if effect != permtypes.EFFECT_ALLOW {
		return sdkmath.ZeroUint(), types.ErrAccessDenied.Wrapf("The operator(%s) has no CreateObject permission of the bucket(%s)",
			operator.String(), bucketName)
	}

	// check secondary sps
	var secondarySPs []string
	for _, sp := range opts.SecondarySpAddresses {
		spAcc := sdk.MustAccAddressFromHex(sp)
		err := k.spKeeper.IsStorageProviderExistAndInService(ctx, spAcc)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
		secondarySPs = append(secondarySPs, spAcc.String())
	}

	// We use the last address in SecondarySpAddresses to store the creator so that it can be identified when canceling create
	if !operator.Equals(sdk.MustAccAddressFromHex(bucketInfo.Owner)) {
		secondarySPs = append(secondarySPs, operator.String())
	}

	// check approval
	if opts.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return sdkmath.ZeroUint(), errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}

	err := k.VerifySPAndSignature(ctx, sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress), opts.ApprovalMsgBytes,
		opts.PrimarySpApproval.Sig)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	objectKey := types.GetObjectKey(bucketName, objectName)
	if store.Has(objectKey) {
		return sdkmath.ZeroUint(), types.ErrObjectAlreadyExists
	}

	// check payload size, the empty object doesn't need sealed
	var objectStatus types.ObjectStatus
	if payloadSize == 0 {
		// empty object does not interact with sp
		objectStatus = types.OBJECT_STATUS_SEALED
	} else {
		objectStatus = types.OBJECT_STATUS_CREATED
	}

	// construct objectInfo
	objectInfo := types.ObjectInfo{
		Owner:                bucketInfo.Owner,
		BucketName:           bucketName,
		ObjectName:           objectName,
		PayloadSize:          payloadSize,
		Visibility:           opts.Visibility,
		ContentType:          opts.ContentType,
		Id:                   k.GenNextObjectID(ctx),
		CreateAt:             ctx.BlockTime().Unix(),
		ObjectStatus:         objectStatus,
		RedundancyType:       opts.RedundancyType,
		SourceType:           opts.SourceType,
		Checksums:            opts.Checksums,
		SecondarySpAddresses: secondarySPs,
	}

	if objectInfo.PayloadSize == 0 {
		// charge directly without lock charge
		err = k.ChargeStoreFee(ctx, bucketInfo, &objectInfo)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
	} else {
		// Lock Fee
		err = k.LockStoreFee(ctx, bucketInfo, &objectInfo)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
	}

	bbz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(objectKey, sequence.EncodeSequence(objectInfo.Id))
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateObject{
		Creator:          operator.String(),
		Owner:            objectInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		BucketId:         bucketInfo.Id,
		ObjectId:         objectInfo.Id,
		CreateAt:         objectInfo.CreateAt,
		PayloadSize:      objectInfo.PayloadSize,
		Visibility:       objectInfo.Visibility,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
		ContentType:      objectInfo.ContentType,
		Status:           objectInfo.ObjectStatus,
		RedundancyType:   objectInfo.RedundancyType,
		SourceType:       objectInfo.SourceType,
		Checksums:        objectInfo.Checksums,
	}); err != nil {
		return objectInfo.Id, err
	}
	return objectInfo.Id, nil
}

func (k Keeper) SetObjectInfo(ctx sdk.Context, objectInfo *types.ObjectInfo) {
	store := ctx.KVStore(k.storeKey)

	obz := k.cdc.MustMarshal(objectInfo)
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)
}

func (k Keeper) GetObjectInfoCount(ctx sdk.Context) sdkmath.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.objectSeq.CurVal(store)
	return seq
}

func (k Keeper) GetObjectInfo(ctx sdk.Context, bucketName string, objectName string) (*types.ObjectInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetObjectKey(bucketName, objectName))
	if bz == nil {
		return nil, false
	}

	return k.GetObjectInfoById(ctx, sequence.DecodeSequence(bz))
}

func (k Keeper) GetObjectInfoById(ctx sdk.Context, objectId sdkmath.Uint) (*types.ObjectInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetObjectByIDKey(objectId))
	if bz == nil {
		return nil, false
	}

	var objectInfo types.ObjectInfo
	k.cdc.MustUnmarshal(bz, &objectInfo)
	return &objectInfo, true
}

type SealObjectOptions struct {
	SecondarySpAddresses  []string
	SecondarySpSignatures [][]byte
}

func (k Keeper) SealObject(
	ctx sdk.Context, spSealAcc sdk.AccAddress,
	bucketName, objectName string, opts SealObjectOptions) error {

	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}

	sp, found := k.spKeeper.GetStorageProviderBySealAddr(ctx, spSealAcc)
	if !found {
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "SP seal address: %s", spSealAcc.String())
	}

	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only SP's seal address is allowed to SealObject")
	}

	objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
	if !found {
		return types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_CREATED {
		return types.ErrObjectAlreadySealed
	}

	// check the signature of secondary sps
	// SecondarySP signs the root hash(checksum) of all pieces stored on it, and needs to verify that the signature here.
	var secondarySps []string
	for i, spAddr := range opts.SecondarySpAddresses {
		spAcc := sdk.MustAccAddressFromHex(spAddr)
		secondarySps = append(secondarySps, spAcc.String())
		sr := types.NewSecondarySpSignDoc(spAcc, objectInfo.Id, objectInfo.Checksums[i+1])
		err := k.VerifySPAndSignature(ctx, spAcc, sr.GetSignBytes(), opts.SecondarySpSignatures[i])
		if err != nil {
			return err
		}
	}
	objectInfo.SecondarySpAddresses = secondarySps

	// unlock and charge store fee
	err := k.UnlockAndChargeStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return err
	}

	objectInfo.ObjectStatus = types.OBJECT_STATUS_SEALED

	store := ctx.KVStore(k.storeKey)
	bbz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(objectInfo)
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventSealObject{
		Operator:             spSealAcc.String(),
		BucketName:           bucketInfo.BucketName,
		ObjectName:           objectInfo.ObjectName,
		ObjectId:             objectInfo.Id,
		Status:               objectInfo.ObjectStatus,
		SecondarySpAddresses: objectInfo.SecondarySpAddresses,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CancelCreateObject(
	ctx sdk.Context, operator sdk.AccAddress,
	bucketName, objectName string, opts CancelCreateObjectOptions) error {
	store := ctx.KVStore(k.storeKey)
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
	if !found {
		return types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_CREATED {
		return types.ErrObjectNotCreated.Wrapf("Object status: %s", objectInfo.ObjectStatus.String())
	}

	if objectInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	var creator sdk.AccAddress
	// We use the last address in SecondarySpAddresses to store the creator so that it can be identified when canceling create
	// if the operator is not the creator, we should return access deny
	if len(objectInfo.SecondarySpAddresses) >= 1 {
		creator = sdk.MustAccAddressFromHex(objectInfo.SecondarySpAddresses[len(objectInfo.SecondarySpAddresses)-1])
	}
	// By default, the creator is the owner
	owner := sdk.MustAccAddressFromHex(objectInfo.Owner)
	if !operator.Equals(creator) && !operator.Equals(owner) {
		return errors.Wrapf(types.ErrAccessDenied, "Only allowed owner/creator to do cancel create object")
	}

	err := k.UnlockStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return err
	}

	bbz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketName, objectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCancelCreateObject{
		Operator:         operator.String(),
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
		ObjectId:         objectInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) DeleteObject(
	ctx sdk.Context, operator sdk.AccAddress, bucketName, objectName string, opts DeleteObjectOptions) error {

	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}

	objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
	if !found {
		return types.ErrNoSuchObject
	}

	if objectInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_SEALED &&
		objectInfo.ObjectStatus != types.OBJECT_STATUS_DISCONTINUED {
		return types.ErrObjectNotSealed
	}

	// check permission
	effect := k.VerifyObjectPermission(ctx, bucketInfo, objectInfo, operator, permtypes.ACTION_DELETE_OBJECT)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf(
			"The operator(%s) has no DeleteObject permission of the bucket(%s), object(%s)",
			operator.String(), bucketName, objectName)
	}

	err := k.ChargeDeleteObject(ctx, bucketInfo, objectInfo)
	if err != nil {
		return err
	}

	err = k.doDeleteObject(ctx, operator, bucketInfo, objectInfo)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) doDeleteObject(ctx sdk.Context, operator sdk.AccAddress, bucketInfo *types.BucketInfo, objectInfo *types.ObjectInfo) error {
	store := ctx.KVStore(k.storeKey)

	bbz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketInfo.BucketName, objectInfo.ObjectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := k.appendResourceIdForGarbageCollection(ctx, resource.RESOURCE_TYPE_OBJECT, objectInfo.Id); err != nil {
		return err
	}

	err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteObject{
		Operator:             operator.String(),
		BucketName:           bucketInfo.BucketName,
		ObjectName:           objectInfo.ObjectName,
		ObjectId:             objectInfo.Id,
		PrimarySpAddress:     bucketInfo.PrimarySpAddress,
		SecondarySpAddresses: objectInfo.SecondarySpAddresses,
	})
	return err
}

// ForceDeleteObject will delete object without permission check, it is used for discontinue request from sps.
func (k Keeper) ForceDeleteObject(ctx sdk.Context, objectId sdkmath.Uint) error {
	objectInfo, found := k.GetObjectInfoById(ctx, objectId)
	if !found { // the object is deleted already
		return nil
	}

	bucketInfo, found := k.GetBucketInfo(ctx, objectInfo.BucketName)
	if !found {
		return types.ErrNoSuchBucket
	}

	objectStatus, err := k.getDiscontinueObjectStatus(ctx, objectId)
	if err != nil {
		return err
	}

	if objectStatus == types.OBJECT_STATUS_CREATED {
		err := k.UnlockStoreFee(ctx, bucketInfo, objectInfo)
		if err != nil {
			return err
		}
	} else if objectStatus == types.OBJECT_STATUS_SEALED {
		err := k.ChargeDeleteObject(ctx, bucketInfo, objectInfo)
		if err != nil {
			ctx.Logger().Error("ChargeDeleteObject error", "err", err)
			return err
		}
	}

	sp := sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)
	err = k.doDeleteObject(ctx, sp, bucketInfo, objectInfo)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) CopyObject(
	ctx sdk.Context, operator sdk.AccAddress, srcBucketName, srcObjectName, dstBucketName, dstObjectName string,
	opts CopyObjectOptions) (sdkmath.Uint, error) {

	store := ctx.KVStore(k.storeKey)

	srcBucketInfo, found := k.GetBucketInfo(ctx, srcBucketName)
	if !found {
		return sdkmath.ZeroUint(), errors.Wrapf(types.ErrNoSuchBucket, "src bucket name (%s)", srcBucketName)
	}

	dstBucketInfo, found := k.GetBucketInfo(ctx, dstBucketName)
	if !found {
		return sdkmath.ZeroUint(), errors.Wrapf(types.ErrNoSuchBucket, "dst bucket name (%s)", dstBucketName)
	}

	srcObjectInfo, found := k.GetObjectInfo(ctx, srcBucketName, srcObjectName)
	if !found {
		return sdkmath.ZeroUint(), errors.Wrapf(types.ErrNoSuchObject, "src object name (%s)", srcObjectName)
	}

	if srcObjectInfo.SourceType != opts.SourceType {
		return sdkmath.ZeroUint(), types.ErrSourceTypeMismatch
	}

	// check permission
	effect := k.VerifyObjectPermission(ctx, srcBucketInfo, srcObjectInfo, operator, permtypes.ACTION_COPY_OBJECT)
	if effect != permtypes.EFFECT_ALLOW {
		return sdkmath.ZeroUint(), types.ErrAccessDenied.Wrapf("The operator("+
			"%s) has no CopyObject permission of the bucket(%s), object(%s)",
			operator.String(), srcObjectInfo.BucketName, srcObjectInfo.ObjectName)
	}

	if opts.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return sdkmath.ZeroUint(), errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}

	err := k.VerifySPAndSignature(ctx, sdk.MustAccAddressFromHex(dstBucketInfo.PrimarySpAddress),
		opts.ApprovalMsgBytes,
		opts.PrimarySpApproval.Sig)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	// check payload size, the empty object doesn't need sealed
	var objectStatus types.ObjectStatus
	if srcObjectInfo.PayloadSize == 0 {
		// empty object does not interact with sp
		objectStatus = types.OBJECT_STATUS_SEALED
	} else {
		objectStatus = types.OBJECT_STATUS_CREATED
	}

	objectInfo := types.ObjectInfo{
		Owner:          operator.String(),
		BucketName:     dstBucketInfo.BucketName,
		ObjectName:     dstObjectName,
		PayloadSize:    srcObjectInfo.PayloadSize,
		Visibility:     opts.Visibility,
		ContentType:    srcObjectInfo.ContentType,
		CreateAt:       ctx.BlockTime().Unix(),
		Id:             k.GenNextObjectID(ctx),
		ObjectStatus:   objectStatus,
		RedundancyType: srcObjectInfo.RedundancyType,
		SourceType:     opts.SourceType,
		Checksums:      srcObjectInfo.Checksums,
	}

	if srcObjectInfo.PayloadSize == 0 {
		err = k.ChargeStoreFee(ctx, dstBucketInfo, &objectInfo)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
	} else {
		err = k.LockStoreFee(ctx, dstBucketInfo, &objectInfo)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
	}

	bbz := k.cdc.MustMarshal(dstBucketInfo)
	store.Set(types.GetBucketByIDKey(dstBucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(types.GetObjectKey(dstBucketName, dstObjectName), sequence.EncodeSequence(objectInfo.Id))
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCopyObject{
		Operator:      operator.String(),
		SrcBucketName: srcObjectInfo.BucketName,
		SrcObjectName: srcObjectInfo.ObjectName,
		DstBucketName: objectInfo.BucketName,
		DstObjectName: objectInfo.ObjectName,
		SrcObjectId:   srcObjectInfo.Id,
		DstObjectId:   objectInfo.Id,
	}); err != nil {
		return sdkmath.ZeroUint(), err
	}
	return objectInfo.Id, nil
}

func (k Keeper) RejectSealObject(ctx sdk.Context, operator sdk.AccAddress, bucketName, objectName string) error {
	store := ctx.KVStore(k.storeKey)
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
	if !found {
		return types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_CREATED {
		return types.ErrObjectNotCreated.Wrapf("Object status: %s", objectInfo.ObjectStatus.String())
	}

	sp, found := k.spKeeper.GetStorageProviderBySealAddr(ctx, operator)
	if found {
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "SP seal address: %s", operator.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}
	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only allowed primary SP to do cancel create object")
	}

	err := k.UnlockStoreFee(ctx, bucketInfo, objectInfo)
	if err != nil {
		return err
	}

	bbz := k.cdc.MustMarshal(bucketInfo)

	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketName, objectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventRejectSealObject{
		Operator:   operator.String(),
		BucketName: bucketInfo.BucketName,
		ObjectName: objectInfo.ObjectName,
		ObjectId:   objectInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) DiscontinueObject(ctx sdk.Context, operator sdk.AccAddress, bucketName string, objectIds []sdkmath.Uint, reason string) error {
	sp, found := k.spKeeper.GetStorageProviderByGcAddr(ctx, operator)
	if !found {
		return types.ErrNoSuchStorageProvider.Wrapf("SP operator address: %s", operator.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}

	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	if bucketInfo.BucketStatus == types.BUCKET_STATUS_DISCONTINUED {
		return types.ErrInvalidBucketStatus
	}

	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "only primary sp is allowed to do discontinue objects")
	}

	count := k.getDiscontinueObjectCount(ctx, operator)
	max := k.DiscontinueObjectMax(ctx)
	if count+uint64(len(objectIds)) > max {
		return types.ErrNoMoreDiscontinue.Wrapf("only %d objects can be requested in this window", max-count)
	}

	store := ctx.KVStore(k.storeKey)
	for _, objectId := range objectIds {
		object, found := k.GetObjectInfoById(ctx, objectId)
		if !found {
			return types.ErrInvalidObjectIds.Wrapf("object not found, id: %s", objectId)
		}
		if object.BucketName != bucketName {
			return types.ErrInvalidObjectIds.Wrapf("object %s should in bucket: %s", objectId, bucketName)
		}
		if object.ObjectStatus != types.OBJECT_STATUS_SEALED && object.ObjectStatus != types.OBJECT_STATUS_CREATED {
			return types.ErrInvalidObjectIds.Wrapf("object %s should in created or sealed status", objectId)
		}

		// remember object status
		k.saveDiscontinueObjectStatus(ctx, object)

		// update object status
		object.ObjectStatus = types.OBJECT_STATUS_DISCONTINUED
		store.Set(types.GetObjectByIDKey(object.Id), k.cdc.MustMarshal(object))
	}

	deleteAt := ctx.BlockTime().Unix() + k.DiscontinueConfirmPeriod(ctx)
	k.AppendDiscontinueObjectIds(ctx, deleteAt, objectIds)
	k.setDiscontinueObjectCount(ctx, operator, count+uint64(len(objectIds)))

	events := make([]proto.Message, 0)
	for _, objectId := range objectIds {
		events = append(events, &types.EventDiscontinueObject{
			BucketName: bucketName,
			ObjectId:   objectId,
			Reason:     reason,
			DeleteAt:   deleteAt,
		})
	}
	if err := ctx.EventManager().EmitTypedEvents(events...); err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateObjectInfo(ctx sdk.Context, operator sdk.AccAddress, bucketName, objectName string, visibility types.VisibilityType) error {
	store := ctx.KVStore(k.storeKey)

	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}

	objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
	if !found {
		return types.ErrNoSuchObject
	}

	// check permission
	effect := k.VerifyObjectPermission(ctx, bucketInfo, objectInfo, operator, permtypes.ACTION_UPDATE_OBJECT_INFO)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf("The operator(%s) has no UpdateObjectInfo permission of the bucket(%s), object(%s)",
			operator.String(), bucketName, objectName)
	}

	objectInfo.Visibility = visibility

	obz := k.cdc.MustMarshal(objectInfo)
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateObjectInfo{
		Operator:   operator.String(),
		BucketName: bucketName,
		ObjectName: objectName,
		Visibility: visibility,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CreateGroup(
	ctx sdk.Context, owner sdk.AccAddress,
	groupName string, opts CreateGroupOptions) (sdkmath.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	groupInfo := types.GroupInfo{
		Owner:      owner.String(),
		SourceType: opts.SourceType,
		Id:         k.GenNextGroupId(ctx),
		GroupName:  groupName,
		Extra:      opts.Extra,
	}

	// Can not create a group with the same name.
	groupKey := types.GetGroupKey(owner, groupName)
	if store.Has(groupKey) {
		return sdkmath.ZeroUint(), types.ErrGroupAlreadyExists
	}

	gbz := k.cdc.MustMarshal(&groupInfo)
	store.Set(groupKey, sequence.EncodeSequence(groupInfo.Id))
	store.Set(types.GetGroupByIDKey(groupInfo.Id), gbz)

	// need to limit the size of Msg.Members to avoid taking too long to execute the msg
	for _, member := range opts.Members {
		memberAddress, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return sdkmath.ZeroUint(), err
		}
		err = k.permKeeper.AddGroupMember(ctx, groupInfo.Id, memberAddress)
		if err != nil {
			return sdkmath.Uint{}, err
		}
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateGroup{
		Owner:      groupInfo.Owner,
		GroupName:  groupInfo.GroupName,
		GroupId:    groupInfo.Id,
		SourceType: groupInfo.SourceType,
		Members:    opts.Members,
	}); err != nil {
		return sdkmath.ZeroUint(), err
	}
	return groupInfo.Id, nil
}

func (k Keeper) SetGroupInfo(ctx sdk.Context, groupInfo *types.GroupInfo) {
	store := ctx.KVStore(k.storeKey)

	gbz := k.cdc.MustMarshal(groupInfo)
	store.Set(types.GetGroupByIDKey(groupInfo.Id), gbz)
}

func (k Keeper) GetGroupInfo(ctx sdk.Context, ownerAddr sdk.AccAddress,
	groupName string) (*types.GroupInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGroupKey(ownerAddr, groupName))
	if bz == nil {
		return nil, false
	}

	return k.GetGroupInfoById(ctx, sequence.DecodeSequence(bz))
}

func (k Keeper) GetGroupInfoById(ctx sdk.Context, groupId sdkmath.Uint) (*types.GroupInfo, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGroupByIDKey(groupId))
	if bz == nil {
		return nil, false
	}

	var groupInfo types.GroupInfo
	k.cdc.MustUnmarshal(bz, &groupInfo)
	return &groupInfo, true
}

type DeleteGroupOptions struct {
	SourceType types.SourceType
}

func (k Keeper) DeleteGroup(ctx sdk.Context, operator sdk.AccAddress, groupName string, opts DeleteGroupOptions) error {
	store := ctx.KVStore(k.storeKey)

	groupInfo, found := k.GetGroupInfo(ctx, operator, groupName)
	if !found {
		return types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}
	// check permission
	effect := k.VerifyGroupPermission(ctx, groupInfo, operator, permtypes.ACTION_DELETE_GROUP)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf(
			"The operator(%s) has no DeleteGroup permission of the group(%s), owner(%s)",
			operator.String(), groupInfo.GroupName, groupInfo.Owner)
	}
	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	store.Delete(types.GetGroupKey(operator, groupName))
	store.Delete(types.GetGroupByIDKey(groupInfo.Id))

	if err := k.appendResourceIdForGarbageCollection(ctx, resource.RESOURCE_TYPE_GROUP, groupInfo.Id); err != nil {
		return err
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGroup{
		Owner:     groupInfo.Owner,
		GroupName: groupInfo.GroupName,
		GroupId:   groupInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) LeaveGroup(
	ctx sdk.Context, member sdk.AccAddress, owner sdk.AccAddress,
	groupName string, opts LeaveGroupOptions) error {

	groupInfo, found := k.GetGroupInfo(ctx, owner, groupName)
	if !found {
		return types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	err := k.permKeeper.RemoveGroupMember(ctx, groupInfo.Id, member)
	if err != nil {
		return err
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGroup{
		Owner:     groupInfo.Owner,
		GroupName: groupInfo.GroupName,
		GroupId:   groupInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateGroupMember(ctx sdk.Context, operator sdk.AccAddress, groupInfo *types.GroupInfo, opts UpdateGroupMemberOptions) error {
	if groupInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	// check permission
	effect := k.VerifyGroupPermission(ctx, groupInfo, operator, permtypes.ACTION_UPDATE_GROUP_MEMBER)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf(
			"The operator(%s) has no UpdateGroupMember permission of the group(%s), operator(%s)",
			operator.String(), groupInfo.GroupName, groupInfo.Owner)
	}

	for _, member := range opts.MembersToAdd {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return err
		}
		err = k.permKeeper.AddGroupMember(ctx, groupInfo.Id, memberAcc)
		if err != nil {
			return err
		}
	}

	for _, member := range opts.MembersToDelete {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return err
		}
		err = k.permKeeper.RemoveGroupMember(ctx, groupInfo.Id, memberAcc)
		if err != nil {
			return err
		}

	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGroupMember{
		Operator:        operator.String(),
		Owner:           groupInfo.Owner,
		GroupName:       groupInfo.GroupName,
		GroupId:         groupInfo.Id,
		MembersToAdd:    opts.MembersToAdd,
		MembersToDelete: opts.MembersToDelete,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateGroupExtra(ctx sdk.Context, operator sdk.AccAddress, groupInfo *types.GroupInfo, extra string) error {

	// check permission
	effect := k.VerifyGroupPermission(ctx, groupInfo, operator, permtypes.ACTION_UPDATE_GROUP_EXTRA)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf(
			"The operator(%s) has no UpdateGroupExtra permission of the group(%s), operator(%s)",
			operator.String(), groupInfo.GroupName, groupInfo.Owner)
	}

	if extra != groupInfo.Extra {
		groupInfo.Extra = extra
		obz := k.cdc.MustMarshal(groupInfo)
		ctx.KVStore(k.storeKey).Set(types.GetGroupByIDKey(groupInfo.Id), obz)
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGroupExtra{
		Operator:  operator.String(),
		Owner:     groupInfo.Owner,
		GroupName: groupInfo.GroupName,
		GroupId:   groupInfo.Id,
		Extra:     extra,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) VerifySPAndSignature(ctx sdk.Context, spAcc sdk.AccAddress, sigData []byte, signature []byte) error {
	sp, found := k.spKeeper.GetStorageProvider(ctx, spAcc)
	if !found {
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "SP operator address: %s", spAcc.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}

	approvalAccAddress := sdk.MustAccAddressFromHex(sp.ApprovalAddress)

	err := types.VerifySignature(approvalAccAddress, sdk.Keccak256(sigData), signature)
	if err != nil {
		return errors.Wrapf(types.ErrInvalidApproval, "verify signature error: %s", err)
	}
	return nil
}

func (k Keeper) GenNextBucketId(ctx sdk.Context) sdkmath.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.bucketSeq.NextVal(store)
	return seq
}

func (k Keeper) GenNextObjectID(ctx sdk.Context) sdkmath.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.objectSeq.NextVal(store)
	return seq
}

func (k Keeper) GenNextGroupId(ctx sdk.Context) sdkmath.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.groupSeq.NextVal(store)
	return seq
}

func (k Keeper) GenNextExecutionTaskId(ctx sdk.Context) sdkmath.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.executionTaskSeq.NextVal(store)
	return seq
}

func (k Keeper) isNonEmptyBucket(ctx sdk.Context, bucketName string) bool {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(bucketName))

	iter := objectStore.Iterator(nil, nil)
	return iter.Valid()
}

func (k Keeper) getDiscontinueObjectCount(ctx sdk.Context, operator sdk.AccAddress) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DiscontinueObjectCountPrefix)
	bz := store.Get(operator.Bytes())

	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) setDiscontinueObjectCount(ctx sdk.Context, operator sdk.AccAddress, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DiscontinueObjectCountPrefix)

	countBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(countBytes, count)

	store.Set(operator.Bytes(), countBytes)
}

func (k Keeper) ClearDiscontinueObjectCount(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DiscontinueObjectCountPrefix)

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) AppendDiscontinueObjectIds(ctx sdk.Context, timestamp int64, objectIds []types.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDiscontinueObjectIdsKey(timestamp)
	bz := store.Get(key)
	if bz != nil {
		var existedIds types.Ids
		k.cdc.MustUnmarshal(bz, &existedIds)
		objectIds = append(existedIds.Id, objectIds...)
	}

	store.Set(key, k.cdc.MustMarshal(&types.Ids{Id: objectIds}))
}

func (k Keeper) DeleteDiscontinueObjectsUntil(ctx sdk.Context, timestamp int64, maxObjectsToDelete uint64) (deleted uint64, err error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDiscontinueObjectIdsKey(timestamp)
	iterator := store.Iterator(types.DiscontinueObjectIdsPrefix, storetypes.InclusiveEndBytes(key))
	defer iterator.Close()

	deleted = uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		if deleted >= maxObjectsToDelete {
			break
		}
		var ids types.Ids
		k.cdc.MustUnmarshal(iterator.Value(), &ids)

		left := make([]types.Uint, 0)
		for _, id := range ids.Id {
			if deleted >= maxObjectsToDelete {
				left = append(left, id)
				continue
			}

			err = k.ForceDeleteObject(ctx, id)
			if err != nil {
				ctx.Logger().Error("delete object error", "err", err, "height", ctx.BlockHeight())
				return deleted, err
			}
			deleted++
		}
		if len(left) > 0 {
			store.Set(iterator.Key(), k.cdc.MustMarshal(&types.Ids{Id: left}))
		} else {
			store.Delete(iterator.Key())
		}
	}

	return deleted, nil
}

func (k Keeper) getDiscontinueBucketCount(ctx sdk.Context, operator sdk.AccAddress) uint64 {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DiscontinueBucketCountPrefix)
	bz := store.Get(operator.Bytes())

	if bz == nil {
		return 0
	}
	return binary.BigEndian.Uint64(bz)
}

func (k Keeper) setDiscontinueBucketCount(ctx sdk.Context, operator sdk.AccAddress, count uint64) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DiscontinueBucketCountPrefix)

	countBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(countBytes, count)

	store.Set(operator.Bytes(), countBytes)
}

func (k Keeper) ClearDiscontinueBucketCount(ctx sdk.Context) {
	store := prefix.NewStore(ctx.KVStore(k.storeKey), types.DiscontinueBucketCountPrefix)

	iterator := storetypes.KVStorePrefixIterator(store, []byte{})
	defer iterator.Close()

	for ; iterator.Valid(); iterator.Next() {
		store.Delete(iterator.Key())
	}
}

func (k Keeper) appendDiscontinueBucketIds(ctx sdk.Context, timestamp int64, bucketIds []types.Uint) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDiscontinueBucketIdsKey(timestamp)

	bz := store.Get(key)
	if bz != nil {
		var existedIds types.Ids
		k.cdc.MustUnmarshal(bz, &existedIds)
		bucketIds = append(existedIds.Id, bucketIds...)
	}

	store.Set(key, k.cdc.MustMarshal(&types.Ids{Id: bucketIds}))
}

func (k Keeper) DeleteDiscontinueBucketsUntil(ctx sdk.Context, timestamp int64, maxObjectsToDelete uint64) (uint64, error) {
	store := ctx.KVStore(k.storeKey)
	key := types.GetDiscontinueBucketIdsKey(timestamp)
	iterator := store.Iterator(types.DiscontinueBucketIdsPrefix, storetypes.InclusiveEndBytes(key))
	defer iterator.Close()

	deleted := uint64(0)
	for ; iterator.Valid(); iterator.Next() {
		if deleted >= maxObjectsToDelete {
			break
		}
		var ids types.Ids
		k.cdc.MustUnmarshal(iterator.Value(), &ids)

		left := make([]types.Uint, 0)
		for _, id := range ids.Id {
			if deleted >= maxObjectsToDelete {
				left = append(left, id)
				continue
			}

			bucketDeleted, objectDeleted, err := k.ForceDeleteBucket(ctx, id, maxObjectsToDelete-deleted)
			if err != nil {
				ctx.Logger().Error("force delete bucket error", "err", err)
				return deleted, err
			}
			deleted = deleted + objectDeleted

			if !bucketDeleted {
				left = append(left, id)
			}
		}
		if len(left) > 0 {
			store.Set(iterator.Key(), k.cdc.MustMarshal(&types.Ids{Id: left}))
		} else {
			store.Delete(iterator.Key())
		}
	}

	return deleted, nil
}

func (k Keeper) saveDiscontinueObjectStatus(ctx sdk.Context, object *types.ObjectInfo) {
	store := ctx.KVStore(k.storeKey)
	bz := make([]byte, 4)
	binary.BigEndian.PutUint32(bz, uint32(object.ObjectStatus))
	store.Set(types.GetDiscontinueObjectStatusKey(object.Id), bz)
}

func (k Keeper) getDiscontinueObjectStatus(ctx sdk.Context, objectId types.Uint) (types.ObjectStatus, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetDiscontinueObjectStatusKey(objectId))
	if bz == nil {
		return types.OBJECT_STATUS_DISCONTINUED, errors.Wrapf(types.ErrInvalidObjectStatus, "object id: %s", objectId)
	}
	status := int32(binary.BigEndian.Uint32(bz))
	store.Delete(types.GetDiscontinueObjectStatusKey(objectId)) //remove it at the same time
	return types.ObjectStatus(status), nil
}

func (k Keeper) appendResourceIdForGarbageCollection(ctx sdk.Context, resourceType resource.ResourceType, resourceID sdkmath.Uint) error {
	if !k.permKeeper.ExistAccountPolicyForResource(ctx, resourceType, resourceID) &&
		!k.permKeeper.ExistGroupPolicyForResource(ctx, resourceType, resourceID) {

		if resourceType != resource.RESOURCE_TYPE_GROUP ||
			(resourceType == resource.RESOURCE_TYPE_GROUP && !k.permKeeper.ExistGroupMemberForGroup(ctx, resourceID)) {
			return nil
		}
	}
	tStore := ctx.TransientStore(k.tStoreKey)
	var deleteInfo types.DeleteInfo
	if !tStore.Has(types.CurrentBlockDeleteStalePoliciesKey) {
		deleteInfo = types.DeleteInfo{
			BucketIds: &types.Ids{},
			ObjectIds: &types.Ids{},
			GroupIds:  &types.Ids{},
		}
	} else {
		bz := tStore.Get(types.CurrentBlockDeleteStalePoliciesKey)
		k.cdc.MustUnmarshal(bz, &deleteInfo)
	}
	switch resourceType {
	case resource.RESOURCE_TYPE_BUCKET:
		bucketIds := deleteInfo.BucketIds.Id
		bucketIds = append(bucketIds, resourceID)
		deleteInfo.BucketIds = &types.Ids{Id: bucketIds}
	case resource.RESOURCE_TYPE_OBJECT:
		objectIds := deleteInfo.ObjectIds.Id
		objectIds = append(objectIds, resourceID)
		deleteInfo.ObjectIds = &types.Ids{Id: objectIds}
	case resource.RESOURCE_TYPE_GROUP:
		groupIds := deleteInfo.GroupIds.Id
		groupIds = append(groupIds, resourceID)
		deleteInfo.GroupIds = &types.Ids{Id: groupIds}
	default:
		return types.ErrInvalidResource
	}
	tStore.Set(types.CurrentBlockDeleteStalePoliciesKey, k.cdc.MustMarshal(&deleteInfo))
	return nil
}

func (k Keeper) PersistDeleteInfo(ctx sdk.Context) {
	tStore := ctx.TransientStore(k.tStoreKey)
	if !tStore.Has(types.CurrentBlockDeleteStalePoliciesKey) {
		return
	}
	bz := tStore.Get(types.CurrentBlockDeleteStalePoliciesKey)
	deleteInfo := &types.DeleteInfo{}
	k.cdc.MustUnmarshal(bz, deleteInfo)

	// persist current block stale permission info to store if exists
	if !deleteInfo.IsEmpty() {
		store := ctx.KVStore(k.storeKey)
		store.Set(types.GetDeleteStalePoliciesKey(ctx.BlockHeight()), bz)
		_ = ctx.EventManager().EmitTypedEvents(&types.EventStalePolicyCleanup{
			BlockNum:   ctx.BlockHeight(),
			DeleteInfo: deleteInfo,
		})
	}
}

func (k Keeper) GarbageCollectResourcesStalePolicy(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)
	deleteStalePoliciesPrefixStore := prefix.NewStore(store, types.DeleteStalePoliciesPrefix)

	iterator := deleteStalePoliciesPrefixStore.Iterator(nil, nil)
	defer iterator.Close()

	maxCleanup := k.StalePolicyCleanupMax(ctx)

	var deletedTotal uint64
	var done bool

	for ; iterator.Valid(); iterator.Next() {
		deleteInfo := &types.DeleteInfo{}
		k.cdc.MustUnmarshal(iterator.Value(), deleteInfo)
		deletedTotal, done = k.garbageCollectionForResource(ctx, deleteStalePoliciesPrefixStore, iterator, deleteInfo, resource.RESOURCE_TYPE_OBJECT, deleteInfo.ObjectIds, maxCleanup, deletedTotal)
		if !done {
			return
		}
		deleteInfo.ObjectIds = nil
		deletedTotal, done = k.garbageCollectionForResource(ctx, deleteStalePoliciesPrefixStore, iterator, deleteInfo, resource.RESOURCE_TYPE_BUCKET, deleteInfo.BucketIds, maxCleanup, deletedTotal)
		if !done {
			return
		}
		deleteInfo.BucketIds = nil
		deletedTotal, done = k.garbageCollectionForResource(ctx, deleteStalePoliciesPrefixStore, iterator, deleteInfo, resource.RESOURCE_TYPE_GROUP, deleteInfo.GroupIds, maxCleanup, deletedTotal)
		if !done {
			return
		}
		deleteInfo.GroupIds = nil
		// the specified block height(iterator-key)'s stale resource permission metadata is purged
		if deleteInfo.IsEmpty() {
			deleteStalePoliciesPrefixStore.Delete(iterator.Key())
		}
	}
}

func (k Keeper) garbageCollectionForResource(ctx sdk.Context, deleteStalePoliciesPrefixStore prefix.Store, iterator storetypes.Iterator,
	deleteInfo *types.DeleteInfo, resourceType resource.ResourceType, resourceIds *types.Ids, maxCleanup, deletedTotal uint64) (uint64, bool) {
	var done bool
	if resourceIds != nil && len(resourceIds.Id) > 0 {
		ids := resourceIds.Id
		temp := ids
		for idx, id := range ids {
			deletedTotal, done = k.permKeeper.ForceDeleteAccountPolicyForResource(ctx, maxCleanup, deletedTotal, resourceType, id)
			if !done {
				resourceIds.Id = temp

				deleteStalePoliciesPrefixStore.Set(iterator.Key(), k.cdc.MustMarshal(deleteInfo))
				return deletedTotal, false
			}
			if resourceType == resource.RESOURCE_TYPE_GROUP {
				deletedTotal, done = k.permKeeper.ForceDeleteGroupMembers(ctx, maxCleanup, deletedTotal, id)
				if !done {
					deleteInfo.GroupIds.Id = temp
					deleteStalePoliciesPrefixStore.Set(iterator.Key(), k.cdc.MustMarshal(deleteInfo))
					return deletedTotal, false
				}
				// no need to deal with group policy when resource type is group
				continue
			}
			deletedTotal, done = k.permKeeper.ForceDeleteGroupPolicyForResource(ctx, maxCleanup, deletedTotal, resourceType, id)
			if !done {
				resourceIds.Id = temp
				deleteStalePoliciesPrefixStore.Set(iterator.Key(), k.cdc.MustMarshal(deleteInfo))
				return deletedTotal, false
			}
			//  remove current resource id from list of ids to be deleted
			temp = ids[idx+1:]
		}
	}
	return deletedTotal, true
}

func (k Keeper) CheckInvokePermissions(ctx sdk.Context, executableId sdkmath.Uint, inputIds []sdkmath.Uint) error {
	_, exist := k.GetObjectInfoById(ctx, executableId)
	if !exist {
		return types.ErrNoSuchObject.Wrapf("executable id %d", executableId)
	}

	for _, inputId := range inputIds {
		_, exist := k.GetObjectInfoById(ctx, inputId)
		if !exist {
			return types.ErrNoSuchObject.Wrapf("input id %d", inputId)
		}
	}

	// todo: check further persmissions if needed
	return nil
}

func (k Keeper) InvokeExecution(ctx sdk.Context, operator sdk.AccAddress, executableObjectId sdkmath.Uint, ops InvokeExecutionOptions) error {
	taskId := k.GenNextExecutionTaskId(ctx)

	_ = ctx.EventManager().EmitTypedEvents(&types.EventExecutionTask{
		TaskId:             taskId,
		Operator:           operator.String(),
		ExecutableObjectId: executableObjectId,
		InputObjectIds:     ops.InputObjectIds,
		MaxGas:             ops.MaxGas,
		Method:             ops.Method,
		Params:             ops.Params,
	})

	return nil
}

func (k Keeper) SubmitExecutionResult(ctx sdk.Context, operator sdk.AccAddress, taskId sdkmath.Uint, status uint32, dataUri string) error {
	if taskId.Equal(sdkmath.ZeroUint()) || taskId.GT(k.executionTaskSeq.CurVal(ctx.KVStore(k.storeKey))) {
		return types.ErrInvalidTaskId
	}
	executionResult, exist := k.GetExecutionResult(ctx, taskId)
	if exist {
		return types.ErrExecutionResultSubmitted
	}

	executionResult = &types.ExecutionResult{
		Status: status,
	}
	k.SetExecutionResult(ctx, taskId, executionResult)

	_ = ctx.EventManager().EmitTypedEvents(&types.EventExecutionResult{
		TaskId:        taskId,
		Operator:      operator.String(),
		Status:        status,
		ResultDataUri: dataUri,
	})

	return nil
}

func (k Keeper) GetExecutionResult(ctx sdk.Context, taskId sdkmath.Uint) (*types.ExecutionResult, bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetExecutionResultKey(taskId))
	if bz == nil {
		return nil, false
	}

	var executionResult types.ExecutionResult
	k.cdc.MustUnmarshal(bz, &executionResult)

	return &executionResult, true
}

func (k Keeper) SetExecutionResult(ctx sdk.Context, taskId sdkmath.Uint, executionResult *types.ExecutionResult) {
	store := ctx.KVStore(k.storeKey)
	store.Set(types.GetExecutionResultKey(taskId), k.cdc.MustMarshal(executionResult))
}
