package keeper

import (
	"context"
	"encoding/hex"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}
	ctx := sdk.UnwrapSDKContext(c)

	return &types.QueryParamsResponse{Params: k.GetParams(ctx)}, nil
}

func (k Keeper) LatestAttestedChallenges(goCtx context.Context, req *types.QueryLatestAttestedChallengesRequest) (*types.QueryLatestAttestedChallengesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	challengeId := k.GetAttestChallengeIds(ctx)

	return &types.QueryLatestAttestedChallengesResponse{
		ChallengeIds: challengeId,
	}, nil
}

func (k Keeper) InturnAttestationSubmitter(goCtx context.Context, req *types.QueryInturnAttestationSubmitterRequest) (*types.QueryInturnAttestationSubmitterResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	blsKey, interval, err := k.getInturnSubmitter(ctx, k.GetParams(ctx).AttestationInturnInterval)
	if err != nil {
		return nil, err
	}

	return &types.QueryInturnAttestationSubmitterResponse{
		BlsPubKey:      hex.EncodeToString(blsKey),
		SubmitInterval: interval,
	}, nil

}
