package bank

import (
	"context"
	"testing"

	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/sdk/keys"
)

func TestBankBalance(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryBalanceRequest{
		Address: test.TEST_ADDR,
		Denom:   "bnb",
	}
	res, err := client.BankQueryClient.Balance(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.Balance.String())
}

func TestBankAllBalances(t *testing.T) {
	km, err := keys.NewPrivateKeyManager("e3ac46e277677f0f103774019d03bd89c7b4b5ecc554b2650bd5d5127992c20c")
	assert.NoError(t, err)
	t.Log(km)
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryAllBalancesRequest{
		Address: test.TEST_ADDR,
	}
	res, err := client.BankQueryClient.AllBalances(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.Balances.String())
}

func TestBankDenomMetadata(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryDenomMetadataRequest{}
	res, err := client.BankQueryClient.DenomMetadata(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankDenomOwners(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryDenomOwnersRequest{
		Denom: "bnb",
	}
	res, err := client.BankQueryClient.DenomOwners(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankDenomsMetadata(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryDenomsMetadataRequest{}
	res, err := client.BankQueryClient.DenomsMetadata(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryParamsRequest{}
	res, err := client.BankQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankSpendableBalance(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QuerySpendableBalancesRequest{
		Address: test.TEST_ADDR,
	}
	res, err := client.BankQueryClient.SpendableBalances(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetBalances().String())
}

func TestBankSupplyOf(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QuerySupplyOfRequest{
		Denom: "bnb",
	}
	res, err := client.BankQueryClient.SupplyOf(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}

func TestBankTotalSupply(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := banktypes.QueryTotalSupplyRequest{}
	res, err := client.BankQueryClient.TotalSupply(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.Supply.String())
}
