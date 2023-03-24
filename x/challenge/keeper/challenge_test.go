package keeper_test

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestGetChallengeId(t *testing.T) {
	keeper, ctx := keepertest.ChallengeKeeper(t)
	keeper.SaveChallenge(ctx, types.Challenge{
		Id:            100,
		ExpiredHeight: 1000,
	})
	require.True(t, keeper.GetChallengeId(ctx) == 100)
}
