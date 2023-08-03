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

func (k Keeper) StreamRecords(c context.Context, req *types.QueryStreamRecordsRequest) (*types.QueryStreamRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if err := query.CheckOffsetQueryNotAllowed(ctx, req.Pagination); err != nil {
		return nil, err
	}

	var streamRecords []types.StreamRecord
	store := ctx.KVStore(k.storeKey)
	streamRecordStore := prefix.NewStore(store, types.StreamRecordKeyPrefix)

	pageRes, err := query.Paginate(streamRecordStore, req.Pagination, func(key []byte, value []byte) error {
		var streamRecord types.StreamRecord
		if err := k.cdc.Unmarshal(value, &streamRecord); err != nil {
			return err
		}
		streamRecord.Account = sdk.AccAddress(key).String()

		streamRecords = append(streamRecords, streamRecord)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryStreamRecordsResponse{StreamRecords: streamRecords, Pagination: pageRes}, nil
}

func (k Keeper) StreamRecord(c context.Context, req *types.QueryGetStreamRecordRequest) (*types.QueryGetStreamRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	account, err := sdk.AccAddressFromHexUnsafe(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account")
	}
	val, found := k.GetStreamRecord(
		ctx,
		account,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetStreamRecordResponse{StreamRecord: *val}, nil
}
