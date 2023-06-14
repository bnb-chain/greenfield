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
