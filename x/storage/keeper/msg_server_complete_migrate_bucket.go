package keeper

import (
	"context"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CompleteMigrateBucket(goCtx context.Context, msg *types.MsgCompleteMigrateBucket) (*types.MsgCompleteMigrateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	bucketInfo, found := k.GetBucketInfo(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	dstSP, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operator)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("dst SP not found.")
	}

	if bucketInfo.BucketStatus != types.BUCKET_STATUS_MIGRATING {
		return nil, types.ErrInvalidBucketStatus.Wrapf("The bucket not been migrating")
	}

	migrationBucketInfo, found := k.GetMigrationBucketInfo(ctx, bucketInfo.Id)
	if !found {
		return nil, types.ErrMigtationBucketFailed.Wrapf("migration bucket info not found.")
	}

	if dstSP.Id != migrationBucketInfo.DstSpId {
		return nil, types.ErrMigtationBucketFailed.Wrapf("dst sp info not match")
	}

	_, found = k.virtualGroupKeeper.GetGVGFamily(ctx, dstSP.Id, msg.GlobalVirtualGroupFamilyId)
	if !found {
		return nil, virtualgroupmoduletypes.ErrGVGFamilyNotExist
	}

	oldBucketInfo := &types.BucketInfo{
		PaymentAddress:             bucketInfo.PaymentAddress,
		PrimarySpId:                bucketInfo.PrimarySpId,
		GlobalVirtualGroupFamilyId: bucketInfo.GlobalVirtualGroupFamilyId,
		ChargedReadQuota:           bucketInfo.ChargedReadQuota,
		BillingInfo:                bucketInfo.BillingInfo,
	}

	bucketInfo.PrimarySpId = migrationBucketInfo.DstSpId
	bucketInfo.GlobalVirtualGroupFamilyId = msg.GlobalVirtualGroupFamilyId
	k.SetBucketInfo(ctx, bucketInfo)
	k.DeleteMigrationBucketInfo(ctx, bucketInfo.Id)

	err := k.ChargeBucketMigration(ctx, oldBucketInfo, bucketInfo)
	if err != nil {
		return nil, types.ErrMigtationBucketFailed.Wrapf("update payment info failed.")
	}

	return &types.MsgCompleteMigrateBucketResponse{}, nil
}
