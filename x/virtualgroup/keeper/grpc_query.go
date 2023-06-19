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

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) GlobalVirtualGroup(goCtx context.Context, req *types.QueryGlobalVirtualGroupRequest) (*types.QueryGlobalVirtualGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	gvg, found := k.GetGVG(ctx, req.GlobalVirtualGroupId)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	return &types.QueryGlobalVirtualGroupResponse{GlobalVirtualGroup: gvg}, nil
}

func (k Keeper) GlobalVirtualGroupByFamilyID(goCtx context.Context, req *types.QueryGlobalVirtualGroupByFamilyIDRequest) (*types.QueryGlobalVirtualGroupByFamilyIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	gvgFamily, found := k.GetGVGFamily(ctx, req.StorageProviderId, req.GlobalVirtualGroupFamilyId)
	if !found {
		return nil, types.ErrGVGFamilyNotExist
	}
	var gvgs []*types.GlobalVirtualGroup
	for _, gvgID := range gvgFamily.GlobalVirtualGroupIds {
		gvg, found := k.GetGVG(ctx, gvgID)
		if !found {
			panic("gvg not found, but id exists in family")
		}
		gvgs = append(gvgs, gvg)
	}

	return &types.QueryGlobalVirtualGroupByFamilyIDResponse{GlobalVirtualGroups: gvgs}, nil
}

func (k Keeper) GlobalVirtualGroupBySPID(goCtx context.Context, req *types.QueryGlobalVirtualGroupBySPIDRequest) (*types.QueryGlobalVirtualGroupBySPIDResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var uint32Sequenct sequence.Sequence[uint32]
	store := ctx.KVStore(k.storeKey)
	gvgStore := prefix.NewStore(store, append(types.GVGKey, uint32Sequenct.EncodeSequence(req.StorageProviderId)...))

	var gvgs []*types.GlobalVirtualGroup
	pageRes, err := query.Paginate(gvgStore, req.Pagination, func(key []byte, value []byte) error {
		var gvg types.GlobalVirtualGroup
		k.cdc.MustUnmarshal(value, &gvg)
		gvgs = append(gvgs, &gvg)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryGlobalVirtualGroupBySPIDResponse{GlobalVirtualGroups: gvgs, Pagination: pageRes}, nil
}

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

func (k Keeper) GlobalVirtualGroupFamily(goCtx context.Context, req *types.QueryGlobalVirtualGroupFamilyRequest) (*types.QueryGlobalVirtualGroupFamilyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	gvgFamily, found := k.GetGVGFamily(ctx, req.StorageProviderId, req.FamilyId)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	return &types.QueryGlobalVirtualGroupFamilyResponse{GlobalVirtualGroupFamily: gvgFamily}, nil
}
