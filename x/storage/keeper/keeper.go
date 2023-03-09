package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
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
		accountKeeper types.AccountKeeper

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
	accountKeeper types.AccountKeeper,
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
		accountKeeper: accountKeeper,
		spKeeper:      spKeeper,
		paymentKeeper: paymentKeeper,
	}

	k.bucketSeq = NewSequence(types.BucketSequencePrefix)
	k.objectSeq = NewSequence(types.ObjectSequencePrefix)
	k.groupSeq = NewSequence(types.GroupSequencePrefix)
	return &k
}

func (k Keeper) Logger(ctx sdk.Context) log.Logger {
	return ctx.Logger().With("module", fmt.Sprintf("x/%s", types.ModuleName))
}

func (k Keeper) CreateBucket(
	ctx sdk.Context, ownerAcc sdk.AccAddress, bucketName string,
	primarySpAcc sdk.AccAddress, opts CreateBucketOptions) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	// check if the bucket exist
	bucketKey := types.GetBucketKey(bucketName)
	if store.Has(bucketKey) {
		return math.ZeroUint(), types.ErrBucketAlreadyExists
	}

	// check payment account
	paymentAcc, err := k.VerifyPaymentAccount(ctx, opts.PaymentAddress, ownerAcc)
	if err != nil {
		return math.ZeroUint(), err
	}

	// check primary sp approval
	if opts.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return math.ZeroUint(), errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}
	err = k.VerifySPAndSignature(ctx, primarySpAcc, opts.ApprovalMsgBytes, opts.PrimarySpApproval.Sig)
	if err != nil {
		return math.ZeroUint(), err
	}

	bucketInfo := types.BucketInfo{
		Owner:            ownerAcc.String(),
		BucketName:       bucketName,
		IsPublic:         opts.IsPublic,
		CreateAt:         ctx.BlockHeight(),
		SourceType:       opts.SourceType,
		ReadQuota:        opts.ReadQuota,
		PaymentAddress:   paymentAcc.String(),
		PrimarySpAddress: primarySpAcc.String(),
	}

	// charge by read quota
	if opts.ReadQuota != types.READ_QUOTA_FREE {
		err := k.paymentKeeper.ChargeInitialReadFee(ctx, &bucketInfo)
		if err != nil {
			return math.ZeroUint(), err
		}
	}

	// Generate bucket Id
	bucketInfo.Id = k.GenNextBucketId(ctx)

	// store the bucket
	bz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(bucketKey, types.EncodeSequence(bucketInfo.Id))
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bz)

	// emit CreateBucket Event
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateBucket{
		OwnerAddress:     bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		IsPublic:         bucketInfo.IsPublic,
		CreateAt:         bucketInfo.CreateAt,
		Id:               bucketInfo.Id,
		SourceType:       bucketInfo.SourceType,
		ReadQuota:        bucketInfo.ReadQuota,
		PaymentAddress:   bucketInfo.PaymentAddress,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
	}); err != nil {
		return math.Uint{}, err
	}
	return bucketInfo.Id, nil
}

func (k Keeper) DeleteBucket(ctx sdk.Context, operator sdk.AccAddress, bucketName string, opts DeleteBucketOptions) error {
	store := ctx.KVStore(k.storeKey)
	bucketKey := types.GetBucketKey(bucketName)

	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return types.ErrNoSuchBucket
	}
	if bucketInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	// check permission
	OwnerAcc := sdk.MustAccAddressFromHex(bucketInfo.Owner)
	if !OwnerAcc.Equals(operator) {
		return types.ErrAccessDenied
	}

	// check if the bucket empty
	if k.isNonEmptyBucket(ctx, bucketName) {
		return types.ErrBucketNotEmpty
	}

	store.Delete(bucketKey)
	store.Delete(types.GetBucketByIDKey(bucketInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteBucket{
		OperatorAddress:  operator.String(),
		OwnerAddress:     bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		Id:               bucketInfo.Id,
		PrimarySpAddress: OwnerAcc.String(),
	}); err != nil {
		return err
	}
	return nil
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
	OwnerAcc := sdk.MustAccAddressFromHex(bucketInfo.Owner)
	if !OwnerAcc.Equals(operator) {
		return types.ErrAccessDenied
	}

	// update charge
	paymentAcc, err := k.VerifyPaymentAccount(ctx, opts.PaymentAddress, OwnerAcc)
	if err != sdk.ErrEmptyHexAddress {
		err := k.paymentKeeper.ChargeUpdatePaymentAccount(ctx, &bucketInfo, &opts.PaymentAddress)
		if err != nil {
			return err
		}
		bucketInfo.PaymentAddress = paymentAcc.String()
	}

	// update quota
	if opts.ReadQuota != bucketInfo.ReadQuota {
		err := k.paymentKeeper.ChargeUpdateReadQuota(ctx, &bucketInfo, opts.ReadQuota)
		if err != nil {
			return err
		}
		bucketInfo.ReadQuota = opts.ReadQuota
	}

	store := ctx.KVStore(k.storeKey)
	bz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateBucketInfo{
		OperatorAddress:      operator.String(),
		BucketName:           bucketName,
		Id:                   bucketInfo.Id,
		ReadQuotaBefore:      bucketInfo.ReadQuota,
		ReadQuotaAfter:       opts.ReadQuota,
		PaymentAddressBefore: bucketInfo.PaymentAddress,
		PaymentAddressAfter:  paymentAcc.String(),
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) GetBucketInfo(ctx sdk.Context, bucketName string) (bucketInfo types.BucketInfo, found bool) {
	store := ctx.KVStore(k.storeKey)

	bucketKey := types.GetBucketKey(bucketName)
	bz := store.Get(bucketKey)
	if bz == nil {
		return bucketInfo, false
	}

	return k.GetBucketInfoById(ctx, types.DecodeSequence(bz))
}

func (k Keeper) GetBucketInfoById(ctx sdk.Context, bucketId math.Uint) (bucketInfo types.BucketInfo, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetBucketByIDKey(bucketId))
	if bz == nil {
		return bucketInfo, false
	}

	k.cdc.MustUnmarshal(bz, &bucketInfo)

	return bucketInfo, true
}

func (k Keeper) CreateObject(
	ctx sdk.Context, ownerAcc sdk.AccAddress, bucketName, objectName string,
	payloadSize uint64, opts CreateObjectOptions) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	// check payload size
	if payloadSize > k.MaxPayloadSize(ctx) {
		return math.ZeroUint(), types.ErrTooLargeObject
	}

	// check bucket
	bucketInfo, found := k.GetBucketInfo(ctx, bucketName)
	if !found {
		return math.ZeroUint(), types.ErrNoSuchBucket
	}

	// check secondary sps
	var secondarySPs []string
	for _, sp := range opts.SecondarySpAddresses {
		spAcc, err := sdk.AccAddressFromHexUnsafe(sp)
		if err != nil {
			return math.ZeroUint(), err
		}
		err = k.spKeeper.IsStorageProviderExistAndInService(ctx, spAcc)
		if err != nil {
			return math.ZeroUint(), err
		}
		secondarySPs = append(secondarySPs, spAcc.String())
	}

	// check approval
	if opts.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return math.ZeroUint(), errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}

	err := k.VerifySPAndSignature(ctx, sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress), opts.ApprovalMsgBytes,
		opts.PrimarySpApproval.Sig)
	if err != nil {
		return math.ZeroUint(), err
	}

	objectKey := types.GetObjectKey(bucketName, objectName)
	if store.Has(objectKey) {
		return math.ZeroUint(), types.ErrObjectAlreadyExists
	}

	// construct objectInfo
	objectInfo := types.ObjectInfo{
		Owner:                ownerAcc.String(),
		BucketName:           bucketName,
		ObjectName:           objectName,
		PayloadSize:          payloadSize,
		IsPublic:             opts.IsPublic,
		ContentType:          opts.ContentType,
		Id:                   k.GenNextObjectID(ctx),
		CreateAt:             ctx.BlockHeight(),
		ObjectStatus:         types.OBJECT_STATUS_CREATED,
		RedundancyType:       opts.RedundancyType, // TODO: base on redundancy policy
		SourceType:           opts.SourceType,
		Checksums:            opts.Checksums,
		SecondarySpAddresses: secondarySPs,
	}

	// Lock Fee
	err = k.paymentKeeper.LockStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return math.ZeroUint(), err
	}

	// TODO(fynn): consider remove the lock fee meta from bucketInfo
	bbz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(objectKey, types.EncodeSequence(objectInfo.Id))
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateObject{
		CreatorAddress:   ownerAcc.String(),
		OwnerAddress:     objectInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		BucketId:         bucketInfo.Id,
		ObjectId:         objectInfo.Id,
		CreateAt:         bucketInfo.CreateAt,
		PayloadSize:      objectInfo.PayloadSize,
		IsPublic:         objectInfo.IsPublic,
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

func (k Keeper) GetObjectInfoCount(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.objectSeq.CurVal(store)
	return seq
}

func (k Keeper) GetObjectInfo(ctx sdk.Context, bucketName string, objectName string) (objectInfo types.ObjectInfo, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetObjectKey(bucketName, objectName))
	if bz == nil {
		return objectInfo, false
	}

	return k.GetObjectInfoById(ctx, types.DecodeSequence(bz))
}

func (k Keeper) GetObjectInfoById(ctx sdk.Context, objectId math.Uint) (objectInfo types.ObjectInfo, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetObjectByIDKey(objectId))
	if bz == nil {
		return objectInfo, false
	}

	k.cdc.MustUnmarshal(bz, &objectInfo)
	return objectInfo, true
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
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "sealAddr: %s, status: %s", spSealAcc.String(), sp.Status.String())
	}

	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only SP's seal address is allowed to SealObject")
	}

	objectInfo, found := k.GetObjectInfo(ctx, bucketName, objectName)
	if !found {
		return types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_CREATED {
		return types.ErrObjectAlreadyExists
	}

	// check the signature of secondary sps
	// SecondarySP signs the root hash(checksum) of all pieces stored on it, and needs to verify that the signature here.
	var secondarySps []string
	for i, spAddr := range opts.SecondarySpAddresses {
		spAcc, err := sdk.AccAddressFromHexUnsafe(spAddr)
		if err != nil {
			return err
		}
		secondarySps = append(secondarySps, spAcc.String())
		sr := types.NewSecondarySpSignDoc(spAcc, objectInfo.Checksums[i+1])
		err = k.VerifySPAndSignature(ctx, spAcc, sr.GetSignBytes(), opts.SecondarySpSignatures[i])
		if err != nil {
			return err
		}
	}

	// unlock fee
	err := k.paymentKeeper.UnlockAndChargeStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return err
	}

	objectInfo.ObjectStatus = types.OBJECT_STATUS_SEALED
	objectInfo.SecondarySpAddresses = secondarySps
	objectInfo.LockedBalance = nil

	// TODO(fynn): consider remove the lock fee meta from bucketInfo
	store := ctx.KVStore(k.storeKey)
	bbz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventSealObject{
		OperatorAddress:    spSealAcc.String(),
		BucketName:         bucketInfo.BucketName,
		ObjectName:         objectInfo.ObjectName,
		Id:                 objectInfo.Id,
		Status:             objectInfo.ObjectStatus,
		SecondarySpAddress: objectInfo.SecondarySpAddresses,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CancelCreateObject(
	ctx sdk.Context, ownAcc sdk.AccAddress,
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
		return types.ErrObjectNotInit
	}

	if objectInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	if !ownAcc.Equals(sdk.MustAccAddressFromHex(objectInfo.Owner)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only allowed owner to do cancel create object")
	}

	err := k.paymentKeeper.UnlockStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return err
	}

	// TODO(fynn): consider remove the lock fee meta from bucketInfo
	bbz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketName, objectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCancelCreateObject{
		OperatorAddress:  ownAcc.String(),
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
		Id:               objectInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) DeleteObject(
	ctx sdk.Context, operator sdk.AccAddress, bucketName, objectName string, opts DeleteObjectOptions) error {
	store := ctx.KVStore(k.storeKey)
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

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_SEALED {
		return types.ErrObjectNotInService
	}

	// Currently, only the owner is allowed to delete object
	if !operator.Equals(sdk.MustAccAddressFromHex(objectInfo.Owner)) {
		return errors.Wrapf(types.ErrAccessDenied, "no permission")
	}

	err := k.paymentKeeper.ChargeDeleteObject(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return err
	}

	// TODO(fynn): consider remove the lock fee meta from bucketInfo
	bbz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketName, objectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteObject{
		OperatorAddress:      operator.String(),
		BucketName:           bucketInfo.BucketName,
		ObjectName:           objectInfo.ObjectName,
		Id:                   objectInfo.Id,
		PrimarySpAddress:     bucketInfo.PrimarySpAddress,
		SecondarySpAddresses: objectInfo.SecondarySpAddresses,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CopyObject(
	ctx sdk.Context, operator sdk.AccAddress, srcBucketName, srcObjectName, dstBucketName, dstObjectName string,
	opts CopyObjectOptions) (math.Uint, error) {

	store := ctx.KVStore(k.storeKey)

	_, found := k.GetBucketInfo(ctx, srcBucketName)
	if !found {
		return math.ZeroUint(), errors.Wrapf(types.ErrNoSuchBucket, "src bucket name (%s)", srcBucketName)
	}

	dstBucketInfo, found := k.GetBucketInfo(ctx, dstBucketName)
	if !found {
		return math.ZeroUint(), errors.Wrapf(types.ErrNoSuchBucket, "dst bucket name (%s)", dstBucketName)
	}

	srcObjectInfo, found := k.GetObjectInfo(ctx, srcBucketName, srcObjectName)
	if !found {
		return math.ZeroUint(), errors.Wrapf(types.ErrNoSuchObject, "src object name (%s)", srcObjectName)
	}

	if srcObjectInfo.SourceType != opts.SourceType {
		return math.ZeroUint(), types.ErrSourceTypeMismatch
	}

	if !operator.Equals(sdk.MustAccAddressFromHex(srcObjectInfo.Owner)) {
		return math.ZeroUint(), errors.Wrapf(types.ErrAccessDenied, "No permission")
	}

	if opts.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return math.ZeroUint(), errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}

	err := k.VerifySPAndSignature(ctx, sdk.MustAccAddressFromHex(dstBucketInfo.PrimarySpAddress),
		opts.ApprovalMsgBytes,
		opts.PrimarySpApproval.Sig)
	if err != nil {
		return math.ZeroUint(), err
	}

	objectInfo := types.ObjectInfo{
		Owner:          operator.String(),
		BucketName:     dstBucketInfo.BucketName,
		ObjectName:     dstObjectName,
		PayloadSize:    srcObjectInfo.PayloadSize,
		IsPublic:       opts.IsPublic,
		ContentType:    srcObjectInfo.ContentType,
		CreateAt:       ctx.BlockHeight(),
		Id:             k.GenNextObjectID(ctx),
		ObjectStatus:   types.OBJECT_STATUS_CREATED,
		RedundancyType: srcObjectInfo.RedundancyType,
		SourceType:     opts.SourceType,
		Checksums:      srcObjectInfo.Checksums,
	}

	err = k.paymentKeeper.LockStoreFee(ctx, &dstBucketInfo, &objectInfo)
	if err != nil {
		return math.ZeroUint(), err
	}

	// TODO(fynn): consider remove the lock fee meta from bucketInfo
	bbz := k.cdc.MustMarshal(&dstBucketInfo)
	store.Set(types.GetBucketByIDKey(dstBucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(types.GetObjectKey(dstBucketName, dstObjectName), types.EncodeSequence(objectInfo.Id))
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCopyObject{
		OperatorAddress: operator.String(),
		SrcBucketName:   srcObjectInfo.BucketName,
		SrcObjectName:   srcObjectInfo.ObjectName,
		DstBucketName:   objectInfo.BucketName,
		DstObjectName:   objectInfo.ObjectName,
		SrcObjectId:     srcObjectInfo.Id,
		DstObjectId:     objectInfo.Id,
	}); err != nil {
		return math.ZeroUint(), err
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
		return types.ErrObjectNotInit
	}

	sp, found := k.spKeeper.GetStorageProviderBySealAddr(ctx, operator)
	if found {
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "sealAddr: %s, status: %s", operator.String(), sp.Status.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}
	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only allowed primary sp to do cancel create object")
	}

	err := k.paymentKeeper.UnlockStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return err
	}

	// TODO(fynn): consider remove the lock fee meta from bucketInfo
	bbz := k.cdc.MustMarshal(&bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketName, objectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventRejectSealObject{
		OperatorAddress: operator.String(),
		BucketName:      bucketInfo.BucketName,
		ObjectName:      objectInfo.ObjectName,
		Id:              objectInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) CreateGroup(
	ctx sdk.Context, owner sdk.AccAddress,
	groupName string, opts CreateGroupOptions) (math.Uint, error) {
	store := ctx.KVStore(k.storeKey)

	groupInfo := types.GroupInfo{
		Owner:      owner.String(),
		SourceType: opts.SourceType,
		Id:         k.GenNextGroupId(ctx),
		GroupName:  groupName,
	}

	gbz := k.cdc.MustMarshal(&groupInfo)
	store.Set(types.GetGroupKey(owner, groupName), types.EncodeSequence(groupInfo.Id))
	store.Set(types.GetGroupByIDKey(groupInfo.Id), gbz)

	// need to limit the size of Msg.Members to avoid taking too long to execute the msg
	for _, member := range opts.Members {
		memberAddress, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return math.ZeroUint(), err
		}
		groupMemberInfo := types.GroupMemberInfo{
			Member:     memberAddress.String(),
			Id:         groupInfo.Id,
			ExpireTime: 0,
		}
		mbz := k.cdc.MustMarshal(&groupMemberInfo)
		store.Set(types.GetGroupMemberKey(groupInfo.Id, memberAddress), mbz)
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateGroup{
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		Id:           groupInfo.Id,
		SourceType:   groupInfo.SourceType,
		Members:      opts.Members,
	}); err != nil {
		return math.ZeroUint(), err
	}
	return groupInfo.Id, nil
}

func (k Keeper) GetGroupInfo(ctx sdk.Context, ownerAddr sdk.AccAddress, groupName string) (groupInfo types.GroupInfo, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGroupKey(ownerAddr, groupName))
	if bz == nil {
		return groupInfo, false
	}

	return k.GetGroupInfoById(ctx, types.DecodeSequence(bz))
}

func (k Keeper) GetGroupInfoById(ctx sdk.Context, groupId math.Uint) (groupInfo types.GroupInfo, found bool) {
	store := ctx.KVStore(k.storeKey)

	bz := store.Get(types.GetGroupByIDKey(groupId))
	if bz == nil {
		return groupInfo, false
	}

	k.cdc.MustUnmarshal(bz, &groupInfo)
	return groupInfo, true
}

type DeleteGroupOptions struct {
	types.SourceType
}

func (k Keeper) DeleteGroup(ctx sdk.Context, ownerAddr sdk.AccAddress, groupName string, opts DeleteGroupOptions) error {
	store := ctx.KVStore(k.storeKey)

	groupInfo, found := k.GetGroupInfo(ctx, ownerAddr, groupName)
	if !found {
		return types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}
	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	store.Delete(types.GetGroupKey(ownerAddr, groupName))
	store.Delete(types.GetGroupByIDKey(groupInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGroup{
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		Id:           groupInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) LeaveGroup(
	ctx sdk.Context, member sdk.AccAddress, owner sdk.AccAddress,
	groupName string, opts LeaveGroupOptions) error {
	store := ctx.KVStore(k.storeKey)

	groupInfo, found := k.GetGroupInfo(ctx, owner, groupName)
	if !found {
		return types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	store.Delete(types.GetGroupMemberKey(groupInfo.Id, member))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGroup{
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		Id:           groupInfo.Id,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) UpdateGroupMember(ctx sdk.Context, owner sdk.AccAddress, groupName string, opts UpdateGroupMemberOptions) error {
	store := ctx.KVStore(k.storeKey)
	groupInfo, found := k.GetGroupInfo(ctx, owner, groupName)
	if !found {
		return types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	for _, member := range opts.MembersToAdd {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return err
		}
		memberInfo := types.GroupMemberInfo{
			Member:     memberAcc.String(),
			Id:         groupInfo.Id,
			ExpireTime: 0,
		}

		mbz := k.cdc.MustMarshal(&memberInfo)
		store.Set(types.GetGroupMemberKey(groupInfo.Id, memberAcc), mbz)
	}

	for _, member := range opts.MembersToDelete {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return err
		}
		store.Delete(types.GetGroupMemberKey(groupInfo.Id, memberAcc))
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGroupMember{
		OperatorAddress: owner.String(),
		OwnerAddress:    groupInfo.Owner,
		GroupName:       groupInfo.GroupName,
		Id:              groupInfo.Id,
		MembersToAdd:    opts.MembersToAdd,
		MembersToDelete: opts.MembersToDelete,
	}); err != nil {
		return err
	}
	return nil
}

func (k Keeper) VerifySPAndSignature(ctx sdk.Context, spAcc sdk.AccAddress, sigData []byte, signature []byte) error {
	sp, found := k.spKeeper.GetStorageProvider(ctx, spAcc)
	if !found {
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "spAddr: %s, status: %s", sp.OperatorAddress, sp.Status.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}

	approvalAccAddress, err := sdk.AccAddressFromHexUnsafe(sp.ApprovalAddress)
	if err != nil {
		return err
	}

	err = types.VerifySignature(approvalAccAddress, sdk.Keccak256(sigData), signature)
	if err != nil {
		return err
	}
	return nil
}

func (k Keeper) GenNextBucketId(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.bucketSeq.NextVal(store)
	return seq
}

func (k Keeper) GenNextObjectID(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.objectSeq.NextVal(store)
	return seq
}

func (k Keeper) GenNextGroupId(ctx sdk.Context) math.Uint {
	store := ctx.KVStore(k.storeKey)

	seq := k.groupSeq.NextVal(store)
	return seq
}

func (k Keeper) isNonEmptyBucket(ctx sdk.Context, bucketName string) bool {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(bucketName))

	iter := objectStore.Iterator(nil, nil)
	return iter.Valid()
}
