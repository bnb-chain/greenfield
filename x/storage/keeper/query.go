package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) HeadBucket(goCtx context.Context, req *types.QueryHeadBucketRequest) (*types.QueryHeadBucketResponse, error) {
	return k.headBucketHelper(goCtx, req)
}

func (k Keeper) HeadBucketById(goCtx context.Context, req *types.QueryHeadBucketRequest) (*types.QueryHeadBucketResponse, error) {
	return k.headBucketHelper(goCtx, req)
}

func (k Keeper) headBucketHelper(goCtx context.Context, req *types.QueryHeadBucketRequest) (*types.QueryHeadBucketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	var bucketInfo types.BucketInfo
	var found bool

	if req.BucketId != "" {
		id, err := math.ParseUint(req.BucketId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid bucket id")
		}
		bucketInfo, found = k.GetBucketInfoById(ctx, id)
	} else {
		if req.BucketName == "" {
			return nil, status.Error(codes.InvalidArgument, "invalid request, both bucket name and bucket id are empty")
		}
		bucketInfo, found = k.GetBucketInfo(ctx, req.BucketName)
	}

	if found {
		return &types.QueryHeadBucketResponse{
			BucketInfo: &bucketInfo,
		}, nil
	}
	return nil, types.ErrNoSuchBucket
}

func (k Keeper) HeadObject(goCtx context.Context, req *types.QueryHeadObjectRequest) (*types.QueryHeadObjectResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	objectInfo, found := k.GetObjectInfo(ctx, req.BucketName, req.ObjectName)
	if found {
		return &types.QueryHeadObjectResponse{
			ObjectInfo: &objectInfo,
		}, nil
	}
	return nil, types.ErrNoSuchObject
}

func (k Keeper) HeadObjectById(goCtx context.Context, req *types.QueryHeadObjectByIdRequest) (*types.QueryHeadObjectResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	id, err := math.ParseUint(req.ObjectId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid object id")
	}

	objectInfo, found := k.GetObjectInfoById(ctx, id)
	if found {
		return &types.QueryHeadObjectResponse{
			ObjectInfo: &objectInfo,
		}, nil
	}
	return nil, types.ErrNoSuchObject
}

func (k Keeper) ListBuckets(goCtx context.Context, req *types.QueryListBucketsRequest) (*types.QueryListBucketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var bucketInfos []types.BucketInfo
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketPrefix)

	pageRes, err := query.Paginate(bucketStore, req.Pagination, func(key []byte, value []byte) error {
		var bucketInfo types.BucketInfo
		k.cdc.MustUnmarshal(value, &bucketInfo)
		bucketInfos = append(bucketInfos, bucketInfo)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListBucketsResponse{BucketInfos: bucketInfos, Pagination: pageRes}, nil
}

func (k Keeper) ListObjects(goCtx context.Context, req *types.QueryListObjectsRequest) (*types.QueryListObjectsResponse, error) {
	return k.listObjectHelper(
		getObjectPrefixStore,
		goCtx,
		types.GetBucketKey(req.BucketName),
		req,
	)
}

func (k Keeper) ListObjectsByBucketId(goCtx context.Context, req *types.QueryListObjectsRequest) (*types.QueryListObjectsResponse, error) {
	id, err := math.ParseUint(req.BucketId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid bucket id")
	}
	return k.listObjectHelper(
		getObjectPrefixStore,
		goCtx,
		types.GetBucketByIDKey(id),
		req,
	)
}

func (k Keeper) listObjectHelper(objectPrefixStoreFun func(s sdk.KVStore, b []byte) sdk.KVStore,
	goCtx context.Context,
	bucketStoreKey []byte,
	req *types.QueryListObjectsRequest) (*types.QueryListObjectsResponse, error) {

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var objectInfos []types.ObjectInfo
	store := ctx.KVStore(k.storeKey)
	objectStore := prefix.NewStore(store, types.ObjectPrefix)

	pageRes, err := query.Paginate(objectPrefixStoreFun(objectStore, bucketStoreKey), req.Pagination, func(key []byte, value []byte) error {
		var objectInfo types.ObjectInfo
		if err := k.cdc.Unmarshal(value, &objectInfo); err != nil {
			return err
		}
		objectInfos = append(objectInfos, objectInfo)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryListObjectsResponse{ObjectInfos: objectInfos, Pagination: pageRes}, nil
}

func getObjectPrefixStore(store sdk.KVStore, bz []byte) sdk.KVStore {
	return prefix.NewStore(store, bz)
}
