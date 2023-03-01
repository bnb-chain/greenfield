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

	SPs := types.StorageProviders{}
	for _, sp := range sps {
		SPs = append(SPs, *sp)
	}

	return &types.QueryStorageProvidersResponse{Sps: SPs, Pagination: pageRes}, nil
}

func (k Keeper) QueryGetSpStoragePriceByTime(goCtx context.Context, req *types.QueryGetSpStoragePriceByTimeRequest) (*types.QueryGetSpStoragePriceByTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Timestamp < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid timestamp")
	}
	if req.Timestamp == 0 {
		req.Timestamp = ctx.BlockTime().Unix()
	}
	_, err := sdk.AccAddressFromHexUnsafe(req.SpAddr)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid sp address")
	}
	spStoragePrice, err := k.GetSpStoragePriceByTime(ctx, req.SpAddr, req.Timestamp)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "not found, err: %s", err)
	}
	return &types.QueryGetSpStoragePriceByTimeResponse{SpStoragePrice: spStoragePrice}, nil
}

func (k Keeper) QueryGetSecondarySpStorePriceByTime(goCtx context.Context, req *types.QueryGetSecondarySpStorePriceByTimeRequest) (*types.QueryGetSecondarySpStorePriceByTimeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	if req.Timestamp < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid timestamp")
	}
	if req.Timestamp == 0 {
		req.Timestamp = ctx.BlockTime().Unix()
	}

	price, err := k.GetSecondarySpStorePriceByTime(ctx, req.Timestamp)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "err: %s", err)
	}
	return &types.QueryGetSecondarySpStorePriceByTimeResponse{SecondarySpStorePrice: price}, nil
}

func (k Keeper) StorageProvider(goCtx context.Context, req *types.QueryStorageProviderRequest) (*types.QueryStorageProviderResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	spAddr, err := sdk.AccAddressFromHexUnsafe(req.SpAddress)
	if err != nil {
		return nil, err
	}
	sp, found := k.GetStorageProvider(ctx, spAddr)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}
	return &types.QueryStorageProviderResponse{StorageProvider: &sp}, nil
}
