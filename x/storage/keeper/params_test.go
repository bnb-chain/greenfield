package keeper_test

import (
	"testing"

	testkeeper "github.com/bnb-chain/inscription/testutil/keeper"
	"github.com/bnb-chain/inscription/x/storage/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.StorageKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
