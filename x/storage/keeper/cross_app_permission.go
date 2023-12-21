package keeper

import (
	"encoding/json"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

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
	err = json.Unmarshal(createPolicyPackage.Data, &policy)
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
