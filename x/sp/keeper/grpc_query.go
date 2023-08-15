package keeper

import (
	"context"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (k Keeper) StorageProviders(goCtx context.Context, req *types.QueryStorageProvidersRequest) (*types.QueryStorageProvidersResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	store := ctx.KVStore(k.storeKey)
	spStore := prefix.NewStore(store, types.StorageProviderKey)

	sps, pageRes, err := query.GenericFilteredPaginate(k.cdc, spStore, req.Pagination, func(key []byte, val *types.StorageProvider) (*types.StorageProvider, error) {
		return val, nil
	}, func() *types.StorageProvider {
		return &types.StorageProvider{}
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryStorageProvidersResponse{Sps: sps, Pagination: pageRes}, nil
}

func (k Keeper) QuerySpStoragePrice(goCtx context.Context, req *types.QuerySpStoragePriceRequest) (*types.QuerySpStoragePriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	spAddr, err := sdk.AccAddressFromHexUnsafe(req.SpAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid sp address")
	}
	sp, found := k.GetStorageProviderByOperatorAddr(ctx, spAddr)
	if !found {
		return nil, status.Error(codes.InvalidArgument, "unknown sp with the operator address")
	}
	spStoragePrice, found := k.GetSpStoragePrice(ctx, sp.Id)
	if !found {
		return nil, status.Errorf(codes.NotFound, "not found, err: %s", err)
	}
	return &types.QuerySpStoragePriceResponse{SpStoragePrice: spStoragePrice}, nil
}

func (k Keeper) QueryGlobalSpStorePriceByTime(goCtx context.Context, req *types.QueryGlobalSpStorePriceByTimeRequest) (*types.QueryGlobalSpStorePriceByTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Timestamp < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid timestamp")
	}
	if req.Timestamp == 0 {
		req.Timestamp = ctx.BlockTime().Unix() + 1
	}

	price, err := k.GetGlobalSpStorePriceByTime(ctx, req.Timestamp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err: %s", err)
	}
	return &types.QueryGlobalSpStorePriceByTimeResponse{GlobalSpStorePrice: price}, nil
}

func (k Keeper) StorageProvider(goCtx context.Context, req *types.QueryStorageProviderRequest) (*types.QueryStorageProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	sp, found := k.GetStorageProvider(ctx, req.Id)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}
	return &types.QueryStorageProviderResponse{StorageProvider: sp}, nil
}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) StorageProviderByOperatorAddress(goCtx context.Context, req *types.QueryStorageProviderByOperatorAddressRequest) (*types.QueryStorageProviderByOperatorAddressResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr, err := sdk.AccAddressFromHexUnsafe(req.OperatorAddress)
	if err != nil {
		return nil, err
	}
	sp, found := k.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}
	return &types.QueryStorageProviderByOperatorAddressResponse{StorageProvider: sp}, nil
}

func (k Keeper) StorageProviderMaintenanceRecordsByOperatorAddress(goCtx context.Context, req *types.QueryStorageProviderMaintenanceRecordsRequest) (*types.QueryStorageProviderMaintenanceRecordsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	operatorAddr, err := sdk.AccAddressFromHexUnsafe(req.OperatorAddress)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid primary storage provider address")
	}
	records := make([]*types.MaintenanceRecord, 0)
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.GetStorageProviderMaintenanceRecordsKey(operatorAddr))
	if bz != nil {
		stats := &types.SpMaintenanceStats{}
		k.cdc.MustUnmarshal(bz, stats)
		records = stats.Records
	}
	return &types.QueryStorageProviderMaintenanceRecordsResponse{Records: records}, nil
}
