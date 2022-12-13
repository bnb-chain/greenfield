package storage_test

import (
	"testing"

	keepertest "github.com/bnb-chain/inscription/testutil/keeper"
	"github.com/bnb-chain/inscription/testutil/nullify"
	"github.com/bnb-chain/inscription/x/storage"
	"github.com/bnb-chain/inscription/x/storage/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.StorageKeeper(t)
	storage.InitGenesis(ctx, *k, genesisState)
	got := storage.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	// this line is used by starport scaffolding # genesis/test/assert
}
