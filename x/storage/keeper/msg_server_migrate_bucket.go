package keeper

import (
	"context"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MigrateBucket(goCtx context.Context, msg *types.MsgMigrateBucket) (*types.MsgMigrateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operator := sdk.MustAccAddressFromHex(msg.Operator)

	bucketInfo, found := k.GetBucketInfo(ctx, msg.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	if bucketInfo.BucketStatus == types.BUCKET_STATUS_MIGRATING {
		return nil, types.ErrInvalidBucketStatus.Wrapf("The bucket already been migrating")

	}
	srcSP, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operator)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	dstSP, found := k.spKeeper.GetStorageProvider(ctx, msg.DstPrimarySpId)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("dst sp not found")
	}

	if srcSP.Status == sptypes.STATUS_IN_SERVICE || dstSP.Status == sptypes.STATUS_IN_SERVICE {
		return nil, sptypes.ErrStorageProviderNotInService.Wrapf(
			"origin SP status: %s, dst SP status: %s", srcSP.Status.String(), dstSP.Status.String())
	}

	// check approval
	if msg.DstPrimarySpApproval.ExpiredHeight < (uint64)(ctx.BlockHeight()) {
		return nil, types.ErrInvalidApproval.Wrap("dst primary sp approval timeout")
	}
	mgs := types.NewMigrationBucketSignDoc(srcSP.Id, dstSP.Id, bucketInfo.Id)
	err := k.VerifySPAndSignature(ctx, dstSP.Id, mgs.GetSignBytes(), msg.DstPrimarySpApproval.Sig)
	if err != nil {
		return nil, err
	}

	err = k.MigrationBucket(ctx, srcSP, dstSP, bucketInfo)
	if err != nil {
		return nil, err
	}

	bucketInfo.BucketStatus = types.BUCKET_STATUS_MIGRATING
	k.SetBucketInfo(ctx, bucketInfo)
	return &types.MsgMigrateBucketResponse{}, nil
}
