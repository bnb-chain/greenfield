package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ChallengeKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
	require.EqualValues(t, params.ChallengeCountPerBlock, k.ChallengeCountPerBlock(ctx))
	require.EqualValues(t, params.SlashCoolingOffPeriod, k.SlashCoolingOffPeriod(ctx))
	require.EqualValues(t, params.SlashAmountSizeRate, k.SlashAmountSizeRate(ctx))
	require.EqualValues(t, params.SlashAmountMin, k.SlashAmountMin(ctx))
	require.EqualValues(t, params.SlashAmountMax, k.SlashAmountMax(ctx))
	require.EqualValues(t, params.RewardValidatorRatio, k.RewardValidatorRatio(ctx))
	require.EqualValues(t, params.RewardSubmitterRatio, k.RewardSubmitterRatio(ctx))
	require.EqualValues(t, params.RewardSubmitterThreshold, k.RewardSubmitterThreshold(ctx))
	require.EqualValues(t, params.HeartbeatInterval, k.HeartbeatInterval(ctx))
}
