package keeper_test

import (
	"testing"

	testkeeper "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/stretchr/testify/require"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.SpKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
