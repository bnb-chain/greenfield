package types_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/types"
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

	// validate heartbeat interval
	params.RewardValidatorRatio = sdk.NewDecWithPrec(5, 1)
	params.RewardChallengerRatio = sdk.NewDecWithPrec(4, 1)
	params.HeartbeatInterval = 0
	require.Error(t, params.Validate())

	// validate heartbeat reward rate
	params.HeartbeatInterval = 100
	params.HeartbeatRewardRate = sdk.NewDec(-1)
	require.Error(t, params.Validate())

	// validate heartbeat reward threshold
	params.HeartbeatRewardRate = sdk.NewDecWithPrec(1, 3)
	params.HeartbeatRewardThreshold = sdk.NewInt(-1)
	require.Error(t, params.Validate())

	// no error
	params.HeartbeatRewardThreshold = sdk.NewInt(100)
	require.NoError(t, params.Validate())
}
