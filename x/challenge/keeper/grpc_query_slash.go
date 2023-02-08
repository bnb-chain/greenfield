package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) SlashAll(c context.Context, req *types.QueryAllSlashRequest) (*types.QueryAllSlashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var recentSlashs []types.Slash
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	recentSlashStore := prefix.NewStore(store, types.KeyPrefix(types.RecentSlashKey))

	pageRes, err := query.Paginate(recentSlashStore, req.Pagination, func(key []byte, value []byte) error {
		var recentSlash types.Slash
		if err := k.cdc.Unmarshal(value, &recentSlash); err != nil {
			return err
		}

		recentSlashs = append(recentSlashs, recentSlash)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllSlashResponse{Slash: recentSlashs, Pagination: pageRes}, nil
}

func (k Keeper) Slash(c context.Context, req *types.QueryGetSlashRequest) (*types.QueryGetSlashResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(c)
	recentSlash, found := k.GetRecentSlash(ctx, req.Id)
	if !found {
		return nil, sdkerrors.ErrKeyNotFound
	}

	return &types.QueryGetSlashResponse{Slash: recentSlash}, nil
}
