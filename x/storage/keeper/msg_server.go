package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/tendermint/tendermint/crypto"
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
	var (
		ownerAcc     sdk.AccAddress
		paymentAcc   sdk.AccAddress
		primarySPAcc sdk.AccAddress
		err          error
	)

	ownerAcc, err = sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	if msg.PaymentAddress != "" {
		// TODO: validate that the paymentAcc is ownered by ownerAcc if payment module ready
		paymentAcc, err = sdk.AccAddressFromHexUnsafe(msg.PaymentAddress)
		if err != nil {
			return nil, err
		}
	} else {
		paymentAcc = ownerAcc
	}

	primarySPAcc, err = sdk.AccAddressFromHexUnsafe(msg.PrimarySpAddress)
	if err != nil {
		return nil, err
	}

	spApproval := msg.PrimarySpApproval
	msg.PrimarySpApproval = []byte("")
	bz, err := msg.Marshal()
  if err != nil {
    return nil, err
  }

  err = k.CheckSPAndSignature(ctx, []string{msg.PrimarySpAddress}, [][]byte{crypto.Sha256(bz)}, [][]byte{spApproval})
  if err != nil {
    return nil, err
  }

	// Check Bucket exist
	bucketKey := types.GetBucketKey(msg.BucketName)
	if k.HasBucket(ctx, bucketKey) {
		return nil, types.ErrBucketAlreadyExists
	}

	// Store bucket meta
	bucketInfo := types.BucketInfo{
		Owner:            ownerAcc.String(),
		BucketName:       msg.BucketName,
		IsPublic:         msg.IsPublic,
		CreateAt:         ctx.BlockHeight(),
		PrimarySpAddress: primarySPAcc.String(),
		ReadQuota:        types.READ_QUOTA_FREE,
		PaymentAddress:   paymentAcc.String(),
	}
	k.SetBucket(ctx, bucketKey, bucketInfo)

	return &types.MsgCreateBucketResponse{}, nil
}

func (k msgServer) DeleteBucket(goCtx context.Context, msg *types.MsgDeleteBucket) (*types.MsgDeleteBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check if have the permission to delete bucket

	// check if the bucket exists
	bucketKey := types.GetBucketKey(msg.BucketName)
	if k.Keeper.HasBucket(ctx, bucketKey) {
		return nil, types.ErrBucketAlreadyExists
	}

	// check if the bucket empty
	if k.Keeper.IsEmptyBucket(ctx, bucketKey) {
		return nil, types.ErrBucketNotEmpty
	}

	k.Keeper.DeleteBucket(ctx, bucketKey)
	return &types.MsgDeleteBucketResponse{}, nil
}

func (k msgServer) CreateObject(goCtx context.Context, msg *types.MsgCreateObject) (*types.MsgCreateObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: check bucket and object permission
	// TODO: pay for the object. Interact with PaymentModule

	var (
		ownerAcc     sdk.AccAddress
		err          error
	)
	// check owner AccAddress
	if ownerAcc, err = sdk.AccAddressFromHexUnsafe(msg.Creator); err != nil {
		return nil, err
	}

	// check bucket
	bucketKey := types.GetBucketKey(msg.BucketName)
	bucketInfo, found := k.GetBucket(ctx, bucketKey)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	// check object
	objectKey := types.GetObjectKey(msg.BucketName, msg.ObjectName)
	if k.HasObject(ctx, objectKey) {
		return nil, types.ErrObjectAlreadyExists
	}

	spApproval := msg.PrimarySpApproval
	msg.PrimarySpApproval = []byte("")
	bz, err := msg.Marshal()
  if err != nil {
    return nil, err
  }
  err = k.CheckSPAndSignature(ctx, []string{bucketInfo.PrimarySpAddress}, [][]byte{crypto.Sha256(bz)}, [][]byte{spApproval})
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
		CreateAt:       ctx.BlockHeight(),
		ObjectStatus:   types.OBJECT_STATUS_INIT,
		RedundancyType: types.REDUNDANCY_EC_TYPE, // TODO: base on redundancy policy
		SourceType:     types.SOURCE_TYPE_ORIGIN,

		Checksums:            msg.ExpectChecksums,
		SecondarySpAddresses: msg.ExpectSecondarySpAddresses,
	}
	k.Keeper.SetObject(ctx, objectKey, objectInfo)

	return &types.MsgCreateObjectResponse{}, nil
}

func (k msgServer) SealObject(goCtx context.Context, msg *types.MsgSealObject) (*types.MsgSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: check permission when permission module ready
	// TODO: submit event/log
	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}
	bucketKey := types.GetBucketKey(msg.BucketName)
	objectKey := types.GetObjectKey(msg.BucketName, msg.ObjectName)

	bucketInfo, found := k.Keeper.GetBucket(ctx, bucketKey)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	if bucketInfo.PrimarySpAddress != spAcc.String() {
		return nil, types.ErrSPAddressMismatch
	}

	objectInfo, found := k.Keeper.GetObject(ctx, objectKey)
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

	k.Keeper.SetObject(ctx, objectKey, objectInfo)

	return &types.MsgSealObjectResponse{}, nil
}

func (k msgServer) CopyObject(goCtx context.Context, msg *types.MsgCopyObject) (*types.MsgCopyObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	var (
		ownerAcc sdk.AccAddress
		err      error
	)
	ownerAcc, err = sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	srcBucketKey := types.GetBucketKey(msg.SrcBucketName)
	_, found := k.Keeper.GetBucket(ctx, srcBucketKey)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoSuchBucket, "src bucket name (%s)", msg.SrcBucketName)
	}

	dstBucketKey := types.GetBucketKey(msg.DstBucketName)
	dstBucketInfo, found := k.Keeper.GetBucket(ctx, dstBucketKey)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoSuchBucket, "dst bucket name (%s)", msg.DstBucketName)
	}

	srcObjectKey := types.GetObjectKey(msg.SrcBucketName, msg.SrcObjectName)
	srcObjectInfo, found := k.Keeper.GetObject(ctx, srcObjectKey)
	if !found {
		return nil, sdkerrors.Wrapf(types.ErrNoSuchObject, "src object name (%s)", msg.SrcObjectName)
	}

	// check if have permission for copy object from this bucket
	// Currently only allowed object owner to CopyObject
	if srcObjectInfo.Owner != msg.Creator {
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

	objectKey := types.GetObjectKey(msg.DstBucketName, msg.DstObjectName)
	k.Keeper.SetObject(ctx, objectKey, objectInfo)

	return &types.MsgCopyObjectResponse{}, nil
}

func (k msgServer) DeleteObject(goCtx context.Context, msg *types.MsgDeleteObject) (*types.MsgDeleteObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	objectKey := types.GetObjectKey(msg.BucketName, msg.ObjectName)
	objectInfo, found := k.Keeper.GetObject(ctx, objectKey)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	// Currently, only the owner is allowed to delete object
	if objectInfo.Owner != msg.Creator {
		return nil, types.ErrAccessDenied
	}

	k.Keeper.DeleteObject(ctx, objectKey)
	return &types.MsgDeleteObjectResponse{}, nil
}

func (k msgServer) RejectUnsealedObject(goCtx context.Context, msg *types.MsgRejectUnsealedObject) (*types.MsgRejectUnsealedObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	objectKey := types.GetObjectKey(msg.BucketName, msg.ObjectName)
	objectInfo, found := k.Keeper.GetObject(ctx, objectKey)
	if !found {
		return nil, types.ErrNoSuchObject
	}

	// Currently, only the owner is allowed to reject object
	if objectInfo.Owner != msg.Creator {
		return nil, types.ErrAccessDenied
	}

	// TODO: Interact with payment. unlock the pre-pay fee.
	k.Keeper.DeleteObject(ctx, objectKey)
	return &types.MsgRejectUnsealedObjectResponse{}, nil
}

func (k msgServer) CreateGroup(goCtx context.Context, msg *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	groupKey := types.GetGroupKey(ownerAcc, msg.GroupName)
	if k.Keeper.HasGroup(ctx, groupKey) {
		return nil, types.ErrObjectAlreadyExists
	}

	groupInfo := types.GroupInfo{
		Owner:     msg.Creator,
		GroupName: msg.GroupName,
	}

	// TODO: add a member means insert a key-value to database.
	// need to limit the size of Msg.Members to avoid taking too long to execute the msg
	for _, member := range msg.Members {
		memberAddress, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return nil, err
		}
		groupMemberInfo := types.GroupMemberInfo{
			Id:         groupInfo.Id,
			ExpireTime: 0,
		}
		groupMemberKey := types.GetGroupMemberKey(groupInfo.Id, memberAddress)
		k.Keeper.SetGroupMember(ctx, groupMemberKey, groupMemberInfo)
	}

	k.Keeper.SetGroup(ctx, groupKey, groupInfo)
	return &types.MsgCreateGroupResponse{}, nil
}

func (k msgServer) DeleteGroup(goCtx context.Context, msg *types.MsgDeleteGroup) (*types.MsgDeleteGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	groupKey := types.GetGroupKey(ownerAcc, msg.GroupName)
	if !k.Keeper.HasGroup(ctx, groupKey) {
		return nil, types.ErrNoSuchGroup
	}

	// Note: Delete group does not require the group is empty. The group member will be deleted by on-chain GC.
	k.Keeper.DeleteGroup(ctx, groupKey)
	return &types.MsgDeleteGroupResponse{}, nil
}

func (k msgServer) LeaveGroup(goCtx context.Context, msg *types.MsgLeaveGroup) (*types.MsgLeaveGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	memberAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.GroupOwner)
	if err != nil {
		return nil, err
	}
	groupKey := types.GetGroupKey(ownerAcc, msg.GroupName)
	groupInfo, found := k.Keeper.GetGroup(ctx, groupKey)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	memberKey := types.GetGroupMemberKey(groupInfo.Id, memberAcc)
	if !k.Keeper.HasGroupMember(ctx, memberKey) {
		return nil, types.ErrNoSuchGroupMember
	}
	k.Keeper.DeleteGroupMember(ctx, memberKey)

	return &types.MsgLeaveGroupResponse{}, nil
}

func (k msgServer) UpdateGroupMember(goCtx context.Context, msg *types.MsgUpdateGroupMember) (*types.MsgUpdateGroupMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return nil, err
	}
	// Now only allowed group owner to update member
	groupKey := types.GetGroupKey(ownerAcc, msg.GroupName)
	groupInfo, found := k.Keeper.GetGroup(ctx, groupKey)
	if !found {
		return nil, types.ErrNoSuchGroup
	}

	for _, member := range msg.MembersToAdd {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return nil, err
		}
		memberKey := types.GetGroupMemberKey(groupInfo.Id, memberAcc)
		memberInfo := types.GroupMemberInfo{
			Id:         groupInfo.Id,
			ExpireTime: 0,
		}
		if !k.Keeper.HasGroupMember(ctx, memberKey) {
			k.Keeper.SetGroupMember(ctx, memberKey, memberInfo)
		} else {
			return nil, types.ErrGroupMemberAlreadyExists
		}
	}

	for _, member := range msg.MembersToDelete {
		memberAcc, err := sdk.AccAddressFromHexUnsafe(member)
		if err != nil {
			return nil, err
		}
		memberKey := types.GetGroupMemberKey(groupInfo.Id, memberAcc)
		if !k.Keeper.HasGroupMember(ctx, memberKey) {
			k.Keeper.DeleteGroupMember(ctx, memberKey)
		} else {
			return nil, types.ErrGroupMemberAlreadyExists
		}
	}

	return &types.MsgUpdateGroupMemberResponse{}, nil
}
