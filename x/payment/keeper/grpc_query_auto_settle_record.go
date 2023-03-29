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

func (k Keeper) AutoSettleRecordAll(c context.Context, req *types.QueryAllAutoSettleRecordRequest) (*types.QueryAllAutoSettleRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var autoSettleRecords []types.AutoSettleRecord
	ctx := sdk.UnwrapSDKContext(c)

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

	return &types.QueryAllAutoSettleRecordResponse{AutoSettleRecord: autoSettleRecords, Pagination: pageRes}, nil
}
