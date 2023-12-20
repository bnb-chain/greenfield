package keeper

import (
	"encoding/json"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ sdk.CrossChainApplication = &PermissionApp{}

type PermissionApp struct {
	permissionKeeper types.PermissionKeeper
}

func NewPermissionApp(keeper types.PermissionKeeper) *PermissionApp {
	return &PermissionApp{
		permissionKeeper: keeper,
	}
}

func (app *PermissionApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.PermissionChannelId, sdk.AckCrossChainPackageType)
	if err != nil {
		panic("deserialize Policy cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.CreatePolicyAckPackage:
		operationType = types.OperationCreatePolicy
		result = app.handleCreatePolicyAckPackage(ctx, appCtx, p)
	case *types.DeletePolicyAckPackage:
		operationType = types.OperationDeletePolicy
		result = app.handleDeletePolicyAckPackage(ctx, appCtx, p)
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

func (app *PermissionApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.PermissionChannelId, sdk.FailAckCrossChainPackageType)
	if err != nil {
		panic("deserialize Policy cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.CreatePolicySynPackage:
		operationType = types.OperationCreatePolicy
		result = app.handleCreatePolicyFailAckPackage(ctx, appCtx, p)
	case *types.DeletePolicySynPackage:
		operationType = types.OperationDeletePolicy
		result = app.handleDeletePolicyFailAckPackage(ctx, appCtx, p)
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

func (app *PermissionApp) handleDeletePolicyAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.DeletePolicyAckPackage) sdk.ExecuteResult {
	return sdk.ExecuteResult{}
}

func (app *PermissionApp) handleDeletePolicyFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.DeletePolicySynPackage) sdk.ExecuteResult {
	return sdk.ExecuteResult{}
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

func (app *PermissionApp) handleCreatePolicyAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.CreatePolicyAckPackage) sdk.ExecuteResult {
	return sdk.ExecuteResult{}
}

func (app *PermissionApp) handleCreatePolicyFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.CreatePolicySynPackage) sdk.ExecuteResult {
	return sdk.ExecuteResult{}
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
