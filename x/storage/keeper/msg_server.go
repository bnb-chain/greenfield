package keeper

import (
	"context"
	"time"

	errorsmod "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	types2 "github.com/bnb-chain/greenfield/types"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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

	ownerAcc := sdk.MustAccAddressFromHex(msg.Creator)

	primarySPAcc := sdk.MustAccAddressFromHex(msg.PrimarySpAddress)

	id, err := k.Keeper.CreateBucket(ctx, ownerAcc, msg.BucketName, primarySPAcc, &storagetypes.CreateBucketOptions{
		PaymentAddress:    msg.PaymentAddress,
		Visibility:        msg.Visibility,
		ChargedReadQuota:  msg.ChargedReadQuota,
		SourceType:        types.SOURCE_TYPE_ORIGIN,
		PrimarySpApproval: msg.PrimarySpApproval,
		ApprovalMsgBytes:  msg.GetApprovalBytes(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateBucketResponse{
		BucketId: id,
	}, nil
}

func (k msgServer) DeleteBucket(goCtx context.Context, msg *types.MsgDeleteBucket) (*types.MsgDeleteBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.DeleteBucket(ctx, operatorAcc, msg.BucketName, storagetypes.DeleteBucketOptions{
		SourceType: types.SOURCE_TYPE_ORIGIN,
	})
	if err != nil {
		return nil, err
	}
	return &types.MsgDeleteBucketResponse{}, nil
}

func (k msgServer) UpdateBucketInfo(goCtx context.Context, msg *types.MsgUpdateBucketInfo) (*types.MsgUpdateBucketInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.Operator)

	var chargedReadQuota *uint64
	if msg.ChargedReadQuota != nil {
		chargedReadQuota = &msg.ChargedReadQuota.Value
	}
	err := k.Keeper.UpdateBucketInfo(ctx, operatorAcc, msg.BucketName, storagetypes.UpdateBucketOptions{
		SourceType:       types.SOURCE_TYPE_ORIGIN,
		PaymentAddress:   msg.PaymentAddress,
		Visibility:       msg.Visibility,
		ChargedReadQuota: chargedReadQuota,
	})
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateBucketInfoResponse{}, nil
}

func (k msgServer) DiscontinueBucket(goCtx context.Context, msg *storagetypes.MsgDiscontinueBucket) (*storagetypes.MsgDiscontinueBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.DiscontinueBucket(ctx, operatorAcc, msg.BucketName, msg.Reason)
	if err != nil {
		return nil, err
	}
	return &types.MsgDiscontinueBucketResponse{}, nil
}

func (k msgServer) CreateObject(goCtx context.Context, msg *types.MsgCreateObject) (*types.MsgCreateObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc := sdk.MustAccAddressFromHex(msg.Creator)

	if len(msg.ExpectChecksums) != int(1+k.GetExpectSecondarySPNumForECObject(ctx, ctx.BlockTime().Unix())) {
		return nil, gnfderrors.ErrInvalidChecksum.Wrapf("ExpectChecksums missing, expect: %d, actual: %d",
			1+k.Keeper.RedundantParityChunkNum(ctx)+k.Keeper.RedundantParityChunkNum(ctx),
			len(msg.ExpectChecksums))
	}

	id, err := k.Keeper.CreateObject(ctx, ownerAcc, msg.BucketName, msg.ObjectName, msg.PayloadSize, storagetypes.CreateObjectOptions{
		SourceType:        types.SOURCE_TYPE_ORIGIN,
		Visibility:        msg.Visibility,
		ContentType:       msg.ContentType,
		RedundancyType:    msg.RedundancyType,
		Checksums:         msg.ExpectChecksums,
		PrimarySpApproval: msg.PrimarySpApproval,
		ApprovalMsgBytes:  msg.GetApprovalBytes(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateObjectResponse{
		ObjectId: id,
	}, nil
}

func (k msgServer) CancelCreateObject(goCtx context.Context, msg *types.MsgCancelCreateObject) (*types.MsgCancelCreateObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.CancelCreateObject(ctx, operatorAcc, msg.BucketName, msg.ObjectName, storagetypes.CancelCreateObjectOptions{SourceType: types.SOURCE_TYPE_ORIGIN})
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelCreateObjectResponse{}, nil
}

func (k msgServer) SealObject(goCtx context.Context, msg *types.MsgSealObject) (*types.MsgSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spSealAcc := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.SealObject(ctx, spSealAcc, msg.BucketName, msg.ObjectName, SealObjectOptions{
		GlobalVirtualGroupId:     msg.GlobalVirtualGroupId,
		SecondarySpBlsSignatures: msg.SecondarySpBlsAggSignatures,
	})

	if err != nil {
		return nil, err
	}

	return &types.MsgSealObjectResponse{}, nil
}

func (k msgServer) CopyObject(goCtx context.Context, msg *types.MsgCopyObject) (*types.MsgCopyObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc := sdk.MustAccAddressFromHex(msg.Operator)

	id, err := k.Keeper.CopyObject(ctx, ownerAcc, msg.SrcBucketName, msg.SrcObjectName, msg.DstBucketName, msg.DstObjectName, storagetypes.CopyObjectOptions{
		SourceType:        types.SOURCE_TYPE_ORIGIN,
		Visibility:        storagetypes.VISIBILITY_TYPE_PRIVATE,
		PrimarySpApproval: msg.DstPrimarySpApproval,
		ApprovalMsgBytes:  msg.GetApprovalBytes(),
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgCopyObjectResponse{
		ObjectId: id,
	}, nil
}

func (k msgServer) DeleteObject(goCtx context.Context, msg *types.MsgDeleteObject) (*types.MsgDeleteObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.DeleteObject(ctx, operatorAcc, msg.BucketName, msg.ObjectName, storagetypes.DeleteObjectOptions{
		SourceType: types.SOURCE_TYPE_ORIGIN,
	})

	if err != nil {
		return nil, err
	}
	return &types.MsgDeleteObjectResponse{}, nil
}

func (k msgServer) RejectSealObject(goCtx context.Context, msg *types.MsgRejectSealObject) (*types.MsgRejectSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc := sdk.MustAccAddressFromHex(msg.Operator)
	err := k.Keeper.RejectSealObject(ctx, spAcc, msg.BucketName, msg.ObjectName)
	if err != nil {
		return nil, err
	}
	return &types.MsgRejectSealObjectResponse{}, nil
}

func (k msgServer) DiscontinueObject(goCtx context.Context, msg *storagetypes.MsgDiscontinueObject) (*storagetypes.MsgDiscontinueObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc := sdk.MustAccAddressFromHex(msg.Operator)
	err := k.Keeper.DiscontinueObject(ctx, spAcc, msg.BucketName, msg.ObjectIds, msg.Reason)
	if err != nil {
		return nil, err
	}
	return &types.MsgDiscontinueObjectResponse{}, nil
}

func (k msgServer) UpdateObjectInfo(goCtx context.Context, msg *types.MsgUpdateObjectInfo) (*types.MsgUpdateObjectInfoResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc := sdk.MustAccAddressFromHex(msg.Operator)
	err := k.Keeper.UpdateObjectInfo(ctx, spAcc, msg.BucketName, msg.ObjectName, msg.Visibility)
	if err != nil {
		return nil, err
	}
	return &types.MsgUpdateObjectInfoResponse{}, nil
}

func (k msgServer) CreateGroup(goCtx context.Context, msg *types.MsgCreateGroup) (*types.MsgCreateGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	ownerAcc := sdk.MustAccAddressFromHex(msg.Creator)

	id, err := k.Keeper.CreateGroup(ctx, ownerAcc, msg.GroupName, storagetypes.CreateGroupOptions{Members: msg.Members, Extra: msg.Extra})
	if err != nil {
		return nil, err
	}

	return &types.MsgCreateGroupResponse{
		GroupId: id,
	}, nil
}

func (k msgServer) DeleteGroup(goCtx context.Context, msg *types.MsgDeleteGroup) (*types.MsgDeleteGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.Operator)
	err := k.Keeper.DeleteGroup(ctx, operatorAcc, msg.GroupName, storagetypes.DeleteGroupOptions{SourceType: types.SOURCE_TYPE_ORIGIN})
	if err != nil {
		return nil, err
	}

	return &types.MsgDeleteGroupResponse{}, nil
}

func (k msgServer) LeaveGroup(goCtx context.Context, msg *types.MsgLeaveGroup) (*types.MsgLeaveGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	memberAcc := sdk.MustAccAddressFromHex(msg.Member)

	ownerAcc := sdk.MustAccAddressFromHex(msg.GroupOwner)

	err := k.Keeper.LeaveGroup(ctx, memberAcc, ownerAcc, msg.GroupName, storagetypes.LeaveGroupOptions{SourceType: types.SOURCE_TYPE_ORIGIN})
	if err != nil {
		return nil, err
	}

	return &types.MsgLeaveGroupResponse{}, nil
}

func (k msgServer) UpdateGroupMember(goCtx context.Context, msg *types.MsgUpdateGroupMember) (*types.MsgUpdateGroupMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	groupOwner := sdk.MustAccAddressFromHex(msg.GroupOwner)

	groupInfo, found := k.GetGroupInfo(ctx, groupOwner, msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	err := k.Keeper.UpdateGroupMember(ctx, operator, groupInfo, storagetypes.UpdateGroupMemberOptions{
		SourceType:      types.SOURCE_TYPE_ORIGIN,
		MembersToAdd:    msg.MembersToAdd,
		MembersToDelete: msg.MembersToDelete,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateGroupMemberResponse{}, nil
}

func (k msgServer) RenewGroupMember(goCtx context.Context, msg *types.MsgRenewGroupMember) (*types.MsgRenewGroupMemberResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	groupOwner := sdk.MustAccAddressFromHex(msg.GroupOwner)

	groupInfo, found := k.GetGroupInfo(ctx, groupOwner, msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}

	members := make([]string, 0, len(msg.Members))
	membersExpiration := make([]time.Time, 0, len(msg.Members))
	for i := range msg.Members {
		members = append(members, msg.Members[i].GetMember())
		membersExpiration = append(membersExpiration, msg.Members[i].GetExpirationTime())
	}

	err := k.Keeper.RenewGroupMember(ctx, operator, groupInfo, storagetypes.RenewGroupMemberOptions{
		SourceType:        types.SOURCE_TYPE_ORIGIN,
		Members:           members,
		MembersExpiration: membersExpiration,
	})
	if err != nil {
		return nil, err
	}

	return &types.MsgRenewGroupMemberResponse{}, nil
}

func (k msgServer) UpdateGroupExtra(goCtx context.Context, msg *types.MsgUpdateGroupExtra) (*types.MsgUpdateGroupExtraResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	groupOwner := sdk.MustAccAddressFromHex(msg.GroupOwner)

	groupInfo, found := k.GetGroupInfo(ctx, groupOwner, msg.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	err := k.Keeper.UpdateGroupExtra(ctx, operator, groupInfo, msg.Extra)
	if err != nil {
		return nil, err
	}

	return &types.MsgUpdateGroupExtraResponse{}, nil
}

func (k msgServer) PutPolicy(goCtx context.Context, msg *types.MsgPutPolicy) (*types.MsgPutPolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.Operator)

	var grn types2.GRN
	err := grn.ParseFromString(msg.Resource, false)
	if err != nil {
		return nil, err
	}

	if msg.ExpirationTime != nil && msg.ExpirationTime.Before(ctx.BlockTime()) {
		return nil, permtypes.ErrPermissionExpired.Wrapf("The specified policy expiration time is less than the current block time, block time: %s", ctx.BlockTime().String())
	}

	for _, s := range msg.Statements {
		if s.ExpirationTime != nil && s.ExpirationTime.Before(ctx.BlockTime()) {
			return nil, permtypes.ErrPermissionExpired.Wrapf("The specified statement expiration time is less than the current block time, block time: %s", ctx.BlockTime().String())
		}
	}

	policy := &permtypes.Policy{
		ResourceType:   grn.ResourceType(),
		Principal:      msg.Principal,
		Statements:     msg.Statements,
		ExpirationTime: msg.ExpirationTime,
	}

	policyID, err := k.Keeper.PutPolicy(ctx, operatorAddr, grn, policy)
	if err != nil {
		return nil, err
	}
	return &types.MsgPutPolicyResponse{PolicyId: policyID}, nil

}

func (k msgServer) DeletePolicy(goCtx context.Context, msg *types.MsgDeletePolicy) (*types.MsgDeletePolicyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	operator := sdk.MustAccAddressFromHex(msg.Operator)

	var grn types2.GRN
	err := grn.ParseFromString(msg.Resource, false)
	if err != nil {
		return nil, err
	}

	policyID, err := k.Keeper.DeletePolicy(ctx, operator, msg.Principal, grn)
	if err != nil {
		return nil, err
	}

	return &types.MsgDeletePolicyResponse{PolicyId: policyID}, nil
}

func (k msgServer) MirrorObject(goCtx context.Context, msg *types.MsgMirrorObject) (*types.MsgMirrorObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	var objectInfo *types.ObjectInfo
	found := false
	if msg.Id.GT(sdk.NewUint(0)) {
		objectInfo, found = k.Keeper.GetObjectInfoById(ctx, msg.Id)
	} else {
		objectInfo, found = k.Keeper.GetObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	}
	if !found {
		return nil, types.ErrNoSuchObject
	}

	if objectInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrAlreadyMirrored
	}

	if objectInfo.ObjectStatus != types.OBJECT_STATUS_SEALED {
		return nil, types.ErrObjectNotSealed
	}

	if operator.String() != objectInfo.Owner {
		return nil, types.ErrAccessDenied
	}

	owner := sdk.MustAccAddressFromHex(objectInfo.Owner)

	mirrorPackage := types.MirrorObjectSynPackage{
		Id:    objectInfo.Id.BigInt(),
		Owner: owner,
	}

	encodedPackage, err := mirrorPackage.Serialize()
	if err != nil {
		return nil, types.ErrInvalidCrossChainPackage
	}

	wrapPackage := types.CrossChainPackage{
		OperationType: types.OperationMirrorObject,
		Package:       encodedPackage,
	}
	encodedWrapPackage := wrapPackage.MustSerialize()

	relayerFee := k.Keeper.MirrorObjectRelayerFee(ctx)
	ackRelayerFee := k.Keeper.MirrorObjectAckRelayerFee(ctx)

	_, err = k.crossChainKeeper.CreateRawIBCPackageWithFee(ctx, k.crossChainKeeper.GetDestBscChainID(),
		types.ObjectChannelId, sdk.SynCrossChainPackageType, encodedWrapPackage, relayerFee, ackRelayerFee)
	if err != nil {
		return nil, err
	}

	// update source type to pending
	objectInfo.SourceType = types.SOURCE_TYPE_MIRROR_PENDING
	k.Keeper.SetObjectInfo(ctx, objectInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorObject{
		Operator:    objectInfo.Owner,
		BucketName:  objectInfo.BucketName,
		ObjectName:  objectInfo.ObjectName,
		ObjectId:    objectInfo.Id,
		DestChainId: uint32(k.crossChainKeeper.GetDestBscChainID()),
	}); err != nil {
		return nil, err
	}
	return nil, nil
}

func (k msgServer) MirrorBucket(goCtx context.Context, msg *types.MsgMirrorBucket) (*types.MsgMirrorBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	var bucketInfo *types.BucketInfo
	found := false
	if msg.Id.GT(sdk.NewUint(0)) {
		bucketInfo, found = k.Keeper.GetBucketInfoById(ctx, msg.Id)
	} else {
		bucketInfo, found = k.Keeper.GetBucketInfo(ctx, msg.BucketName)
	}
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	if bucketInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrAlreadyMirrored
	}

	if operator.String() != bucketInfo.Owner {
		return nil, types.ErrAccessDenied
	}

	owner, err := sdk.AccAddressFromHexUnsafe(bucketInfo.Owner)
	if err != nil {
		return nil, err
	}

	mirrorPackage := types.MirrorBucketSynPackage{
		Id:    bucketInfo.Id.BigInt(),
		Owner: owner,
	}

	encodedPackage, err := mirrorPackage.Serialize()
	if err != nil {
		return nil, types.ErrInvalidCrossChainPackage
	}

	wrapPackage := types.CrossChainPackage{
		OperationType: types.OperationMirrorBucket,
		Package:       encodedPackage,
	}
	encodedWrapPackage := wrapPackage.MustSerialize()

	relayerFee := k.Keeper.MirrorBucketRelayerFee(ctx)
	ackRelayerFee := k.Keeper.MirrorBucketAckRelayerFee(ctx)

	_, err = k.crossChainKeeper.CreateRawIBCPackageWithFee(ctx, k.crossChainKeeper.GetDestBscChainID(),
		types.BucketChannelId, sdk.SynCrossChainPackageType, encodedWrapPackage, relayerFee, ackRelayerFee)
	if err != nil {
		return nil, err
	}

	// update status to pending
	bucketInfo.SourceType = types.SOURCE_TYPE_MIRROR_PENDING
	k.Keeper.SetBucketInfo(ctx, bucketInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorBucket{
		Operator:    bucketInfo.Owner,
		BucketName:  bucketInfo.BucketName,
		BucketId:    bucketInfo.Id,
		DestChainId: uint32(k.crossChainKeeper.GetDestBscChainID()),
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (k msgServer) MirrorGroup(goCtx context.Context, msg *types.MsgMirrorGroup) (*types.MsgMirrorGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	var groupInfo *types.GroupInfo
	found := false
	if msg.Id.GT(sdk.NewUint(0)) {
		groupInfo, found = k.Keeper.GetGroupInfoById(ctx, msg.Id)
	} else {
		groupInfo, found = k.Keeper.GetGroupInfo(ctx, operator, msg.GroupName)
	}
	if !found {
		return nil, types.ErrNoSuchGroup
	}

	if groupInfo.SourceType != types.SOURCE_TYPE_ORIGIN {
		return nil, types.ErrAlreadyMirrored
	}

	if operator.String() != groupInfo.Owner {
		return nil, types.ErrAccessDenied
	}

	mirrorPackage := types.MirrorGroupSynPackage{
		Id:    groupInfo.Id.BigInt(),
		Owner: operator,
	}

	encodedPackage, err := mirrorPackage.Serialize()
	if err != nil {
		return nil, types.ErrInvalidCrossChainPackage
	}

	wrapPackage := types.CrossChainPackage{
		OperationType: types.OperationMirrorGroup,
		Package:       encodedPackage,
	}
	encodedWrapPackage := wrapPackage.MustSerialize()

	relayerFee := k.Keeper.MirrorGroupRelayerFee(ctx)
	ackRelayerFee := k.Keeper.MirrorGroupAckRelayerFee(ctx)

	_, err = k.crossChainKeeper.CreateRawIBCPackageWithFee(ctx, k.crossChainKeeper.GetDestBscChainID(),
		types.GroupChannelId, sdk.SynCrossChainPackageType, encodedWrapPackage, relayerFee, ackRelayerFee)
	if err != nil {
		return nil, err
	}

	// update source type to pending
	groupInfo.SourceType = types.SOURCE_TYPE_MIRROR_PENDING
	k.Keeper.SetGroupInfo(ctx, groupInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorGroup{
		Owner:       groupInfo.Owner,
		GroupName:   groupInfo.GroupName,
		GroupId:     groupInfo.Id,
		DestChainId: uint32(k.crossChainKeeper.GetDestBscChainID()),
	}); err != nil {
		return nil, err
	}

	return nil, nil
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) MigrateBucket(goCtx context.Context, msg *types.MsgMigrateBucket) (*types.MsgMigrateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	operator := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.MigrateBucket(ctx, operator, msg.BucketName, msg.DstPrimarySpId, msg.DstPrimarySpApproval, msg.GetApprovalBytes())
	if err != nil {
		return nil, err
	}

	return &types.MsgMigrateBucketResponse{}, nil
}

func (k msgServer) CompleteMigrateBucket(goCtx context.Context, msg *types.MsgCompleteMigrateBucket) (*types.MsgCompleteMigrateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.Keeper.CompleteMigrateBucket(ctx, operator, msg.BucketName, msg.GlobalVirtualGroupFamilyId, msg.GvgMappings)
	if err != nil {
		return nil, err
	}

	return &types.MsgCompleteMigrateBucketResponse{}, nil
}

func (k msgServer) CancelMigrateBucket(goCtx context.Context, msg *types.MsgCancelMigrateBucket) (*types.MsgCancelMigrateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	err := k.CancelBucketMigration(ctx, operator, msg.BucketName)
	if err != nil {
		return nil, err
	}

	return &types.MsgCancelMigrateBucketResponse{}, nil
}

func (k Keeper) verifyGVGSignatures(ctx sdk.Context, bucketID math.Uint, dstSP *sptypes.StorageProvider, gvgMappings []*storagetypes.GVGMapping) error {
	// verify secondary sp signature
	for _, newLvg2gvg := range gvgMappings {
		dstGVG, found := k.virtualGroupKeeper.GetGVG(ctx, newLvg2gvg.DstGlobalVirtualGroupId)
		if !found {
			return virtualgroupmoduletypes.ErrGVGNotExist.Wrapf("dst gvg not found")
		}
		// validate the aggregated bls signature
		migrationBucketSignDocBlsSignHash := storagetypes.NewSecondarySpMigrationBucketSignDoc(ctx.ChainID(), bucketID, dstSP.Id, newLvg2gvg.SrcGlobalVirtualGroupId, dstGVG.Id).GetBlsSignHash()
		err := k.VerifyGVGSecondarySPsBlsSignature(ctx, dstGVG, migrationBucketSignDocBlsSignHash, newLvg2gvg.SecondarySpBlsSignature)
		if err != nil {
			return err
		}
	}
	return nil
}
