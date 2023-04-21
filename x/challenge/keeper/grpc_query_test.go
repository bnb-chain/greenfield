package keeper_test

import (
	"encoding/hex"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	params := types.DefaultParams()
	keeper.SetParams(ctx, params)

	response, err := keeper.Params(wctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestLatestAttestedChallengesQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	keeper.SetParams(ctx, types.DefaultParams())
	keeper.AppendAttestChallengeId(ctx, 100)
	keeper.AppendAttestChallengeId(ctx, 200)

	response, err := keeper.LatestAttestedChallenges(wctx, &types.QueryLatestAttestedChallengesRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryLatestAttestedChallengesResponse{ChallengeIds: []uint64{100, 200}}, response)
}

func TestInturnAttestationSubmitterQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	keeper.SetParams(ctx, types.DefaultParams())

	ctrl := gomock.NewController(t)
	stakingKeeper := types.NewMockStakingKeeper(ctrl)
	keeper.SetStakingKeeper(stakingKeeper)

	blsKey := []byte("blskey")
	historicalInfo := stakingtypes.HistoricalInfo{
		Header: tmproto.Header{},
		Valset: []stakingtypes.Validator{stakingtypes.Validator{BlsKey: blsKey}},
	}
	stakingKeeper.EXPECT().GetHistoricalInfo(gomock.Any(), gomock.Any()).Return(historicalInfo, true).AnyTimes()

	response, err := keeper.InturnAttestationSubmitter(wctx, &types.QueryInturnAttestationSubmitterRequest{})
	require.NoError(t, err)
	require.Equal(t, blsKey, response.BlsPubKey, hex.EncodeToString(blsKey))
}
