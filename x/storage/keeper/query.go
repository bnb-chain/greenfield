package keeper

import (
	"context"
	"fmt"
	"reflect"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (k Keeper) HeadBucket(goCtx context.Context, req *types.QueryHeadBucketRequest) (*types.QueryHeadBucketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	bucketInfo, found := k.GetBucketInfo(ctx, req.BucketName)
	if found {
		return &types.QueryHeadBucketResponse{
			BucketInfo: &bucketInfo,
		}, nil

	}
	return nil, types.ErrNoSuchBucket
}

func (k Keeper) HeadBucketById(goCtx context.Context, req *types.QueryHeadBucketByIdRequest) (*types.QueryHeadBucketResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	id, err := math.ParseUint(req.BucketId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid bucket id")
	}

	bucketInfo, found := k.GetBucketInfoById(ctx, id)
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
	bucketStore := prefix.NewStore(store, types.BucketByIDPrefix)

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
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var objectInfos []types.ObjectInfo
	store := ctx.KVStore(k.storeKey)
	objectPrefixStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(req.BucketName))

	pageRes, err := query.Paginate(objectPrefixStore, req.Pagination, func(key []byte, value []byte) error {
		objectInfo, found := k.GetObjectInfoById(ctx, types.DecodeSequence(value))
		if found {
			objectInfos = append(objectInfos, objectInfo)
		}
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryListObjectsResponse{ObjectInfos: objectInfos, Pagination: pageRes}, nil
}

func (k Keeper) ListObjectsByBucketId(goCtx context.Context, req *types.QueryListObjectsByBucketIdRequest) (*types.QueryListObjectsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var objectInfos []types.ObjectInfo
	store := ctx.KVStore(k.storeKey)
	id, err := math.ParseUint(req.BucketId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid bucket id")
	}
	bucketInfo, found := k.GetBucketInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	objectPrefixStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(bucketInfo.BucketName))

	pageRes, err := query.Paginate(objectPrefixStore, req.Pagination, func(key []byte, value []byte) error {
		objectInfo, found := k.GetObjectInfoById(ctx, types.DecodeSequence(value))
		if found {
			objectInfos = append(objectInfos, objectInfo)
		}
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryListObjectsResponse{ObjectInfos: objectInfos, Pagination: pageRes}, nil
}

func (k Keeper) HeadBucketNFT(goCtx context.Context, req *types.QueryNFTRequest) (*types.QueryNFTResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	id, err := math.ParseUint(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token id")
	}
	bucketInfo, found := k.GetBucketInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	attributes := make([]types.Trait, 0)
	v := reflect.ValueOf(bucketInfo)
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typ.Field(i).Name
		if fieldName == "PaymentOutFlows" {
			continue
		}
		value := fmt.Sprintf("%v", field.Interface())
		attributes = append(attributes,
			types.Trait{
				TraitType: fieldName,
				Value:     value,
			})
	}
	md := types.MetaData{
		Name: &types.MetaData_BucketName{
			BucketName: bucketInfo.BucketName,
		},
		Attributes: attributes,
	}
	return &types.QueryNFTResponse{
		MetaData: &md,
	}, nil
}

func (k Keeper) HeadObjectNFT(goCtx context.Context, req *types.QueryNFTRequest) (*types.QueryNFTResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	id, err := math.ParseUint(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token id")
	}
	objectInfo, found := k.GetObjectInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchObject
	}
	attributes := make([]types.Trait, 0)
	v := reflect.ValueOf(objectInfo)
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typ.Field(i).Name
		if fieldName == "Checksums" {
			continue
		}
		value := fmt.Sprintf("%v", field.Interface())
		attributes = append(attributes,
			types.Trait{
				TraitType: fieldName,
				Value:     value,
			})
	}
	md := types.MetaData{
		Name: &types.MetaData_ObjectName{
			ObjectName: objectInfo.ObjectName,
		},
		Attributes: attributes,
	}
	return &types.QueryNFTResponse{
		MetaData: &md,
	}, nil
}

func (k Keeper) HeadGroupNFT(goCtx context.Context, req *types.QueryNFTRequest) (*types.QueryNFTResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	id, err := math.ParseUint(req.TokenId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid token id")
	}
	groupInfo, found := k.GetGroupInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchObject
	}
	attributes := make([]types.Trait, 0)
	v := reflect.ValueOf(groupInfo)
	typ := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldName := typ.Field(i).Name
		value := fmt.Sprintf("%v", field.Interface())
		attributes = append(attributes,
			types.Trait{
				TraitType: fieldName,
				Value:     value,
			})
	}
	md := types.MetaData{
		Name: &types.MetaData_GroupName{
			GroupName: groupInfo.GroupName,
		},
		Attributes: attributes,
	}
	return &types.QueryNFTResponse{
		MetaData: &md,
	}, nil
}
