package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/sp/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
)

var _ types.QueryServer = Keeper{}

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
