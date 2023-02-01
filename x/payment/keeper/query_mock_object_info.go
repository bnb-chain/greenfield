package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MockObjectInfoAll(c context.Context, req *types.QueryAllMockObjectInfoRequest) (*types.QueryAllMockObjectInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var mockObjectInfos []types.MockObjectInfo
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	mockObjectInfoStore := prefix.NewStore(store, types.MockObjectInfoKeyPrefix)

	pageRes, err := query.Paginate(mockObjectInfoStore, req.Pagination, func(key []byte, value []byte) error {
		var mockObjectInfo types.MockObjectInfo
		if err := k.cdc.Unmarshal(value, &mockObjectInfo); err != nil {
			return err
		}

		mockObjectInfos = append(mockObjectInfos, mockObjectInfo)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMockObjectInfoResponse{MockObjectInfo: mockObjectInfos, Pagination: pageRes}, nil
}

func (k Keeper) MockObjectInfo(c context.Context, req *types.QueryGetMockObjectInfoRequest) (*types.QueryGetMockObjectInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMockObjectInfo(
		ctx,
		req.BucketName,
		req.ObjectName,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMockObjectInfoResponse{MockObjectInfo: val}, nil
}
