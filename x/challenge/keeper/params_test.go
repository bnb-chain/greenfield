package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()

	k.SetParams(ctx, params)
	require.EqualValues(t, params, k.GetParams(ctx))

	params.AttestationKeptCount = 100
	k.SetParams(ctx, params)
	require.EqualValues(t, params, k.GetParams(ctx))
}
