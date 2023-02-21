package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) CreateBucket(goCtx context.Context, msg *types.MsgCreateBucket) (*types.MsgCreateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check the bucket permission
	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	var paymentAcc sdk.AccAddress
	if msg.PaymentAddress != "" {
		paymentAcc, err = sdk.AccAddressFromHexUnsafe(msg.PaymentAddress)
		if err != nil {
			return nil, err
		}
		if !k.paymentKeeper.IsPaymentAccountOwner(ctx, paymentAcc.String(), ownerAcc.String()) {
			return nil, paymenttypes.ErrNotPaymentAccountOwner
		}
	} else {
		paymentAcc = ownerAcc
	}

	primaryAcc, err := sdk.AccAddressFromHexUnsafe(msg.PrimarySpAddress)
	if err != nil {
		return nil, err
	}
	if msg.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return nil, errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}
	err = k.VerifySPAndSignature(ctx, msg.PrimarySpAddress, msg.GetApprovalBytes(), msg.PrimarySpApproval.Sig)
	if err != nil {
		return nil, err
	}

	// Store bucket meta
	bucketInfo := types.BucketInfo{
		Owner:            ownerAcc.String(),
		BucketName:       msg.BucketName,
		IsPublic:         msg.IsPublic,
		CreateAt:         ctx.BlockHeight(),
		Id:               k.GetBucketId(ctx),
		SourceType:       types.SOURCE_TYPE_ORIGIN,
		ReadQuota:        msg.ReadQuota,
		PaymentAddress:   paymentAcc.String(),
		PrimarySpAddress: primaryAcc.String(),
	}

	if msg.ReadQuota != 0 {
		err := k.ChargeInitialReadFee(ctx, &bucketInfo)
		if err != nil {
			return nil, err
		}
	}
	err = k.Keeper.CreateBucket(ctx, bucketInfo)
	if err != nil {
		return nil, err
	}
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
		return nil, err
	}
	return &types.MsgCreateBucketResponse{}, nil
}

func (k msgServer) DeleteBucket(goCtx context.Context, msg *types.MsgDeleteBucket) (*types.MsgDeleteBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	bucketInfo, found := k.Keeper.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	if bucketInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}
	if bucketInfo.Owner != msg.Operator {
		return nil, types.ErrAccessDenied
	}

	err = k.Keeper.DeleteBucket(ctx, msg.BucketName)
	if err != nil {
		return nil, err
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteBucket{
		OperatorAddress:  operatorAcc.String(),
		OwnerAddress:     bucketInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		Id:               bucketInfo.Id,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
	}); err != nil {
		return nil, err
	}
	return &types.MsgDeleteBucketResponse{}, nil
}

func (k msgServer) UpdateBucketInfo(goCtx context.Context, msg *types.MsgUpdateBucketInfo) (*types.MsgUpdateBucketInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	bucketInfo, found := k.Keeper.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	if bucketInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}
	if bucketInfo.Owner != msg.Operator {
		return nil, types.ErrAccessDenied
	}
	eventUpdateBucketInfo := &types.EventUpdateBucketInfo{
		OperatorAddress:      operatorAcc.String(),
		BucketName:           bucketInfo.BucketName,
		Id:                   bucketInfo.Id,
		ReadQuotaBefore:      bucketInfo.ReadQuota,
		PaymentAddressBefore: bucketInfo.PaymentAddress,
	}

	var paymentAcc sdk.AccAddress
	if msg.PaymentAddress != "" {
		paymentAcc, err = sdk.AccAddressFromHexUnsafe(msg.PaymentAddress)
		if err != nil {
			return nil, err
		}
		if paymentAcc.String() != bucketInfo.PaymentAddress {
			if !k.paymentKeeper.IsPaymentAccountOwner(ctx, bucketInfo.Owner, paymentAcc.String()) {
				return nil, paymenttypes.ErrNotPaymentAccountOwner
			}
			err := k.paymentKeeper.ChargeUpdatePaymentAccount(ctx, &bucketInfo, &msg.PaymentAddress)
			if err != nil {
				return nil, err
			}
			bucketInfo.PaymentAddress = msg.PaymentAddress
		}
	}

	if msg.ReadQuota != bucketInfo.ReadQuota {
		err := k.ChargeUpdateReadQuota(ctx, &bucketInfo, msg.ReadQuota)
		if err != nil {
			return nil, err
		}
		bucketInfo.ReadQuota = msg.ReadQuota
	}
	eventUpdateBucketInfo.ReadQuotaAfter = bucketInfo.ReadQuota
	eventUpdateBucketInfo.PaymentAddressAfter = bucketInfo.PaymentAddress

	k.Keeper.SetBucket(ctx, bucketInfo)
	if err := ctx.EventManager().EmitTypedEvents(eventUpdateBucketInfo); err != nil {
		return nil, err
	}
	return &types.MsgUpdateBucketInfoResponse{}, nil
}

func (k msgServer) CreateObject(goCtx context.Context, msg *types.MsgCreateObject) (*types.MsgCreateObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: check bucket and object permission
	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	if msg.PayloadSize > k.MaxPayloadSize(ctx) {
		return nil, types.ErrTooLargeObject
	}
	// check bucket
	bucketInfo, found := k.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	var secondarySPs []string
	for _, sp := range msg.ExpectSecondarySpAddresses {
		spAcc, err := sdk.AccAddressFromHexUnsafe(sp)
		if err != nil {
			return nil, err
		}
		secondarySPs = append(secondarySPs, spAcc.String())
	}

	if msg.PrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return nil, errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}

	err = k.VerifySPAndSignature(ctx, bucketInfo.PrimarySpAddress, msg.GetApprovalBytes(), msg.PrimarySpApproval.Sig)
	if err != nil {
		return nil, err
	}

	objectInfo := types.ObjectInfo{
		Owner:                ownerAcc.String(),
		BucketName:           msg.BucketName,
		ObjectName:           msg.ObjectName,
		PayloadSize:          msg.PayloadSize,
		IsPublic:             msg.IsPublic,
		ContentType:          msg.ContentType,
		Id:                   k.GetObjectID(ctx),
		CreateAt:             ctx.BlockHeight(),
		ObjectStatus:         types.OBJECT_STATUS_INIT,
		RedundancyType:       types.REDUNDANCY_EC_TYPE, // TODO: base on redundancy policy
		SourceType:           types.SOURCE_TYPE_ORIGIN,
		Checksums:            msg.ExpectChecksums,
		SecondarySpAddresses: secondarySPs,
	}
	err = k.LockStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return nil, err
	}

	err = k.Keeper.CreateObject(ctx, objectInfo)
	if err != nil {
		return nil, err
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateObject{
		CreatorAddress:   ownerAcc.String(),
		OwnerAddress:     objectInfo.Owner,
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		Id:               objectInfo.Id,
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
		return nil, err
	}
	return &types.MsgCreateObjectResponse{}, nil
}

func (k msgServer) SealObject(goCtx context.Context, msg *types.MsgSealObject) (*types.MsgSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check permission when permission module ready
	spSealAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	bucketInfo, found := k.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	spAddr := sdk.MustAccAddressFromHex(bucketInfo.PrimarySpAddress)
	sp, found := k.spKeeper.GetStorageProvider(ctx, spAddr)
	if !found {
		return nil, types.ErrNoSuchStorageProvider
	}

	if sp.SealAddress != spSealAcc.String() {
		return nil, errors.Wrapf(types.ErrAccessDenied, "Only SP's seal address is allowed to SealObject")
	}

	objectInfo, found := k.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_INIT {
		return nil, types.ErrObjectAlreadyExists
	}

	objectInfo.ObjectStatus = types.OBJECT_STATUS_IN_SERVICE

	// SecondarySP signs the root hash(checksum) of all pieces stored on it, and needs to verify that the signature here.
	for i, spAddr := range msg.SecondarySpAddresses {
		spAcc, err := sdk.AccAddressFromHexUnsafe(spAddr)
		if err != nil {
			return nil, err
		}
		sr := types.NewSecondarySpSignDoc(spAcc, objectInfo.Checksums[i+1])
		err = k.VerifySPAndSignature(ctx, spAcc.String(), sr.GetSignBytes(), msg.SecondarySpSignatures[i])
		if err != nil {
			return nil, err
		}
	}

	objectInfo.SecondarySpAddresses = msg.SecondarySpAddresses

	err = k.UnlockAndChargeStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return nil, err
	}
	objectInfo.LockedBalance = nil

	// TODO(fynn): SetBucket/Object will cost a lot every time we set it. So maybe need to split the info to two types
	// key-value, one is immutable attributes, another is mutable attributes, to reduce performance overhead

	k.Keeper.SetBucket(ctx, bucketInfo)
	k.Keeper.SetObject(ctx, objectInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventSealObject{
		OperatorAddress:    spSealAcc.String(),
		BucketName:         bucketInfo.BucketName,
		ObjectName:         objectInfo.ObjectName,
		Id:                 objectInfo.Id,
		Status:             objectInfo.ObjectStatus,
		SecondarySpAddress: objectInfo.SecondarySpAddresses,
	}); err != nil {
		return nil, err
	}
	return &types.MsgSealObjectResponse{}, nil
}

func (k msgServer) CancelCreateObject(goCtx context.Context, msg *types.MsgCancelCreateObject) (*types.MsgCancelCreateObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}
	bucketInfo, found := k.Keeper.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	objectInfo, found := k.Keeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_INIT {
		return nil, types.ErrObjectNotInit
	}

	if operatorAcc.String() != objectInfo.Owner {
		return nil, errors.Wrapf(types.ErrAccessDenied, "Only allowed owner to do cancel create object")
	}
	err = k.UnlockStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return nil, err
	}
	k.Keeper.DeleteObject(ctx, msg.BucketName, msg.ObjectName)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCancelCreateObject{
		OperatorAddress:  operatorAcc.String(),
		BucketName:       bucketInfo.BucketName,
		ObjectName:       objectInfo.ObjectName,
		PrimarySpAddress: bucketInfo.PrimarySpAddress,
		Id:               objectInfo.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCancelCreateObjectResponse{}, nil
}

func (k msgServer) CopyObject(goCtx context.Context, msg *types.MsgCopyObject) (*types.MsgCopyObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	_, found := k.Keeper.GetBucket(ctx, msg.SrcBucketName)
	if !found {
		return nil, errors.Wrapf(types.ErrNoSuchBucket, "src bucket name (%s)", msg.SrcBucketName)
	}

	dstBucketInfo, found := k.Keeper.GetBucket(ctx, msg.DstBucketName)
	if !found {
		return nil, errors.Wrapf(types.ErrNoSuchBucket, "dst bucket name (%s)", msg.DstBucketName)
	}

	srcObjectInfo, found := k.Keeper.GetObject(ctx, msg.SrcBucketName, msg.SrcObjectName)
	if !found {
		return nil, errors.Wrapf(types.ErrNoSuchObject, "src object name (%s)", msg.SrcObjectName)
	}

	if srcObjectInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}

	if msg.DstPrimarySpApproval.ExpiredHeight < uint64(ctx.BlockHeight()) {
		return nil, errors.Wrapf(types.ErrInvalidApproval, "The approval of sp is expired.")
	}

	err = k.VerifySPAndSignature(ctx, dstBucketInfo.PrimarySpAddress, msg.GetApprovalBytes(), msg.DstPrimarySpApproval.Sig)
	if err != nil {
		return nil, err
	}

	// check permission for copy object from this bucket
	// Currently only allowed object owner to CopyObject
	if srcObjectInfo.Owner != msg.Operator {
		return nil, errors.Wrapf(types.ErrAccessDenied, "access denied (%s)", srcObjectInfo.String())
	}
	objectInfo := types.ObjectInfo{
		Owner:          ownerAcc.String(),
		BucketName:     dstBucketInfo.BucketName,
		ObjectName:     msg.DstObjectName,
		PayloadSize:    srcObjectInfo.PayloadSize,
		IsPublic:       srcObjectInfo.IsPublic,
		ContentType:    srcObjectInfo.ContentType,
		CreateAt:       ctx.BlockHeight(),
		Id:             k.GetObjectID(ctx),
		ObjectStatus:   types.OBJECT_STATUS_INIT,
		RedundancyType: types.REDUNDANCY_EC_TYPE,
		SourceType:     types.SOURCE_TYPE_ORIGIN,
		Checksums:      srcObjectInfo.Checksums,
	}

	err = k.LockStoreFee(ctx, &dstBucketInfo, &objectInfo)
	if err != nil {
		return nil, err
	}
	k.Keeper.SetObject(ctx, objectInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCopyObject{
		OperatorAddress: ownerAcc.String(),
		SrcBucketName:   srcObjectInfo.BucketName,
		SrcObjectName:   srcObjectInfo.ObjectName,
		DstBucketName:   objectInfo.BucketName,
		DstObjectName:   objectInfo.ObjectName,
		SrcObjectId:     srcObjectInfo.Id,
		DstObjectId:     objectInfo.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCopyObjectResponse{}, nil
}

func (k msgServer) DeleteObject(goCtx context.Context, msg *types.MsgDeleteObject) (*types.MsgDeleteObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	bucketInfo, found := k.Keeper.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	objectInfo, found := k.Keeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	if objectInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_IN_SERVICE {
		return nil, types.ErrObjectNotInService
	}

	// Currently, only the owner is allowed to delete object
	if objectInfo.Owner != operatorAcc.String() {
		return nil, types.ErrAccessDenied
	}

	err := k.ChargeDeleteObject(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return nil, err
	}
	k.Keeper.DeleteObject(ctx, msg.BucketName, msg.ObjectName)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteObject{
		OperatorAddress:      operatorAcc.String(),
		BucketName:           bucketInfo.BucketName,
		ObjectName:           objectInfo.ObjectName,
		Id:                   objectInfo.Id,
		PrimarySpAddress:     bucketInfo.PrimarySpAddress,
		SecondarySpAddresses: objectInfo.SecondarySpAddresses,
	}); err != nil {
		return nil, err
	}
	return &types.MsgDeleteObjectResponse{}, nil
}

func (k msgServer) RejectSealObject(goCtx context.Context, msg *types.MsgRejectSealObject) (*types.MsgRejectSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	bucketInfo, found := k.Keeper.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	objectInfo, found := k.Keeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_INIT {
		return nil, types.ErrObjectNotInit
	}

	if spAcc.String() != bucketInfo.PrimarySpAddress {
		return nil, errors.Wrapf(types.ErrAccessDenied, "Only allowed primary sp to do cancel create object")
	}

	sp, found := k.spKeeper.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrNoSuchStorageProvider
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return nil, types.ErrStorageProviderNotInService
	}
	err = k.UnlockStoreFee(ctx, &bucketInfo, &objectInfo)
	if err != nil {
		return nil, err
	}

	k.Keeper.DeleteObject(ctx, msg.BucketName, msg.ObjectName)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventRejectSealObject{
		OperatorAddress: spAcc.String(),
		BucketName:      bucketInfo.BucketName,
		ObjectName:      objectInfo.ObjectName,
		Id:              objectInfo.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgRejectSealObjectResponse{}, nil
}

func (k msgServer) CreateGroup(goCtx context.Context, msg *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}
	groupInfo := types.GroupInfo{
		Owner:      ownerAcc.String(),
		SourceType: types.SOURCE_TYPE_ORIGIN,
		Id:         k.GetGroupId(ctx),
		GroupName:  msg.GroupName,
	}
	err = k.Keeper.CreateGroup(ctx, groupInfo)
	if err != nil {
		return nil, err
	}

	// need to limit the size of Msg.Members to avoid taking too long to execute the msg
	for _, member := range msg.Members {
		memberAddress, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return nil, err
		}
		groupMemberInfo := types.GroupMemberInfo{
			Member:     memberAddress.String(),
			Id:         groupInfo.Id,
			ExpireTime: 0,
		}
		err = k.Keeper.AddGroupMember(ctx, groupMemberInfo)
		if err != nil {
			return nil, err
		}
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateGroup{
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		Id:           groupInfo.Id,
		SourceType:   groupInfo.SourceType,
		Members:      msg.Members,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCreateGroupResponse{}, nil
}

func (k msgServer) DeleteGroup(goCtx context.Context, msg *types.MsgDeleteGroup) (*types.MsgDeleteGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	groupInfo, found := k.Keeper.GetGroup(ctx, operatorAcc.String(), msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}
	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	err = k.Keeper.DeleteGroup(ctx, msg.Operator, msg.GroupName)
	if err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeleteGroup{
		OwnerAddress: groupInfo.Owner,
		GroupName:    groupInfo.GroupName,
		Id:           groupInfo.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgDeleteGroupResponse{}, nil
}

func (k msgServer) LeaveGroup(goCtx context.Context, msg *types.MsgLeaveGroup) (*types.MsgLeaveGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	memberAcc, err := sdk.AccAddressFromHexUnsafe(msg.Member)
	if err != nil {
		return nil, err
	}

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.GroupOwner)
	if err != nil {
		return nil, err
	}

	groupInfo, found := k.Keeper.GetGroup(ctx, ownerAcc.String(), msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}

	err = k.Keeper.RemoveGroupMember(ctx, groupInfo.Id, memberAcc.String())
	if err != nil {
		return nil, err
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventLeaveGroup{
		MemberAddress: memberAcc.String(),
		OwnerAddress:  groupInfo.Owner,
		GroupName:     groupInfo.GroupName,
		Id:            groupInfo.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgLeaveGroupResponse{}, nil
}

func (k msgServer) UpdateGroupMember(goCtx context.Context, msg *types.MsgUpdateGroupMember) (*types.MsgUpdateGroupMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}
	// Now only allowed group owner to update member
	groupInfo, found := k.Keeper.GetGroup(ctx, ownerAcc.String(), msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	if groupInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrSourceTypeMismatch
	}

	for _, member := range msg.MembersToAdd {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return nil, err
		}
		memberInfo := types.GroupMemberInfo{
			Member:     memberAcc.String(),
			Id:         groupInfo.Id,
			ExpireTime: 0,
		}

		err = k.Keeper.AddGroupMember(ctx, memberInfo)
		if err != nil {
			return nil, err
		}
	}

	for _, member := range msg.MembersToDelete {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return nil, err
		}
		err = k.Keeper.RemoveGroupMember(ctx, groupInfo.Id, memberAcc.String())
		if err != nil {
			return nil, err
		}
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGroupMember{
		OperatorAddress: ownerAcc.String(),
		OwnerAddress:    groupInfo.Owner,
		GroupName:       groupInfo.GroupName,
		Id:              groupInfo.Id,
		MembersToAdd:    msg.MembersToAdd,
		MembersToDelete: msg.MembersToDelete,
	}); err != nil {
		return nil, err
	}

	return &types.MsgUpdateGroupMemberResponse{}, nil
}
