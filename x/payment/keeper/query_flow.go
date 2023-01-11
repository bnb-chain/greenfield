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

func (k Keeper) FlowAll(c context.Context, req *types.QueryAllFlowRequest) (*types.QueryAllFlowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var flows []types.Flow
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	flowStore := prefix.NewStore(store, types.KeyPrefix(types.FlowKeyPrefix))

	pageRes, err := query.Paginate(flowStore, req.Pagination, func(key []byte, value []byte) error {
		var flow types.Flow
		if err := k.cdc.Unmarshal(value, &flow); err != nil {
			return err
		}

		flows = append(flows, flow)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllFlowResponse{Flow: flows, Pagination: pageRes}, nil
}

func (k Keeper) Flow(c context.Context, req *types.QueryGetFlowRequest) (*types.QueryGetFlowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetFlow(
	    ctx,
	    req.From,
        req.To,
        )
	if !found {
	    return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetFlowResponse{Flow: val}, nil
}