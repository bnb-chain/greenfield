package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func TestGetParams(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()

	err := k.SetParams(ctx, params)
	require.NoError(t, err)
	require.EqualValues(t, params, k.GetParams(ctx))

	params.AttestationKeptCount = 100
	err = k.SetParams(ctx, params)
	require.NoError(t, err)
	require.EqualValues(t, params, k.GetParams(ctx))
}
