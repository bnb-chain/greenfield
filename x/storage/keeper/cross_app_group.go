package keeper

import (
	"encoding/hex"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ sdk.CrossChainApplication = &GroupApp{}

type GroupApp struct {
	storageKeeper types.StorageKeeper
}

func NewGroupApp(keeper types.StorageKeeper) *GroupApp {
	return &GroupApp{
		storageKeeper: keeper,
	}
}

func (app *GroupApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.GroupChannelId, sdk.AckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize group cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize group cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorGroupAckPackage:
		operationType = types.OperationMirrorGroup
		result = app.handleMirrorGroupAckPackage(ctx, appCtx, p)
	case *types.CreateGroupAckPackage:
		operationType = types.OperationCreateGroup
		result = app.handleCreateGroupAckPackage(ctx, appCtx, p)
	case *types.DeleteGroupAckPackage:
		operationType = types.OperationDeleteGroup
		result = app.handleDeleteGroupAckPackage(ctx, appCtx, p)
	case *types.UpdateGroupMemberAckPackage:
		operationType = types.OperationUpdateGroupMember
		result = app.handleUpdateGroupMemberAckPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		result.Payload = wrapPayload.MustSerialize()
	}

	return result
}

func (app *GroupApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.GroupChannelId, sdk.FailAckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize group cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize group cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorGroupSynPackage:
		operationType = types.OperationMirrorGroup
		result = app.handleMirrorGroupFailAckPackage(ctx, appCtx, p)
	case *types.CreateGroupSynPackage:
		operationType = types.OperationCreateGroup
		result = app.handleCreateGroupFailAckPackage(ctx, appCtx, p)
	case *types.DeleteGroupSynPackage:
		operationType = types.OperationDeleteGroup
		result = app.handleDeleteGroupFailAckPackage(ctx, appCtx, p)
	case *types.UpdateGroupMemberSynPackage:
		operationType = types.OperationUpdateGroupMember
		result = app.handleUpdateGroupMemberFailAckPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		result.Payload = wrapPayload.MustSerialize()
	}

	return result
}

func (app *GroupApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.GroupChannelId, sdk.SynCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize group cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize group cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorGroupSynPackage:
		operationType = types.OperationMirrorGroup
		result = app.handleMirrorGroupSynPackage(ctx, appCtx, p)
	case *types.CreateGroupSynPackage:
		operationType = types.OperationCreateGroup
		result = app.handleCreateGroupSynPackage(ctx, appCtx, p)
	case *types.DeleteGroupSynPackage:
		operationType = types.OperationDeleteGroup
		result = app.handleDeleteGroupSynPackage(ctx, appCtx, p)
	case *types.UpdateGroupMemberSynPackage:
		operationType = types.OperationUpdateGroupMember
		result = app.handleUpdateGroupMemberSynPackage(ctx, appCtx, p)
	case *types.UpdateGroupMemberV2SynPackage:
		operationType = types.OperationUpdateGroupMember
		result = app.handleUpdateGroupMemberV2SynPackage(ctx, appCtx, p)
	default:
		return sdk.ExecuteResult{
			Err: types.ErrInvalidCrossChainPackage,
		}
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		result.Payload = wrapPayload.MustSerialize()
	}

	return result
}

func (app *GroupApp) handleDeleteGroupAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.DeleteGroupAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete group ack package ")

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleDeleteGroupFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.DeleteGroupSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete group fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleDeleteGroupSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, deleteGroupPackage *types.DeleteGroupSynPackage) sdk.ExecuteResult {
	err := deleteGroupPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteGroupAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	app.storageKeeper.Logger(ctx).Info("process delete group syn package", "group id", deleteGroupPackage.Id.String())

	groupInfo, found := app.storageKeeper.GetGroupInfoById(ctx, math.NewUintFromBigInt(deleteGroupPackage.Id))
	if !found {
		return sdk.ExecuteResult{
			Payload: types.DeleteGroupAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchGroup,
		}
	}

	err = app.storageKeeper.DeleteGroup(ctx,
		deleteGroupPackage.Operator,
		groupInfo.GroupName,
		types.DeleteGroupOptions{
			SourceType: types.SOURCE_TYPE_BSC_CROSS_CHAIN,
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteGroupAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.DeleteGroupAckPackage{
			Status:    types.StatusSuccess,
			Id:        groupInfo.Id.BigInt(),
			ExtraData: deleteGroupPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *GroupApp) handleCreateGroupAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.CreateGroupAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create group ack package ")

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleCreateGroupFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.CreateGroupSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create group fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleCreateGroupSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, createGroupPackage *types.CreateGroupSynPackage) sdk.ExecuteResult {
	err := createGroupPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateGroupAckPackage{
				Status:    types.StatusFail,
				Creator:   createGroupPackage.Creator,
				ExtraData: createGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}
	app.storageKeeper.Logger(ctx).Info("process create group syn package", "group name", createGroupPackage.GroupName)

	groupId, err := app.storageKeeper.CreateGroup(ctx,
		createGroupPackage.Creator,
		createGroupPackage.GroupName,
		types.CreateGroupOptions{
			SourceType: types.SOURCE_TYPE_BSC_CROSS_CHAIN,
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateGroupAckPackage{
				Status:    types.StatusFail,
				Creator:   createGroupPackage.Creator,
				ExtraData: createGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.CreateGroupAckPackage{
			Status:    types.StatusSuccess,
			Id:        groupId.BigInt(),
			Creator:   createGroupPackage.Creator,
			ExtraData: createGroupPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *GroupApp) handleMirrorGroupAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.MirrorGroupAckPackage) sdk.ExecuteResult {
	groupInfo, found := app.storageKeeper.GetGroupInfoById(ctx, math.NewUintFromBigInt(ackPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("group does not exist", "group id", ackPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchGroup,
		}
	}

	if ackPackage.Status == types.StatusSuccess {
		groupInfo.SourceType = types.SOURCE_TYPE_BSC_CROSS_CHAIN

		app.storageKeeper.SetGroupInfo(ctx, groupInfo)
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorGroupResult{
		Status:      uint32(ackPackage.Status),
		GroupName:   groupInfo.GroupName,
		GroupId:     groupInfo.Id,
		DestChainId: uint32(appCtx.SrcChainId),
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleMirrorGroupFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, mirrorGroupPackage *types.MirrorGroupSynPackage) sdk.ExecuteResult {
	groupInfo, found := app.storageKeeper.GetGroupInfoById(ctx, math.NewUintFromBigInt(mirrorGroupPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("group does not exist", "group id", mirrorGroupPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchGroup,
		}
	}

	groupInfo.SourceType = types.SOURCE_TYPE_ORIGIN
	app.storageKeeper.SetGroupInfo(ctx, groupInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorGroupResult{
		Status:      uint32(types.StatusFail),
		GroupName:   groupInfo.GroupName,
		GroupId:     groupInfo.Id,
		DestChainId: uint32(appCtx.SrcChainId),
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}
	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleMirrorGroupSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, synPackage *types.MirrorGroupSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received mirror group syn ack package ")

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleUpdateGroupMemberSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, updateGroupPackage *types.UpdateGroupMemberSynPackage) sdk.ExecuteResult {
	err := updateGroupPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.UpdateGroupMemberAckPackage{
				Status:    types.StatusFail,
				Operator:  updateGroupPackage.Operator,
				ExtraData: updateGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	groupInfo, found := app.storageKeeper.GetGroupInfoById(ctx, math.NewUintFromBigInt(updateGroupPackage.GroupId))
	if !found {
		return sdk.ExecuteResult{
			Payload: types.UpdateGroupMemberAckPackage{
				Status:    types.StatusFail,
				Operator:  updateGroupPackage.Operator,
				ExtraData: updateGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchGroup,
		}
	}

	options := types.UpdateGroupMemberOptions{
		SourceType: types.SOURCE_TYPE_BSC_CROSS_CHAIN,
	}
	if updateGroupPackage.OperationType == types.OperationAddGroupMember {
		options.MembersToAdd = updateGroupPackage.GetMembers()
	} else {
		options.MembersToDelete = updateGroupPackage.GetMembers()
	}

	err = app.storageKeeper.UpdateGroupMember(
		ctx,
		updateGroupPackage.Operator,
		groupInfo,
		options,
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.UpdateGroupMemberAckPackage{
				Status:    types.StatusFail,
				Operator:  updateGroupPackage.Operator,
				ExtraData: updateGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.UpdateGroupMemberAckPackage{
			Status:        types.StatusSuccess,
			Id:            groupInfo.Id.BigInt(),
			Operator:      updateGroupPackage.Operator,
			OperationType: updateGroupPackage.OperationType,
			Members:       updateGroupPackage.Members,
			ExtraData:     updateGroupPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *GroupApp) handleUpdateGroupMemberV2SynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, updateGroupPackage *types.UpdateGroupMemberV2SynPackage) sdk.ExecuteResult {
	err := updateGroupPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.UpdateGroupMemberAckPackage{
				Status:    types.StatusFail,
				Operator:  updateGroupPackage.Operator,
				ExtraData: updateGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	groupInfo, found := app.storageKeeper.GetGroupInfoById(ctx, math.NewUintFromBigInt(updateGroupPackage.GroupId))
	if !found {
		return sdk.ExecuteResult{
			Payload: types.UpdateGroupMemberAckPackage{
				Status:    types.StatusFail,
				Operator:  updateGroupPackage.Operator,
				ExtraData: updateGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchGroup,
		}
	}

	switch updateGroupPackage.OperationType {
	case types.OperationAddGroupMember, types.OperationDeleteGroupMember:
		err = app.handleAddOrDeleteGroupMemberOperation(ctx, groupInfo, updateGroupPackage)
	case types.OperationRenewGroupMember:
		err = app.handleRenewGroupOperation(ctx, groupInfo, updateGroupPackage)
	}

	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.UpdateGroupMemberAckPackage{
				Status:    types.StatusFail,
				Operator:  updateGroupPackage.Operator,
				ExtraData: updateGroupPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.UpdateGroupMemberAckPackage{
			Status:        types.StatusSuccess,
			Id:            groupInfo.Id.BigInt(),
			Operator:      updateGroupPackage.Operator,
			OperationType: updateGroupPackage.OperationType,
			Members:       updateGroupPackage.Members,
			ExtraData:     updateGroupPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *GroupApp) handleAddOrDeleteGroupMemberOperation(ctx sdk.Context, groupInfo *types.GroupInfo, updateGroupPackage *types.UpdateGroupMemberV2SynPackage) error {
	options := types.UpdateGroupMemberOptions{
		SourceType: types.SOURCE_TYPE_BSC_CROSS_CHAIN,
	}
	if updateGroupPackage.OperationType == types.OperationAddGroupMember {
		options.MembersToAdd = updateGroupPackage.GetMembers()
	} else {
		options.MembersToDelete = updateGroupPackage.GetMembers()
	}

	return app.storageKeeper.UpdateGroupMember(
		ctx,
		updateGroupPackage.Operator,
		groupInfo,
		options,
	)
}

func (app *GroupApp) handleRenewGroupOperation(ctx sdk.Context, groupInfo *types.GroupInfo, updateGroupPackage *types.UpdateGroupMemberV2SynPackage) error {
	options := types.RenewGroupMemberOptions{
		SourceType:        types.SOURCE_TYPE_BSC_CROSS_CHAIN,
		Members:           updateGroupPackage.GetMembers(),
		MembersExpiration: updateGroupPackage.GetMemberExpiration(),
	}

	return app.storageKeeper.RenewGroupMember(
		ctx,
		updateGroupPackage.Operator,
		groupInfo,
		options,
	)
}

func (app *GroupApp) handleUpdateGroupMemberAckPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, createGroupPackage *types.UpdateGroupMemberAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received update group member ack package ")

	return sdk.ExecuteResult{}
}

func (app *GroupApp) handleUpdateGroupMemberFailAckPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, createGroupPackage *types.UpdateGroupMemberSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received update group member fail ack package ")

	return sdk.ExecuteResult{}
}
