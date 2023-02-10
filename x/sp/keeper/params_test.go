package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	testkeeper "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := testkeeper.SpKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(t, params, k.GetParams(ctx))
}
