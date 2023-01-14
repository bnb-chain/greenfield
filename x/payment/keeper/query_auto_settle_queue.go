package keeper

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) AutoSettleQueueAll(c context.Context, req *types.QueryAllAutoSettleQueueRequest) (*types.QueryAllAutoSettleQueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var autoSettleQueues []types.AutoSettleQueue
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	autoSettleQueueStore := prefix.NewStore(store, types.KeyPrefix(types.AutoSettleQueueKeyPrefix))

	pageRes, err := query.Paginate(autoSettleQueueStore, req.Pagination, func(key []byte, value []byte) error {
		var autoSettleQueue types.AutoSettleQueue
		if err := k.cdc.Unmarshal(value, &autoSettleQueue); err != nil {
			return err
		}

		autoSettleQueues = append(autoSettleQueues, autoSettleQueue)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAutoSettleQueueResponse{AutoSettleQueue: autoSettleQueues, Pagination: pageRes}, nil
}

func (k Keeper) AutoSettleQueue(c context.Context, req *types.QueryGetAutoSettleQueueRequest) (*types.QueryGetAutoSettleQueueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetAutoSettleQueue(
		ctx,
		req.Timestamp,
		req.User,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetAutoSettleQueueResponse{AutoSettleQueue: val}, nil
}
