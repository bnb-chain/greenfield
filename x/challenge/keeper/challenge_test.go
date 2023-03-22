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
	keeper.SetChallengeId(ctx, 100)
	require.True(t, keeper.GetChallengeId(ctx) == 100)
}
