package staking

import (
	"context"
	"testing"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/assert"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/client/test"
)

func TestStakingValidator(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryValidatorRequest{
		ValidatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.Validator(context.Background(), &query)
	assert.NoError(t, err)
	assert.Equal(t, res.Validator.SelfDelAddress, test.TEST_VAL_ADDR)
}

func TestStakingValidators(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryValidatorsRequest{
		Status: "",
	}
	res, err := client.StakingQueryClient.Validators(context.Background(), &query)
	assert.NoError(t, err)
	assert.True(t, len(res.Validators) > 0)
}

func TestStakingDelagatorValidator(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryDelegatorValidatorRequest{
		DelegatorAddr: test.TEST_ADDR,
		ValidatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorValidator(context.Background(), &query)
	assert.NoError(t, err)

	assert.Equal(t, res.Validator.SelfDelAddress, test.TEST_VAL_ADDR)
}

func TestStakingDelagatorValidators(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryDelegatorValidatorsRequest{
		DelegatorAddr: test.TEST_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorValidators(context.Background(), &query)
	assert.NoError(t, err)

	assert.True(t, len(res.Validators) > 0)
}

func TestStakingUnbondingDelagation(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryUnbondingDelegationRequest{
		DelegatorAddr: test.TEST_ADDR,
		ValidatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.UnbondingDelegation(context.Background(), &query)
	assert.NoError(t, err)

	assert.Equal(t, res.Unbond.DelegatorAddress, test.TEST_ADDR)
}

func TestStakingDelagatorDelegations(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryDelegatorDelegationsRequest{
		DelegatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingValidatorDelegations(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryValidatorDelegationsRequest{
		ValidatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.ValidatorDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingDelegatorUnbondingDelagation(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryDelegatorUnbondingDelegationsRequest{
		DelegatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.DelegatorUnbondingDelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStaking(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryRedelegationsRequest{
		DelegatorAddr: test.TEST_VAL_ADDR,
	}
	res, err := client.StakingQueryClient.Redelegations(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingParams(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryParamsRequest{}
	res, err := client.StakingQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingPool(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryPoolRequest{}
	res, err := client.StakingQueryClient.Pool(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestStakingHistoricalInfo(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := stakingtypes.QueryHistoricalInfoRequest{
		Height: 1,
	}
	res, err := client.StakingQueryClient.HistoricalInfo(context.Background(), &query)
	assert.NoError(t, err)

	assert.True(t, len(res.GetHist().Valset) > 0)
}
