package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) ChallengeAll(c context.Context, req *types.QueryAllChallengeRequest) (*types.QueryAllChallengeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	var ongoingChallenges []types.Challenge
	ctx := sdk.UnwrapSDKContext(c)

	store := ctx.KVStore(k.storeKey)
	ongoingChallengeStore := prefix.NewStore(store, types.KeyPrefix(types.OngoingChallengeKeyPrefix))

	pageRes, err := query.Paginate(ongoingChallengeStore, req.Pagination, func(key []byte, value []byte) error {
		var ongoingChallenge types.Challenge
		if err := k.cdc.Unmarshal(value, &ongoingChallenge); err != nil {
			return err
		}

		ongoingChallenges = append(ongoingChallenges, ongoingChallenge)
		return nil
	})

	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &types.QueryAllChallengeResponse{Challenge: ongoingChallenges, Pagination: pageRes}, nil
}

func (k Keeper) Challenge(c context.Context, req *types.QueryGetChallengeRequest) (*types.QueryGetChallengeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	val, found := k.GetOngoingChallenge(
		ctx,
		req.ChallengeId,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}

	return &types.QueryGetChallengeResponse{Challenge: val}, nil
}
