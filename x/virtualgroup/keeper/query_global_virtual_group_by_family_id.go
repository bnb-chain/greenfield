package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
		gvg, found := k.GetGVG(ctx, req.StorageProviderId, gvgID)
		if !found {
			panic("gvg not found, but id exists in family")
		}
		gvgs = append(gvgs, gvg)
	}

	return &types.QueryGlobalVirtualGroupByFamilyIDResponse{GlobalVirtualGroups: gvgs}, nil
}
