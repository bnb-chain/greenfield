package keeper

import (
	"context"
	"math"

	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
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

	gvgFamily, found := k.GetGVGFamily(ctx, req.GlobalVirtualGroupFamilyId)
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

func (k Keeper) GlobalVirtualGroupFamily(goCtx context.Context, req *types.QueryGlobalVirtualGroupFamilyRequest) (*types.QueryGlobalVirtualGroupFamilyResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	gvgFamily, found := k.GetGVGFamily(ctx, req.FamilyId)
	if !found {
		return nil, types.ErrGVGFamilyNotExist
	}

	return &types.QueryGlobalVirtualGroupFamilyResponse{GlobalVirtualGroupFamily: gvgFamily}, nil
}

func (k Keeper) GlobalVirtualGroupFamilies(goCtx context.Context, req *types.QueryGlobalVirtualGroupFamiliesRequest) (*types.QueryGlobalVirtualGroupFamiliesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var gvgFamilies []*types.GlobalVirtualGroupFamily
	store := ctx.KVStore(k.storeKey)
	gvgFamilyStore := prefix.NewStore(store, types.GVGFamilyKey)

	pageRes, err := query.Paginate(gvgFamilyStore, req.Pagination, func(key []byte, value []byte) error {
		var gvgFamily types.GlobalVirtualGroupFamily
		k.cdc.MustUnmarshal(value, &gvgFamily)
		gvgFamilies = append(gvgFamilies, &gvgFamily)
		return nil
	})
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &types.QueryGlobalVirtualGroupFamiliesResponse{GvgFamilies: gvgFamilies, Pagination: pageRes}, nil
}

func (k Keeper) AvailableGlobalVirtualGroupFamilies(goCtx context.Context, req *types.AvailableGlobalVirtualGroupFamiliesRequest) (*types.AvailableGlobalVirtualGroupFamiliesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	availableFamilyIds := make([]uint32, 0)
	for _, gvgfID := range req.GlobalVirtualGroupFamilyIds {
		gvgFamily, found := k.GetGVGFamily(ctx, gvgfID)
		if !found {
			return nil, types.ErrGVGFamilyNotExist
		}
		totalStakingSize, stored, err := k.GetGlobalVirtualFamilyTotalStakingAndStoredSize(ctx, gvgFamily)
		if err != nil {
			return nil, err
		}
		if float64(stored) < math.Min(float64(totalStakingSize), float64(k.MaxStoreSizePerFamily(ctx))) && uint32(len(gvgFamily.GlobalVirtualGroupIds)) < k.MaxGlobalVirtualGroupNumPerFamily(ctx) {
			availableFamilyIds = append(availableFamilyIds, gvgfID)
		}
	}
	return &types.AvailableGlobalVirtualGroupFamiliesResponse{GlobalVirtualGroupFamilyIds: availableFamilyIds}, nil
}

func (k Keeper) SwapInInfo(goCtx context.Context, req *types.QuerySwapInInfoRequest) (*types.QuerySwapInInfoResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	swapInInfo, found := k.GetSwapInInfo(ctx, req.GetGlobalVirtualGroupFamilyId(), req.GetGlobalVirtualGroupId())
	if !found {
		return nil, types.ErrSwapInInfoNotExist
	}
	if uint64(ctx.BlockTime().Unix()) >= swapInInfo.ExpirationTime {
		return nil, types.ErrSwapInInfoExpired
	}
	return &types.QuerySwapInInfoResponse{
		SwapInInfo: swapInInfo,
	}, nil
}

func (k Keeper) GetSwapInInfo(ctx sdk.Context, globalVirtualGroupFamilyId, globalVirtualGroupId uint32) (*types.SwapInInfo, bool) {
	store := ctx.KVStore(k.storeKey)
	var key []byte
	if globalVirtualGroupFamilyId != types.NoSpecifiedFamilyId {
		key = types.GetSwapInFamilyKey(globalVirtualGroupFamilyId)
	} else {
		key = types.GetSwapInGVGKey(globalVirtualGroupId)
	}
	bz := store.Get(key)
	if bz == nil {
		return nil, false
	}
	swapInInfo := &types.SwapInInfo{}
	k.cdc.MustUnmarshal(bz, swapInInfo)
	return swapInInfo, true
}
