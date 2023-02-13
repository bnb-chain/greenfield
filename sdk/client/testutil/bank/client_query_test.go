//go:build with_local_net

package bank

import (
	"context"
	"testing"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client/rpc"
	"github.com/bnb-chain/greenfield/sdk/client/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
)

func TestBankBalance(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QueryBalanceRequest{
		Address: testutil.TEST_ADDR,
		Denom:   "bnb",
	}
	res, err := client.BankQueryClient.Balance(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.Balance.String())
}

func TestBankAllBalances(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QueryAllBalancesRequest{
		Address: testutil.TEST_ADDR,
	}
	res, err := client.BankQueryClient.AllBalances(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.Balances.String())
}

func TestBankDenomOwners(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QueryDenomOwnersRequest{
		Denom: "bnb",
	}
	res, err := client.BankQueryClient.DenomOwners(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankDenomsMetadata(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QueryDenomsMetadataRequest{}
	res, err := client.BankQueryClient.DenomsMetadata(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QueryParamsRequest{}
	res, err := client.BankQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankSpendableBalance(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QuerySpendableBalancesRequest{
		Address: testutil.TEST_ADDR,
	}
	res, err := client.BankQueryClient.SpendableBalances(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetBalances().String())
}

func TestBankSupplyOf(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QuerySupplyOfRequest{
		Denom: "bnb",
	}
	res, err := client.BankQueryClient.SupplyOf(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankTotalSupply(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(testutil.TEST_GRPC_ADDR, testutil.TEST_CHAIN_ID)

	query := banktypes.QueryTotalSupplyRequest{}
	res, err := client.BankQueryClient.TotalSupply(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.Supply.String())
}
