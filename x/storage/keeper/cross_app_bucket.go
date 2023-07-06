package keeper

import (
	"encoding/hex"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ sdk.CrossChainApplication = &BucketApp{}

type BucketApp struct {
	storageKeeper Keeper
}

func NewBucketApp(keeper Keeper) *BucketApp {
	return &BucketApp{
		storageKeeper: keeper,
	}
}

func (app *BucketApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.BucketChannelId, sdk.AckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize bucket cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize bucket cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorBucketAckPackage:
		operationType = types.OperationMirrorBucket
		result = app.handleMirrorBucketAckPackage(ctx, appCtx, p)
	case *types.CreateBucketAckPackage:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketAckPackage(ctx, appCtx, p)
	case *types.DeleteBucketAckPackage:
		operationType = types.OperationDeleteBucket
		result = app.handleDeleteBucketAckPackage(ctx, appCtx, p)
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

func (app *BucketApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.BucketChannelId, sdk.FailAckCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize bucket cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize bucket cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorBucketSynPackage:
		operationType = types.OperationMirrorBucket
		result = app.handleMirrorBucketFailAckPackage(ctx, appCtx, p)
	case *types.CreateBucketSynPackage:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketFailAckPackage(ctx, appCtx, p)
	case *types.DeleteBucketSynPackage:
		operationType = types.OperationDeleteBucket
		result = app.handleDeleteBucketFailAckPackage(ctx, appCtx, p)
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

func (app *BucketApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	pack, err := types.DeserializeCrossChainPackage(payload, types.BucketChannelId, sdk.SynCrossChainPackageType)
	if err != nil {
		app.storageKeeper.Logger(ctx).Error("deserialize bucket cross chain package error", "payload", hex.EncodeToString(payload), "error", err.Error())
		panic("deserialize bucket cross chain package error")
	}

	var operationType uint8
	var result sdk.ExecuteResult
	switch p := pack.(type) {
	case *types.MirrorBucketSynPackage:
		operationType = types.OperationMirrorBucket
		result = app.handleMirrorBucketSynPackage(ctx, appCtx, p)
	case *types.CreateBucketSynPackage:
		operationType = types.OperationCreateBucket
		result = app.handleCreateBucketSynPackage(ctx, appCtx, p)
	case *types.DeleteBucketSynPackage:
		operationType = types.OperationDeleteBucket
		result = app.handleDeleteBucketSynPackage(ctx, appCtx, p)
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

func (app *BucketApp) handleMirrorBucketAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.MirrorBucketAckPackage) sdk.ExecuteResult {
	bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, math.NewUintFromBigInt(ackPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("bucket does not exist", "bucket id", ackPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchBucket,
		}
	}

	// update bucket
	if ackPackage.Status == types.StatusSuccess {
		bucketInfo.SourceType = types.SOURCE_TYPE_BSC_CROSS_CHAIN

		app.storageKeeper.SetBucketInfo(ctx, bucketInfo)
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorBucketResult{
		Status:     uint32(ackPackage.Status),
		BucketName: bucketInfo.BucketName,
		BucketId:   bucketInfo.Id,
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleMirrorBucketFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, mirrorBucketPackage *types.MirrorBucketSynPackage) sdk.ExecuteResult {
	bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, math.NewUintFromBigInt(mirrorBucketPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("bucket does not exist", "bucket id", mirrorBucketPackage.Id.String())
		return sdk.ExecuteResult{
			Err: types.ErrNoSuchBucket,
		}
	}

	bucketInfo.SourceType = types.SOURCE_TYPE_ORIGIN
	app.storageKeeper.SetBucketInfo(ctx, bucketInfo)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventMirrorBucketResult{
		Status:     uint32(types.StatusFail),
		BucketName: bucketInfo.BucketName,
		BucketId:   bucketInfo.Id,
	}); err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleMirrorBucketSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, synPackage *types.MirrorBucketSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received mirror bucket syn package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.CreateBucketAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create bucket ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.CreateBucketSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received create bucket fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleCreateBucketSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, createBucketPackage *types.CreateBucketSynPackage) sdk.ExecuteResult {
	err := createBucketPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateBucketAckPackage{
				Status:    types.StatusFail,
				Creator:   createBucketPackage.Creator,
				ExtraData: createBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}
	app.storageKeeper.Logger(ctx).Info("process create bucket syn package", "bucket name", createBucketPackage.BucketName)

	bucketId, err := app.storageKeeper.CreateBucket(ctx,
		createBucketPackage.Creator,
		createBucketPackage.BucketName,
		createBucketPackage.PrimarySpAddress,
		&CreateBucketOptions{
			Visibility:       types.VisibilityType(createBucketPackage.Visibility),
			SourceType:       types.SOURCE_TYPE_BSC_CROSS_CHAIN,
			ChargedReadQuota: createBucketPackage.ChargedReadQuota,
			PaymentAddress:   createBucketPackage.PaymentAddress.String(),
			PrimarySpApproval: &types.Approval{
				ExpiredHeight: createBucketPackage.PrimarySpApprovalExpiredHeight,
				Sig:           createBucketPackage.PrimarySpApprovalSignature,
			},
			ApprovalMsgBytes: createBucketPackage.GetApprovalBytes(),
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.CreateBucketAckPackage{
				Status:    types.StatusFail,
				Creator:   createBucketPackage.Creator,
				ExtraData: createBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	return sdk.ExecuteResult{
		Payload: types.CreateBucketAckPackage{
			Status:    types.StatusSuccess,
			Id:        bucketId.BigInt(),
			Creator:   createBucketPackage.Creator,
			ExtraData: createBucketPackage.ExtraData,
		}.MustSerialize(),
	}
}

func (app *BucketApp) handleDeleteBucketAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, ackPackage *types.DeleteBucketAckPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete bucket ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleDeleteBucketFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, synPackage *types.DeleteBucketSynPackage) sdk.ExecuteResult {
	app.storageKeeper.Logger(ctx).Error("received delete bucket fail ack package ")

	return sdk.ExecuteResult{}
}

func (app *BucketApp) handleDeleteBucketSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, deleteBucketPackage *types.DeleteBucketSynPackage) sdk.ExecuteResult {
	err := deleteBucketPackage.ValidateBasic()
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteBucketAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}

	app.storageKeeper.Logger(ctx).Info("process delete group syn package", "bucket id", deleteBucketPackage.Id.String())

	bucketInfo, found := app.storageKeeper.GetBucketInfoById(ctx, math.NewUintFromBigInt(deleteBucketPackage.Id))
	if !found {
		app.storageKeeper.Logger(ctx).Error("bucket does not exist", "bucket id", deleteBucketPackage.Id.String())
		return sdk.ExecuteResult{
			Payload: types.DeleteBucketAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: types.ErrNoSuchBucket,
		}
	}

	err = app.storageKeeper.DeleteBucket(ctx,
		deleteBucketPackage.Operator,
		bucketInfo.BucketName,
		DeleteBucketOptions{
			SourceType: types.SOURCE_TYPE_BSC_CROSS_CHAIN,
		},
	)
	if err != nil {
		return sdk.ExecuteResult{
			Payload: types.DeleteBucketAckPackage{
				Status:    types.StatusFail,
				ExtraData: deleteBucketPackage.ExtraData,
			}.MustSerialize(),
			Err: err,
		}
	}
	return sdk.ExecuteResult{
		Payload: types.DeleteBucketAckPackage{
			Status:    types.StatusSuccess,
			Id:        bucketInfo.Id.BigInt(),
			ExtraData: deleteBucketPackage.ExtraData,
		}.MustSerialize(),
	}
}
