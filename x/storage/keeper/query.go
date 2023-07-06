package keeper

import (
	"context"

	"cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/internal/sequence"
	errors2 "github.com/bnb-chain/greenfield/types/errors"
	types2 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gnfd "github.com/bnb-chain/greenfield/types"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
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
			BucketInfo: bucketInfo,
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
			BucketInfo: bucketInfo,
		}, nil

	}
	return nil, types.ErrNoSuchBucket
}

func (k Keeper) HeadObject(goCtx context.Context, req *types.QueryHeadObjectRequest) (*types.QueryHeadObjectResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	bucketInfo, found := k.GetBucketInfo(ctx, req.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	objectInfo, objectFound := k.GetObjectInfo(ctx, req.BucketName, req.ObjectName)
	if !objectFound {
		return nil, types.ErrNoSuchObject
	}
	var gvg *types2.GlobalVirtualGroup
	if objectInfo.ObjectStatus == types.OBJECT_STATUS_SEALED {
		gvgFound := false
		gvg, gvgFound = k.GetObjectGVG(ctx, bucketInfo.Id, objectInfo.LocalVirtualGroupId)
		if !gvgFound {
			return nil, types.ErrInvalidGlobalVirtualGroup.Wrapf("gvg not found. objectInfo: %s", objectInfo.String())
		}
	}
	return &types.QueryHeadObjectResponse{
		ObjectInfo:         objectInfo,
		GlobalVirtualGroup: gvg,
	}, nil
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
	if !found {
		return nil, types.ErrNoSuchObject
	}

	bucketInfo, found := k.GetBucketInfo(ctx, objectInfo.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	var gvg *types2.GlobalVirtualGroup
	if objectInfo.ObjectStatus == types.OBJECT_STATUS_SEALED {
		gvgFound := false
		gvg, gvgFound = k.GetObjectGVG(ctx, bucketInfo.Id, objectInfo.LocalVirtualGroupId)
		if !gvgFound {
			return nil, types.ErrInvalidGlobalVirtualGroup.Wrapf("gvg not found. objectInfo: %s", objectInfo.String())
		}
	}
	return &types.QueryHeadObjectResponse{
		ObjectInfo:         objectInfo,
		GlobalVirtualGroup: gvg,
	}, nil

}

func (k Keeper) ListBuckets(goCtx context.Context, req *types.QueryListBucketsRequest) (*types.QueryListBucketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Pagination != nil && req.Pagination.Limit > types.MaxPaginationLimit {
		return nil, status.Errorf(codes.InvalidArgument, "exceed pagination limit %d", types.MaxPaginationLimit)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var bucketInfos []*types.BucketInfo
	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketByIDPrefix)

	pageRes, err := query.Paginate(bucketStore, req.Pagination, func(key []byte, value []byte) error {
		var bucketInfo types.BucketInfo
		k.cdc.MustUnmarshal(value, &bucketInfo)
		bucketInfos = append(bucketInfos, &bucketInfo)
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
	if req.Pagination != nil && req.Pagination.Limit > types.MaxPaginationLimit {
		return nil, status.Errorf(codes.InvalidArgument, "exceed pagination limit %d", types.MaxPaginationLimit)
	}
	ctx := sdk.UnwrapSDKContext(goCtx)

	var objectInfos []*types.ObjectInfo
	store := ctx.KVStore(k.storeKey)
	objectPrefixStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(req.BucketName))

	pageRes, err := query.Paginate(objectPrefixStore, req.Pagination, func(key []byte, value []byte) error {
		objectInfo, found := k.GetObjectInfoById(ctx, k.objectSeq.DecodeSequence(value))
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
	if req.Pagination != nil && req.Pagination.Limit > types.MaxPaginationLimit {
		return nil, status.Errorf(codes.InvalidArgument, "exceed pagination limit %d", types.MaxPaginationLimit)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var objectInfos []*types.ObjectInfo
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
		u256Seq := sequence.Sequence[math.Uint]{}
		objectInfo, found := k.GetObjectInfoById(ctx, u256Seq.DecodeSequence(value))
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

func (k Keeper) HeadBucketNFT(goCtx context.Context, req *types.QueryNFTRequest) (*types.QueryBucketNFTResponse, error) {
	id, err := validateAndGetId(req)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	bucketInfo, found := k.GetBucketInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchBucket
	}
	return &types.QueryBucketNFTResponse{
		MetaData: bucketInfo.ToNFTMetadata(),
	}, nil
}

func (k Keeper) HeadObjectNFT(goCtx context.Context, req *types.QueryNFTRequest) (*types.QueryObjectNFTResponse, error) {
	id, err := validateAndGetId(req)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	objectInfo, found := k.GetObjectInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchObject
	}
	return &types.QueryObjectNFTResponse{
		MetaData: objectInfo.ToNFTMetadata(),
	}, nil
}

func (k Keeper) HeadGroupNFT(goCtx context.Context, req *types.QueryNFTRequest) (*types.QueryGroupNFTResponse, error) {
	id, err := validateAndGetId(req)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	groupInfo, found := k.GetGroupInfoById(ctx, id)
	if !found {
		return nil, types.ErrNoSuchObject
	}
	return &types.QueryGroupNFTResponse{
		MetaData: groupInfo.ToNFTMetadata(),
	}, nil
}

func validateAndGetId(req *types.QueryNFTRequest) (math.Uint, error) {
	if req == nil {
		return math.ZeroUint(), status.Error(codes.InvalidArgument, "invalid request")
	}
	id, err := math.ParseUint(req.TokenId)
	if err != nil {
		return math.ZeroUint(), status.Error(codes.InvalidArgument, "invalid token id")
	}
	return id, nil
}

func (k Keeper) QueryPolicyForAccount(goCtx context.Context, req *types.QueryPolicyForAccountRequest) (*types.
	QueryPolicyForAccountResponse,
	error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	principalAcc, err := sdk.AccAddressFromHexUnsafe(req.PrincipalAddress)
	if err != nil {
		return nil, err
	}
	var grn gnfd.GRN
	err = grn.ParseFromString(req.Resource, false)
	if err != nil {
		return nil, err
	}

	policy, err := k.GetPolicy(ctx, &grn, permtypes.NewPrincipalWithAccount(principalAcc))
	if err != nil {
		return nil, err
	}

	return &types.QueryPolicyForAccountResponse{Policy: policy}, nil
}

func (k Keeper) QueryPolicyForGroup(goCtx context.Context, req *types.QueryPolicyForGroupRequest) (*types.
	QueryPolicyForGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	id, err := math.ParseUint(req.PrincipalGroupId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid group id")
	}

	var grn gnfd.GRN
	err = grn.ParseFromString(req.Resource, false)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to parse GRN %s: %v", req.Resource, err)
	}

	policy, err := k.GetPolicy(
		ctx, &grn, permtypes.NewPrincipalWithGroupId(id),
	)
	if err != nil {
		return nil, err
	}
	return &types.QueryPolicyForGroupResponse{Policy: policy}, nil
}

func (k Keeper) VerifyPermission(goCtx context.Context, req *types.QueryVerifyPermissionRequest) (*types.QueryVerifyPermissionResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	operator, err := sdk.AccAddressFromHexUnsafe(req.Operator)
	if err != nil && err != sdk.ErrEmptyHexAddress {
		return nil, errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid operator address (%s)", err)
	}

	if req.BucketName == "" {
		return nil, errors.Wrapf(errors2.ErrInvalidParameter, "No bucket specified")
	}

	bucketInfo, found := k.GetBucketInfo(ctx, req.BucketName)
	if !found {
		return nil, types.ErrNoSuchBucket
	}

	var effect permtypes.Effect
	if req.ObjectName == "" {
		effect = k.VerifyBucketPermission(ctx, bucketInfo, operator, req.ActionType, nil)
	} else {
		objectInfo, found := k.GetObjectInfo(ctx, req.BucketName, req.ObjectName)
		if !found {
			return nil, types.ErrNoSuchObject
		}
		effect = k.VerifyObjectPermission(ctx, bucketInfo, objectInfo, operator, req.ActionType)
	}

	return &types.QueryVerifyPermissionResponse{
		Effect: effect,
	}, nil
}

func (k Keeper) HeadGroup(goCtx context.Context, req *types.QueryHeadGroupRequest) (*types.QueryHeadGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromHexUnsafe(req.GroupOwner)
	if err != nil {
		return nil, err
	}
	groupInfo, found := k.GetGroupInfo(ctx, owner, req.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	return &types.QueryHeadGroupResponse{GroupInfo: groupInfo}, nil
}

func (k Keeper) ListGroup(goCtx context.Context, req *types.QueryListGroupRequest) (*types.QueryListGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	if req.Pagination != nil && req.Pagination.Limit > types.MaxPaginationLimit {
		return nil, status.Errorf(codes.InvalidArgument, "exceed pagination limit %d", types.MaxPaginationLimit)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	owner, err := sdk.AccAddressFromHexUnsafe(req.GroupOwner)
	if err != nil {
		return nil, err
	}

	var groupInfos []*types.GroupInfo
	store := ctx.KVStore(k.storeKey)
	groupStore := prefix.NewStore(store, types.GetGroupKeyOnlyOwnerPrefix(owner))

	pageRes, err := query.Paginate(groupStore, req.Pagination, func(key []byte, value []byte) error {
		groupInfo, found := k.GetGroupInfoById(ctx, k.groupSeq.DecodeSequence(value))
		if found {
			groupInfos = append(groupInfos, groupInfo)
		}
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryListGroupResponse{GroupInfos: groupInfos, Pagination: pageRes}, nil
}

func (k Keeper) HeadGroupMember(goCtx context.Context, req *types.QueryHeadGroupMemberRequest) (*types.QueryHeadGroupMemberResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	member, err := sdk.AccAddressFromHexUnsafe(req.Member)
	if err != nil {
		return nil, err
	}
	owner, err := sdk.AccAddressFromHexUnsafe(req.GroupOwner)
	if err != nil {
		return nil, err
	}
	groupInfo, found := k.GetGroupInfo(ctx, owner, req.GroupName)
	if !found {
		return nil, types.ErrNoSuchGroup
	}
	groupMember, found := k.permKeeper.GetGroupMember(ctx, groupInfo.Id, member)
	if !found {
		return nil, types.ErrNoSuchGroupMember
	}
	return &types.QueryHeadGroupMemberResponse{GroupMember: groupMember}, nil
}

func (k Keeper) QueryPolicyById(goCtx context.Context, req *types.QueryPolicyByIdRequest) (*types.
	QueryPolicyByIdResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	policyId, err := math.ParseUint(req.PolicyId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid policy id")
	}
	policy, found := k.permKeeper.GetPolicyByID(ctx, policyId)
	if !found {
		return nil, types.ErrNoSuchPolicy
	}
	return &types.QueryPolicyByIdResponse{Policy: policy}, nil
}
