package keeper

import (
	"context"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) LatestAttestedChallenge(goCtx context.Context, req *types.QueryLatestAttestedChallengeRequest) (*types.QueryLatestAttestedChallengeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	challengeId := k.GetAttestChallengeId(ctx)

	return &types.QueryLatestAttestedChallengeResponse{
		ChallengeId: challengeId,
	}, nil
}
