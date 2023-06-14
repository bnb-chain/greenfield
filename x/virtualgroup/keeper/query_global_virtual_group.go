package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) GlobalVirtualGroup(goCtx context.Context, req *types.QueryGlobalVirtualGroupRequest) (*types.QueryGlobalVirtualGroupResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	gvg, found := k.GetGVG(ctx, req.StorageProviderId, req.GlobalVirtualGroupId)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	return &types.QueryGlobalVirtualGroupResponse{GlobalVirtualGroup: gvg}, nil
}
