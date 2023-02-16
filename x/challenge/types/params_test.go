package types_test

import (
	"testing"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func Test_validateParams(t *testing.T) {
	params := types.DefaultParams()

	// default params have no error
	require.NoError(t, params.Validate())

	// validate slash amount min
	params.SlashAmountMin = sdk.NewInt(-1)
	require.Error(t, params.Validate())

	// validate slash amount max
	params.SlashAmountMin = sdk.NewInt(1)
	params.SlashAmountMax = sdk.NewInt(-1)
	require.Error(t, params.Validate())

	params.SlashAmountMin = sdk.NewInt(10)
	params.SlashAmountMax = sdk.NewInt(1)
	require.Error(t, params.Validate())

	params.SlashAmountMin = sdk.NewInt(1)
	params.SlashAmountMax = sdk.NewInt(10)
	require.NoError(t, params.Validate())

	// validate reward validator ratio
	params.RewardValidatorRatio = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	// validate reward challenger ratio
	params.RewardValidatorRatio = sdk.NewDecWithPrec(5, 1)
	params.RewardChallengerRatio = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	params.RewardValidatorRatio = sdk.NewDecWithPrec(5, 1)
	params.RewardChallengerRatio = sdk.NewDecWithPrec(7, 1)
	require.Error(t, params.Validate())
}
