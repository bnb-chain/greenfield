package gashub

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
)

func TestGashubParams(t *testing.T) {
	client, err := gnfdclient.NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID)
	assert.NoError(t, err)

	query := gashubtypes.QueryParamsRequest{}
	res, err := client.GashubQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}
