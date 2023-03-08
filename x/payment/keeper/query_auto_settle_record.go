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
		var autoSettleRecord types.AutoSettleRecord
		if err := k.cdc.Unmarshal(value, &autoSettleRecord); err != nil {
			return err
		}

		autoSettleRecords = append(autoSettleRecords, autoSettleRecord)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllAutoSettleRecordResponse{AutoSettleRecord: autoSettleRecords, Pagination: pageRes}, nil
}

func (k Keeper) AutoSettleRecord(c context.Context, req *types.QueryGetAutoSettleRecordRequest) (*types.QueryGetAutoSettleRecordResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	addr, err := sdk.AccAddressFromBech32(req.Addr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	val, found := k.GetAutoSettleRecord(
		ctx,
		req.Timestamp,
		addr,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetAutoSettleRecordResponse{AutoSettleRecord: *val}, nil
}
