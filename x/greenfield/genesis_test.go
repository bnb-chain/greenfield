package greenfield_test

import (
	"testing"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/greenfield"
	"github.com/bnb-chain/greenfield/x/greenfield/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.GreenfieldKeeper(t)
	greenfield.InitGenesis(ctx, *k, genesisState)
	got := greenfield.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
