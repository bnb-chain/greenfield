package bank

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
)

func TestOracleParams(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := oracletypes.QueryParamsRequest{}
	res, err := client.OracleQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetParams())
}
