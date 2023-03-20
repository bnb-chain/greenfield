package keeper

import (
	"encoding/hex"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ sdk.CrossChainApplication = &ObjectApp{}

type ObjectApp struct {
	storageKeeper Keeper
}

func NewObjectApp(keeper Keeper) *ObjectApp {
	return &ObjectApp{
		storageKeeper: keeper,
	}
}

func (app *ObjectApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.ObjectChannelId, sdk.AckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize object cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize object cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult

	switch p := pack.(type) {
	case *types.MirrorObjectAckPackage:
		operationType = types.OperationMirrorObject
		result = app.handleMirrorObjectAckPackage(ctx, appCtx, p)
	case *types.DeleteObjectAckPackage:
		operationType = types.OperationDeleteObject
		result = app.handleDeleteObjectAckPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		wrapPayloadBts, err := rlp.EncodeToBytes(wrapPayload)
		if err != nil {
			panic(err)
		}
		result.Payload = wrapPayloadBts
	}

	return result
}

func (app *ObjectApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.ObjectChannelId, sdk.FailAckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize object cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize object cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorObjectSynPackage:
		operationType = types.OperationMirrorObject
		result = app.handleMirrorObjectFailAckPackage(ctx, appCtx, p)
	case *types.DeleteObjectSynPackage:
		operationType = types.OperationDeleteObject
		result = app.handleDeleteObjectFailAckPackage(ctx, appCtx, p)
	default:
		panic("unknown cross chain ack package type")
	}

	if len(result.Payload) != 0 {
		wrapPayload := types.CrossChainPackage{
			OperationType: operationType,
			Package:       result.Payload,
		}
		wrapPayloadBts, err := rlp.EncodeToBytes(wrapPayload)
		if err != nil {
			panic(err)
		}
		result.Payload = wrapPayloadBts
	}

	return result
}

func (app *ObjectApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.ObjectChannelId, sdk.SynCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize object cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize object cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorObjectSynPackage:
		operationType = types.OperationMirrorObject
		result = app.handleMirrorObjectSynPackage(ctx, appCtx, p)
	case *types.DeleteObjectSynPackage:
		operationType = types.OperationDeleteObject
		result = app.handleDeleteObjectSynPackage(ctx, appCtx, p)
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
		wrapPayloadBts, err := rlp.EncodeToBytes(wrapPayload)
		if err != nil {
			panic(err)
		}
		result.Payload = wrapPayloadBts
	}

	return result
}

func (app *ObjectApp) handleMirrorObjectAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.MirrorObjectAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received mirror object ack package ")

	objectInfo, found := app.storageKeeper.GetObjectInfoById(ctx, math.NewUintFromBigInt(ackPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("object does not exist", "object id", ackPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchObject,
		}
	}

	// update object
	if ackPackage.Status == types.StatusSuccess {
		objectInfo.SourceType = types.SOURCE_TYPE_BSC_CROSS_CHAIN

		app.storageKeeper.SetObjectInfo(ctx, objectInfo)
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorObjectResult{
		Status:     uint32(ackPackage.Status),
		BucketName: objectInfo.BucketName,
		ObjectName: objectInfo.ObjectName,
		ObjectId:   objectInfo.Id,
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *ObjectApp) handleMirrorObjectFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, mirrorObjectPackage *types.MirrorObjectSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received mirror object fail ack package ")

	objectInfo, found := app.storageKeeper.GetObjectInfoById(ctx, math.NewUintFromBigInt(mirrorObjectPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("object does not exist", "object id", mirrorObjectPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchObject,
		}
	}

	objectInfo.SourceType = types.SOURCE_TYPE_ORIGIN
	app.storageKeeper.SetObjectInfo(ctx, objectInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorObjectResult{
		Status:     uint32(types.StatusFail),
		BucketName: objectInfo.BucketName,
		ObjectName: objectInfo.ObjectName,
		ObjectId:   objectInfo.Id,
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *ObjectApp) handleMirrorObjectSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, synPackage *types.MirrorObjectSynPackage) sdk.ExecuteResult {
	return sdk.ExecuteResult{}
}

func (app *ObjectApp) handleDeleteObjectSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, deleteObjectPackage *types.DeleteObjectSynPackage) sdk.ExecuteResult {
	err := deleteObjectPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteObjectAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteObjectPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	app.storageKeeper.Logger(ctx).Info("process delete object syn package", "object id", deleteObjectPackage.Id.String())

	objectInfo, found := app.storageKeeper.GetObjectInfoById(ctx, math.NewUintFromBigInt(deleteObjectPackage.Id))
	if !found {
		return sdk.ExecuteResult{
			Payload: types.DeleteObjectAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteObjectPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchObject,
		}
	}

	err = app.storageKeeper.DeleteObject(ctx,
		deleteObjectPackage.Operator,
		objectInfo.BucketName,
		objectInfo.ObjectName,
		DeleteObjectOptions{
			SourceType: types.SOURCE_TYPE_BSC_CROSS_CHAIN,
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteObjectAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteObjectPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.DeleteObjectAckPackage{
			Status:    types.StatusSuccess,
			Id:        objectInfo.Id.BigInt(),
			ExtraData: deleteObjectPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *ObjectApp) handleDeleteObjectAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.DeleteObjectAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete object ack package ")

	return sdk.ExecuteResult{}
}

func (app *ObjectApp) handleDeleteObjectFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.DeleteObjectSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete object fail ack package ")

	return sdk.ExecuteResult{}
}
