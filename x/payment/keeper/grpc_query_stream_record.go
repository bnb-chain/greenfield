package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) StreamRecordAll(c context.Context, req *types.QueryAllStreamRecordRequest) (*types.QueryAllStreamRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var streamRecords []types.StreamRecord
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	streamRecordStore := prefix.NewStore(store, types.StreamRecordKeyPrefix)

	pageRes, err := query.Paginate(streamRecordStore, req.Pagination, func(key []byte, value []byte) error {
		var streamRecord types.StreamRecord
		if err := k.cdc.Unmarshal(value, &streamRecord); err != nil {
			return err
		}

		streamRecords = append(streamRecords, streamRecord)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllStreamRecordResponse{StreamRecord: streamRecords, Pagination: pageRes}, nil
}

func (k Keeper) StreamRecord(c context.Context, req *types.QueryGetStreamRecordRequest) (*types.QueryGetStreamRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetStreamRecord(
		ctx,
		req.Account,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetStreamRecordResponse{StreamRecord: val}, nil
}
