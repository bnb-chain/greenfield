package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestOngoingChallengeId(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	keeper.SetOngoingChallengeId(ctx, 100)
	require.True(t, keeper.GetOngoingChallengeId(ctx) == 100)
}
