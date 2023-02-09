package challenge_test

import (
	"testing"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/challenge"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		OngoingChallenges: []types.Challenge{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		RecentSlashes: []types.Slash{
			{
				Id: 0,
			},
			{
				Id: 1,
			},
		},
		RecentSlashCount: 2,
		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.ChallengeKeeper(t)
	challenge.InitGenesis(ctx, *k, genesisState)
	got := challenge.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.OngoingChallenges, got.OngoingChallenges)
	require.ElementsMatch(t, genesisState.RecentSlashes, got.RecentSlashes)
	require.Equal(t, genesisState.RecentSlashCount, got.RecentSlashCount)
	// this line is used by starport scaffolding # genesis/test/assert
}
