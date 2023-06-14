package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

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
