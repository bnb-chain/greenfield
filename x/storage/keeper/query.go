package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) Bucket(goCtx context.Context, req *types.QueryBucketRequest) (*types.QueryBucketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

  bucketInfo, found := k.GetBucket(ctx, types.GetBucketKey(req.BucketName))
  if found {
    return &types.QueryBucketResponse {
      BucketInfo: &bucketInfo,
    }, nil

  } else {
    return nil, types.ErrNoSuchBucket
  }
}

func (k Keeper) Object(goCtx context.Context, req *types.QueryObjectRequest) (*types.QueryObjectResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

  objectInfo, found := k.GetObject(ctx, types.GetObjectKey(req.BucketName, req.ObjectName))
  if found {
    return &types.QueryObjectResponse{
      ObjectInfo: &objectInfo,
    }, nil
  } else {
    return nil, types.ErrNoSuchObject
  }
}
