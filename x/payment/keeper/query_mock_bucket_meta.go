package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/bnb-chain/bfs/x/payment/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) MockBucketMetaAll(c context.Context, req *types.QueryAllMockBucketMetaRequest) (*types.QueryAllMockBucketMetaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var mockBucketMetas []types.MockBucketMeta
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	mockBucketMetaStore := prefix.NewStore(store, types.KeyPrefix(types.MockBucketMetaKeyPrefix))

	pageRes, err := query.Paginate(mockBucketMetaStore, req.Pagination, func(key []byte, value []byte) error {
		var mockBucketMeta types.MockBucketMeta
		if err := k.cdc.Unmarshal(value, &mockBucketMeta); err != nil {
			return err
		}

		mockBucketMetas = append(mockBucketMetas, mockBucketMeta)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllMockBucketMetaResponse{MockBucketMeta: mockBucketMetas, Pagination: pageRes}, nil
}

func (k Keeper) MockBucketMeta(c context.Context, req *types.QueryGetMockBucketMetaRequest) (*types.QueryGetMockBucketMetaResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetMockBucketMeta(
	    ctx,
	    req.BucketName,
        )
	if !found {
	    return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetMockBucketMetaResponse{MockBucketMeta: val}, nil
}