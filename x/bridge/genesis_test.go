package bridge_test

import (
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/bridge"
	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.BridgeKeeper(t)
	bridge.InitGenesis(ctx, *k, genesisState)
	got := bridge.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
