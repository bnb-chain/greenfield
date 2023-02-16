package gashub

import (
	"context"
	"testing"

	gnfdclient "github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	"github.com/stretchr/testify/assert"
)

func TestGashubParams(t *testing.T) {
	client := gnfdclient.NewGreenfieldClient(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID)

	query := gashubtypes.QueryParamsRequest{}
	res, err := client.GashubQueryClient.Params(context.Background(), &query)
	assert.NoError(t, err)

	t.Log(res.String())
}
