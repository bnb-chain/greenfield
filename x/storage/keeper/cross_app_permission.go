package keeper

import (
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	types2 "github.com/bnb-chain/greenfield/types"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	gnfdresource "github.com/bnb-chain/greenfield/types/resource"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ sdk.CrossChainApplication = &PermissionApp{}

type PermissionApp struct {
	storageKeeper    types.StorageKeeper
	permissionKeeper types.PermissionKeeper
}

func NewPermissionApp(keeper types.StorageKeeper, permissionKeeper types.PermissionKeeper) *PermissionApp {
	return &PermissionApp{
		storageKeeper:    keeper,
		permissionKeeper: permissionKeeper,
	}
}

func (app *PermissionApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	panic("invalid cross chain ack package")
}

func (app *PermissionApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	panic("invalid cross chain fail ack package")
}

func (app *PermissionApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.PermissionChannelId, sdk.SynCrossChainPackageType)
	if err != nil {
		panic("deserialize Policy cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.CreatePolicySynPackage:
		operationType = types.OperationCreatePolicy
		result = app.handleCreatePolicySynPackage(ctx, p)
	case *types.DeletePolicySynPackage:
		operationType = types.OperationDeletePolicy
		result = app.handleDeletePolicySynPackage(ctx, p)
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

func (app *PermissionApp) handleDeletePolicySynPackage(ctx sdk.Context, deletePolicyPackage *types.DeletePolicySynPackage) sdk.ExecuteResult {
	err := deletePolicyPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeletePolicyAckPackage{
				Status:    types.StatusFail,
				ExtraData: deletePolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	policy, found := app.permissionKeeper.GetPolicyByID(ctx, math.NewUintFromBigInt(deletePolicyPackage.Id))
	if !found {
		return sdk.ExecuteResult{
			Payload: types.DeletePolicyAckPackage{
				Status:    types.StatusFail,
				ExtraData: deletePolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchPolicy,
		}
	}

	resOwner, err := app.getResourceOwner(ctx, policy)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   deletePolicyPackage.Operator,
				ExtraData: deletePolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	if !deletePolicyPackage.Operator.Equals(resOwner) {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   deletePolicyPackage.Operator,
				ExtraData: deletePolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrAccessDenied.Wrapf(
				"Only resource owner can delete bucket policy, operator (%s), owner(%s)",
				deletePolicyPackage.Operator.String(), resOwner.String()),
		}
	}

	_, err = app.permissionKeeper.DeletePolicy(ctx, policy.Principal, policy.ResourceType, policy.ResourceId)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeletePolicyAckPackage{
				Status:    types.StatusFail,
				ExtraData: deletePolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.DeletePolicyAckPackage{
			Status:    types.StatusSuccess,
			Id:        policy.Id.BigInt(),
			ExtraData: deletePolicyPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *PermissionApp) handleCreatePolicySynPackage(ctx sdk.Context, createPolicyPackage *types.CreatePolicySynPackage) sdk.ExecuteResult {
	err := createPolicyPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   createPolicyPackage.Operator,
				ExtraData: createPolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	var policy permtypes.Policy
	var grn types2.GRN
	err = policy.Unmarshal(createPolicyPackage.Data)
	if err != nil {
		var msgPutPolicy types.MsgPutPolicy
		err = msgPutPolicy.Unmarshal(createPolicyPackage.Data)
		if err != nil {
			return sdk.ExecuteResult{
				Payload: types.CreatePolicyAckPackage{
					Status:    types.StatusFail,
					Creator:   createPolicyPackage.Operator,
					ExtraData: createPolicyPackage.ExtraData,
				}.MustSerialize(),
				Err: err,
			}
		}
		if msgPutPolicy.GetOperator() != createPolicyPackage.Operator.String() {
			return sdk.ExecuteResult{
				Payload: types.CreatePolicyAckPackage{
					Status:    types.StatusFail,
					Creator:   createPolicyPackage.Operator,
					ExtraData: createPolicyPackage.ExtraData,
				}.MustSerialize(),
				Err: types.ErrAccessDenied.Wrapf(
					"Only resource owner can put policy, operator (%s)",
					msgPutPolicy.GetOperator()),
			}
		}

		err = grn.ParseFromString(msgPutPolicy.Resource, true)
		if err != nil {
			return sdk.ExecuteResult{
				Payload: types.CreatePolicyAckPackage{
					Status:    types.StatusFail,
					Creator:   createPolicyPackage.Operator,
					ExtraData: createPolicyPackage.ExtraData,
				}.MustSerialize(),
				Err: err,
			}
		}

		if msgPutPolicy.ExpirationTime != nil && msgPutPolicy.ExpirationTime.Before(ctx.BlockTime()) {
			return sdk.ExecuteResult{
				Payload: types.CreatePolicyAckPackage{
					Status:    types.StatusFail,
					Creator:   createPolicyPackage.Operator,
					ExtraData: createPolicyPackage.ExtraData,
				}.MustSerialize(),
				Err: permtypes.ErrPermissionExpired.Wrapf("The specified policy expiration time is less than the current block time, block time: %s", ctx.BlockTime().String()),
			}
		}

		for _, s := range msgPutPolicy.Statements {
			if s.ExpirationTime != nil && s.ExpirationTime.Before(ctx.BlockTime()) {
				return sdk.ExecuteResult{
					Payload: types.CreatePolicyAckPackage{
						Status:    types.StatusFail,
						Creator:   createPolicyPackage.Operator,
						ExtraData: createPolicyPackage.ExtraData,
					}.MustSerialize(),
					Err: permtypes.ErrPermissionExpired.Wrapf("The specified statement expiration time is less than the current block time, block time: %s", ctx.BlockTime().String()),
				}
			}
		}

		policy = permtypes.Policy{
			ResourceType:   grn.ResourceType(),
			Principal:      msgPutPolicy.Principal,
			Statements:     msgPutPolicy.Statements,
			ExpirationTime: msgPutPolicy.ExpirationTime,
		}
		_, resID, err := app.getResourceOwnerAndIdFromGRN(ctx, grn)
		if err != nil {
			return sdk.ExecuteResult{
				Payload: types.CreatePolicyAckPackage{
					Status:    types.StatusFail,
					Creator:   createPolicyPackage.Operator,
					ExtraData: createPolicyPackage.ExtraData,
				}.MustSerialize(),
				Err: err,
			}
		}
		policy.ResourceId = resID
	}

	resOwner, err := app.getResourceOwner(ctx, &policy)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   createPolicyPackage.Operator,
				ExtraData: createPolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	if !createPolicyPackage.Operator.Equals(resOwner) {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   createPolicyPackage.Operator,
				ExtraData: createPolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrAccessDenied.Wrapf(
				"Only resource owner can put policy, operator (%s), owner(%s)",
				createPolicyPackage.Operator.String(), resOwner.String()),
		}
	}

	app.storageKeeper.NormalizePrincipal(ctx, policy.Principal)
	err = app.storageKeeper.ValidatePrincipal(ctx, resOwner, policy.Principal)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   createPolicyPackage.Operator,
				ExtraData: createPolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	PolicyId, err := app.permissionKeeper.PutPolicy(ctx, &policy)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreatePolicyAckPackage{
				Status:    types.StatusFail,
				Creator:   createPolicyPackage.Operator,
				ExtraData: createPolicyPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.CreatePolicyAckPackage{
			Status:    types.StatusSuccess,
			Id:        PolicyId.BigInt(),
			Creator:   createPolicyPackage.Operator,
			ExtraData: createPolicyPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *PermissionApp) getResourceOwner(ctx sdk.Context, policy *permtypes.Policy) (resOwner sdk.AccAddress, err error) {
	switch policy.ResourceType {
	case gnfdresource.RESOURCE_TYPE_BUCKET:
		bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, policy.ResourceId)
		if !found {
			return resOwner, types.ErrNoSuchBucket
		}
		resOwner = sdk.MustAccAddressFromHex(bucketInfo.Owner)
	case gnfdresource.RESOURCE_TYPE_OBJECT:
		objectInfo, found := app.storageKeeper.GetObjectInfoById(ctx, policy.ResourceId)
		if !found {
			return resOwner, types.ErrNoSuchObject
		}
		resOwner = sdk.MustAccAddressFromHex(objectInfo.Owner)
	case gnfdresource.RESOURCE_TYPE_GROUP:
		groupInfo, found := app.storageKeeper.GetGroupInfoById(ctx, policy.ResourceId)
		if !found {
			return resOwner, types.ErrNoSuchGroup
		}
		resOwner = sdk.MustAccAddressFromHex(groupInfo.Owner)
	default:
		return resOwner, gnfderrors.ErrInvalidGRN.Wrap("Unknown resource type in greenfield resource name")
	}
	return resOwner, nil
}

func (app *PermissionApp) getResourceOwnerAndIdFromGRN(ctx sdk.Context, grn types2.GRN) (resOwner sdk.AccAddress, resID math.Uint, err error) {
	switch grn.ResourceType() {
	case gnfdresource.RESOURCE_TYPE_BUCKET:
		bucketName, grnErr := grn.GetBucketName()
		if grnErr != nil {
			return resOwner, resID, grnErr
		}
		bucketInfo, found := app.storageKeeper.GetBucketInfo(ctx, bucketName)
		if !found {
			return resOwner, resID, types.ErrNoSuchBucket.Wrapf("bucketName: %s", bucketName)
		}
		resOwner = sdk.MustAccAddressFromHex(bucketInfo.Owner)
		resID = bucketInfo.Id
	case gnfdresource.RESOURCE_TYPE_OBJECT:
		bucketName, objectName, grnErr := grn.GetBucketAndObjectName()
		if grnErr != nil {
			return resOwner, resID, grnErr
		}
		objectInfo, found := app.storageKeeper.GetObjectInfo(ctx, bucketName, objectName)
		if !found {
			return resOwner, resID, types.ErrNoSuchObject.Wrapf("BucketName: %s, objectName: %s", bucketName, objectName)
		}
		resOwner = sdk.MustAccAddressFromHex(objectInfo.Owner)
		resID = objectInfo.Id
	case gnfdresource.RESOURCE_TYPE_GROUP:
		groupOwner, groupName, grnErr := grn.GetGroupOwnerAndAccount()
		if grnErr != nil {
			return resOwner, resID, grnErr
		}
		groupInfo, found := app.storageKeeper.GetGroupInfo(ctx, groupOwner, groupName)
		if !found {
			return resOwner, resID, types.ErrNoSuchBucket.Wrapf("groupOwner: %s, groupName: %s", groupOwner.String(), groupName)
		}
		resOwner = sdk.MustAccAddressFromHex(groupInfo.Owner)
		resID = groupInfo.Id
	default:
		return resOwner, resID, gnfderrors.ErrInvalidGRN.Wrap("Unknown resource type in greenfield resource name")
	}

	return resOwner, resID, nil
}
