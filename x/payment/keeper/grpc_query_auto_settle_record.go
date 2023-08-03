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

func (k Keeper) AutoSettleRecords(c context.Context, req *types.QueryAutoSettleRecordsRequest) (*types.QueryAutoSettleRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	if err := query.CheckOffsetQueryNotAllowed(ctx, req.Pagination); err != nil {
		return nil, err
	}

	var autoSettleRecords []types.AutoSettleRecord
	store := ctx.KVStore(k.storeKey)
	autoSettleRecordStore := prefix.NewStore(store, types.AutoSettleRecordKeyPrefix)

	pageRes, err := query.Paginate(autoSettleRecordStore, req.Pagination, func(key []byte, value []byte) error {
		autoSettleRecord := types.ParseAutoSettleRecordKey(key)
		autoSettleRecords = append(autoSettleRecords, autoSettleRecord)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAutoSettleRecordsResponse{AutoSettleRecords: autoSettleRecords, Pagination: pageRes}, nil
}
