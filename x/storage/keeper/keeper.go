package keeper

import (
	"fmt"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/bnb-chain/greenfield/internal/sequence"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type (
	Keeper struct {
		cdc              codec.BinaryCodec
		storeKey         storetypes.StoreKey
		memKey           storetypes.StoreKey
		paramStore       paramtypes.Subspace
		spKeeper         types.SpKeeper
		paymentKeeper    types.PaymentKeeper
		accountKeeper    types.AccountKeeper
		permKeeper       types.PermissionKeeper
		crossChainKeeper types.CrossChainKeeper

		// sequence
		bucketSeq sequence.U256
		objectSeq sequence.U256
		groupSeq  sequence.U256
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
	permKeeper types.PermissionKeeper,
	crossChainKeeper types.CrossChainKeeper,
) *Keeper {
	// set KeyTable if it has not already been set
	if !ps.HasKeyTable() {
		ps = ps.WithKeyTable(types.ParamKeyTable())
	}

	k := Keeper{
		cdc:              cdc,
		storeKey:         storeKey,
		memKey:           memKey,
		paramStore:       ps,
		accountKeeper:    accountKeeper,
		spKeeper:         spKeeper,
		paymentKeeper:    paymentKeeper,
		permKeeper:       permKeeper,
		crossChainKeeper: crossChainKeeper,
	}

	k.bucketSeq = sequence.NewSequence256(types.BucketSequencePrefix)
	k.objectSeq = sequence.NewSequence256(types.ObjectSequencePrefix)
	k.groupSeq = sequence.NewSequence256(types.GroupSequencePrefix)
	return &k
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
		OwnerAddress:     bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		Visibility:       bucketInfo.Visibility,
		CreateAt:         bucketInfo.CreateAt,
		BucketId:         bucketInfo.Id,
		SourceType:       bucketInfo.SourceType,
		ChargedReadQuota: bucketInfo.ChargedReadQuota,
		PaymentAddress:   bucketInfo.PaymentAddress,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
	}); err != nil {
		return sdkmath.Uint{}, err
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

	store.Delete(bucketKey)
	store.Delete(types.GetBucketByIDKey(bucketInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteBucket{
		OperatorAddress:  operator.String(),
		OwnerAddress:     bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		BucketId:         bucketInfo.Id,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
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
	effect := k.VerifyBucketPermission(ctx, bucketInfo, operator, permtypes.ACTION_UPDATE_BUCKET_INFO, nil)
	if effect != permtypes.EFFECT_ALLOW {
		return types.ErrAccessDenied.Wrapf("The operator(%s) has no UpdateBucketInfo permission of the bucket(%s)",
			operator.String(), bucketName)
	}

	// handle fields not changed
	if opts.ChargedReadQuota == nil {
		opts.ChargedReadQuota = &bucketInfo.ChargedReadQuota
	}
	bucketInfo.Visibility = opts.Visibility

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
		OperatorAddress:        operator.String(),
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
		ObjectStatus:         types.OBJECT_STATUS_CREATED,
		RedundancyType:       opts.RedundancyType,
		SourceType:           opts.SourceType,
		Checksums:            opts.Checksums,
		SecondarySpAddresses: secondarySPs,
	}

	// Lock Fee
	err = k.LockStoreFee(ctx, bucketInfo, &objectInfo)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	bbz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(objectKey, sequence.EncodeSequence(objectInfo.Id))
	store.Set(types.GetObjectByIDKey(objectInfo.Id), obz)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateObject{
		CreatorAddress:   operator.String(),
		OwnerAddress:     objectInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		BucketId:         bucketInfo.Id,
		ObjectId:         objectInfo.Id,
		CreateAt:         bucketInfo.CreateAt,
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
		OperatorAddress:      spSealAcc.String(),
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
		return types.ErrObjectNotCreated.Wrapf("Object status: %s", objectInfo.ObjectStatus.String())
	}

	if objectInfo.SourceType != opts.SourceType {
		return types.ErrSourceTypeMismatch
	}

	if !ownAcc.Equals(sdk.MustAccAddressFromHex(objectInfo.Owner)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only allowed owner to do cancel create object")
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
		OperatorAddress:  ownAcc.String(),
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

	bbz := k.cdc.MustMarshal(bucketInfo)
	store.Set(types.GetBucketByIDKey(bucketInfo.Id), bbz)

	store.Delete(types.GetObjectKey(bucketName, objectName))
	store.Delete(types.GetObjectByIDKey(objectInfo.Id))

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteObject{
		OperatorAddress:      operator.String(),
		BucketName:           bucketInfo.BucketName,
		ObjectName:           objectInfo.ObjectName,
		ObjectId:             objectInfo.Id,
		PrimarySpAddress:     bucketInfo.PrimarySpAddress,
		SecondarySpAddresses: objectInfo.SecondarySpAddresses,
	}); err != nil {
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

	objectInfo := types.ObjectInfo{
		Owner:          operator.String(),
		BucketName:     dstBucketInfo.BucketName,
		ObjectName:     dstObjectName,
		PayloadSize:    srcObjectInfo.PayloadSize,
		Visibility:     opts.Visibility,
		ContentType:    srcObjectInfo.ContentType,
		CreateAt:       ctx.BlockHeight(),
		Id:             k.GenNextObjectID(ctx),
		ObjectStatus:   types.OBJECT_STATUS_CREATED,
		RedundancyType: srcObjectInfo.RedundancyType,
		SourceType:     opts.SourceType,
		Checksums:      srcObjectInfo.Checksums,
	}

	err = k.LockStoreFee(ctx, dstBucketInfo, &objectInfo)
	if err != nil {
		return sdkmath.ZeroUint(), err
	}

	bbz := k.cdc.MustMarshal(dstBucketInfo)
	store.Set(types.GetBucketByIDKey(dstBucketInfo.Id), bbz)

	obz := k.cdc.MustMarshal(&objectInfo)
	store.Set(types.GetObjectKey(dstBucketName, dstObjectName), sequence.EncodeSequence(objectInfo.Id))
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
		return errors.Wrapf(types.ErrNoSuchStorageProvider, "sealAddr: %s, status: %s", operator.String(), sp.Status.String())
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return sptypes.ErrStorageProviderNotInService
	}
	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)) {
		return errors.Wrapf(types.ErrAccessDenied, "Only allowed primary sp to do cancel create object")
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
		OperatorAddress: operator.String(),
		BucketName:      bucketInfo.BucketName,
		ObjectName:      objectInfo.ObjectName,
		ObjectId:        objectInfo.Id,
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
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		GroupId:      groupInfo.Id,
		SourceType:   groupInfo.SourceType,
		Members:      opts.Members,
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

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGroup{
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		GroupId:      groupInfo.Id,
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
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		GroupId:      groupInfo.Id,
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
		OperatorAddress: operator.String(),
		OwnerAddress:    groupInfo.Owner,
		GroupName:       groupInfo.GroupName,
		GroupId:         groupInfo.Id,
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

func (k Keeper) isNonEmptyBucket(ctx sdk.Context, bucketName string) bool {
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(bucketName))

	iter := objectStore.Iterator(nil, nil)
	return iter.Valid()
}
