package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/internal/sequence"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GlobalVirtualGroupFamilies(goCtx context.Context, req *types.QueryGlobalVirtualGroupFamiliesRequest) (*types.QueryGlobalVirtualGroupFamiliesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	store := ctx.KVStore(k.storeKey)
	var uint32Sequence sequence.Sequence[uint32]
	gvgFamiliesStore := prefix.NewStore(store, append(types.GVGFamilyKey, uint32Sequence.EncodeSequence(req.StorageProviderId)...))
	var gvgFamiles []*types.GlobalVirtualGroupFamily
	pageRes, err := query.Paginate(gvgFamiliesStore, req.Pagination, func(key []byte, value []byte) error {
		var gvgFamily types.GlobalVirtualGroupFamily
		k.cdc.MustUnmarshal(value, &gvgFamily)
		gvgFamiles = append(gvgFamiles, &gvgFamily)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGlobalVirtualGroupFamiliesResponse{GlobalVirtualGroupFamilies: gvgFamiles, Pagination: pageRes}, nil
}
