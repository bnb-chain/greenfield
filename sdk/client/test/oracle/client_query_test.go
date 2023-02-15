package bank

import (
	"context"
	"testing"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client/chain"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/stretchr/testify/assert"
)

func TestOracleParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := oracletypes.QueryParamsRequest{}
	res, err := client.OracleQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.GetParams())
}
