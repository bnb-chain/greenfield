package keeper_test

import (
	"testing"

	testkeeper "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.ChallengeKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
	require.EqualValues(t, params.EventCountPerBlock, k.EventCountPerBlock(ctx))
	require.EqualValues(t, params.ChallengeExpirePeriod, k.ChallengeExpirePeriod(ctx))
	require.EqualValues(t, params.SlashCoolingOffPeriod, k.SlashCoolingOffPeriod(ctx))
	require.EqualValues(t, params.SlashAmountSizeRate, k.SlashAmountSizeRate(ctx))
	require.EqualValues(t, params.SlashAmountMin, k.SlashAmountMin(ctx))
	require.EqualValues(t, params.SlashAmountMax, k.SlashAmountMax(ctx))
	require.EqualValues(t, params.RewardValidatorRatio, k.RewardValidatorRatio(ctx))
	require.EqualValues(t, params.RewardChallengerRatio, k.RewardChallengerRatio(ctx))
}
