package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type msgServer struct {
	Keeper
}

// TODO: add event for all message

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
		// TODO: validate that the paymentAcc is ownered by ownerAcc if payment module is ready
		paymentAcc, err = sdk.AccAddressFromHexUnsafe(msg.PaymentAddress)
		if err != nil {
			return nil, err
		}
	} else {
		paymentAcc = ownerAcc
	}

	primarySPAcc, err := sdk.AccAddressFromHexUnsafe(msg.PrimarySpAddress)
	if err != nil {
		return nil, err
	}

	// TODO: this is a very tricky implement. Will be refactor later.
	spApproval := msg.PrimarySpApprovalSignature
	msg.PrimarySpApprovalSignature = []byte("")
	bz, err := msg.Marshal()
	if err != nil {
		return nil, err
	}

	err = k.CheckSPAndSignature(ctx, []string{msg.PrimarySpAddress}, [][]byte{sdk.Keccak256(bz)}, [][]byte{spApproval})
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
		ReadQuota:        types.READ_QUOTA_FREE,
		PaymentAddress:   paymentAcc.String(),
		PrimarySpAddress: primarySPAcc.String(),
	}

	return &types.MsgCreateBucketResponse{}, k.Keeper.CreateBucket(ctx, bucketInfo)
}

func (k msgServer) DeleteBucket(goCtx context.Context, msg *types.MsgDeleteBucket) (*types.MsgDeleteBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check if have the permission to delete bucket
	return &types.MsgDeleteBucketResponse{}, k.Keeper.DeleteBucket(ctx, msg.BucketName)
}

func (k msgServer) CreateObject(goCtx context.Context, msg *types.MsgCreateObject) (*types.MsgCreateObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: check bucket and object permission
	// TODO: pay for the object. Interact with PaymentModule

	// check owner AccAddress
	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	// check bucket
	bucketInfo, found := k.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	// TODO: this is a very tricky implement. Will be refactor later.
	spApproval := msg.PrimarySpApprovalSignature
	msg.PrimarySpApprovalSignature = []byte("")
	bz, err := msg.Marshal()
	if err != nil {
		return nil, err
	}
	err = k.CheckSPAndSignature(ctx, []string{bucketInfo.PrimarySpAddress}, [][]byte{sdk.Keccak256(bz)}, [][]byte{spApproval})
	if err != nil {
		return nil, err
	}

	objectInfo := types.ObjectInfo{
		Owner:          ownerAcc.String(),
		BucketName:     msg.BucketName,
		ObjectName:     msg.ObjectName,
		PayloadSize:    msg.PayloadSize,
		IsPublic:       msg.IsPublic,
		ContentType:    msg.ContentType,
		Id:             k.GetObjectID(ctx),
		CreateAt:       ctx.BlockHeight(),
		ObjectStatus:   types.OBJECT_STATUS_INIT,
		RedundancyType: types.REDUNDANCY_EC_TYPE, // TODO: base on redundancy policy
		SourceType:     types.SOURCE_TYPE_ORIGIN,

		Checksums:            msg.ExpectChecksums,
		SecondarySpAddresses: msg.ExpectSecondarySpAddresses,
	}

	return &types.MsgCreateObjectResponse{}, k.Keeper.CreateObject(ctx, objectInfo)
}

func (k msgServer) SealObject(goCtx context.Context, msg *types.MsgSealObject) (*types.MsgSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check permission when permission module ready
	// TODO: submit event/log
	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}
	bucketInfo, found := k.Keeper.GetBucket(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	if bucketInfo.PrimarySpAddress != spAcc.String() {
		return nil, types.ErrSPAddressMismatch
	}

	objectInfo, found := k.Keeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}
	if objectInfo.ObjectStatus != types.OBJECT_STATUS_INIT {
		return nil, types.ErrObjectAlreadyExists
	} else {
		objectInfo.ObjectStatus = types.OBJECT_STATUS_IN_SERVICE
	}

	err = k.CheckSPAndSignature(ctx, msg.SecondarySpAddresses, objectInfo.Checksums[1:], msg.SecondarySpSignatures)
	if err != nil {
		return nil, err
	}

	k.Keeper.SetObject(ctx, objectInfo)

	return &types.MsgSealObjectResponse{}, nil
}

func (k msgServer) CopyObject(goCtx context.Context, msg *types.MsgCopyObject) (*types.MsgCopyObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}

	_, found := k.Keeper.GetBucket(ctx, msg.SrcBucketName)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoSuchBucket, "src bucket name (%s)", msg.SrcBucketName)
	}

	dstBucketInfo, found := k.Keeper.GetBucket(ctx, msg.DstBucketName)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoSuchBucket, "dst bucket name (%s)", msg.DstBucketName)
	}

	srcObjectInfo, found := k.Keeper.GetObject(ctx, msg.SrcBucketName, msg.SrcObjectName)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoSuchObject, "src object name (%s)", msg.SrcObjectName)
	}

	// TODO: this is a very tricky implement. Will be refactor later.
	spApproval := msg.DstPrimarySpApprovalSignature
	msg.DstPrimarySpApprovalSignature = []byte("")
	bz, err := msg.Marshal()
	if err != nil {
		return nil, err
	}
	err = k.CheckSPAndSignature(ctx, []string{dstBucketInfo.PrimarySpAddress}, [][]byte{sdk.Keccak256(bz)}, [][]byte{spApproval})
	if err != nil {
		return nil, err
	}

	// check if have permission for copy object from this bucket
	// Currently only allowed object owner to CopyObject
	if srcObjectInfo.Owner != msg.Operator {
		return nil, sdkerrors.Wrapf(types.ErrAccessDenied, "access denied (%s)", srcObjectInfo.String())
	}
	objectInfo := types.ObjectInfo{
		Owner:          ownerAcc.String(),
		BucketName:     dstBucketInfo.BucketName,
		ObjectName:     msg.DstObjectName,
		PayloadSize:    srcObjectInfo.PayloadSize,
		IsPublic:       srcObjectInfo.IsPublic,
		ContentType:    srcObjectInfo.ContentType,
		CreateAt:       ctx.BlockHeight(),
		ObjectStatus:   types.OBJECT_STATUS_INIT,
		RedundancyType: types.REDUNDANCY_EC_TYPE,
		SourceType:     types.SOURCE_TYPE_ORIGIN,

		Checksums: srcObjectInfo.Checksums,
	}

	k.Keeper.SetObject(ctx, objectInfo)

	return &types.MsgCopyObjectResponse{}, nil
}

func (k msgServer) DeleteObject(goCtx context.Context, msg *types.MsgDeleteObject) (*types.MsgDeleteObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	objectInfo, found := k.Keeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	// Currently, only the owner is allowed to delete object
	if objectInfo.Owner != msg.Operator {
		return nil, types.ErrAccessDenied
	}

	k.Keeper.DeleteObject(ctx, msg.BucketName, msg.ObjectName)
	return &types.MsgDeleteObjectResponse{}, nil
}

func (k msgServer) RejectSealObject(goCtx context.Context, msg *types.MsgRejectSealObject) (*types.MsgRejectSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	objectInfo, found := k.Keeper.GetObject(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_INIT {
		return nil, types.ErrObjectStatusNotInit
	}

	// Currently, only the primary sp is allowed to reject seal object
	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.Operator)
	if err != nil {
		return nil, err
	}
	// TODO: operator address or other address (for reject seal object)
	sp, found := k.spKeeper.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrNoSuchStorageProvider
	}
	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return nil, types.ErrStorageProviderNotInService
	}

	// TODO: Interact with payment. unlock the pre-pay fee.
	k.Keeper.DeleteObject(ctx, msg.BucketName, msg.ObjectName)
	return &types.MsgRejectSealObjectResponse{}, nil
}

func (k msgServer) CreateGroup(goCtx context.Context, msg *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupInfo := types.GroupInfo{
		Owner:     msg.Creator,
		Id:        k.GetGroupId(ctx),
		GroupName: msg.GroupName,
	}
	err := k.Keeper.CreateGroup(ctx, groupInfo)
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
	return &types.MsgCreateGroupResponse{}, nil
}

func (k msgServer) DeleteGroup(goCtx context.Context, msg *types.MsgDeleteGroup) (*types.MsgDeleteGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	return &types.MsgDeleteGroupResponse{}, k.Keeper.DeleteGroup(ctx, msg.Operator, msg.GroupName)
}

func (k msgServer) LeaveGroup(goCtx context.Context, msg *types.MsgLeaveGroup) (*types.MsgLeaveGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	groupInfo, found := k.Keeper.GetGroup(ctx, msg.GroupOwner, msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	return &types.MsgLeaveGroupResponse{}, k.Keeper.RemoveGroupMember(ctx, groupInfo.Id, msg.Member)
}

func (k msgServer) UpdateGroupMember(goCtx context.Context, msg *types.MsgUpdateGroupMember) (*types.MsgUpdateGroupMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// Now only allowed group owner to update member
	groupInfo, found := k.Keeper.GetGroup(ctx, msg.Operator, msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
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

	return &types.MsgUpdateGroupMemberResponse{}, nil
}
