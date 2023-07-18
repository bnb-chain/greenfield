package client

import (
	"context"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTmClient(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID, WithKeyManager(km), WithWebSocketClient())
	assert.NoError(t, err)
	//to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	//assert.NoError(t, err)
	//transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 12)))
	//response, err := gnfdCli.BroadcastTx(context.Background(), []sdk.Msg{transfer}, nil)
	//assert.NoError(t, err)
	//time.Sleep(2 * time.Second)
	//// get tx
	//res, err := gnfdCli.Tx(context.Background(), response.TxResponse.TxHash)
	//assert.NoError(t, err)
	//t.Log(res.Hash)
	//t.Log(res.TxResult.String())

	// get the latest block
	block, err := gnfdCli.GetBlock(context.Background(), nil)
	assert.NoError(t, err)
	t.Log(block)

	h := block.Block.Height

	block, err = gnfdCli.GetBlock(context.Background(), &h)
	assert.NoError(t, err)
	t.Log(block)

	// get block result
	blockResult, err := gnfdCli.GetBlockResults(context.Background(), &h)
	assert.NoError(t, err)
	t.Log(blockResult)

	// get validator
	validators, err := gnfdCli.GetValidators(context.Background(), &h)
	assert.NoError(t, err)
	t.Log(validators.Validators)

	for i := 1; i < 10000; i++ {
		h := int64(i)
		ctx, _ := context.WithTimeout(context.Background(), 15*time.Second)
		_, err := gnfdCli.GetBlockResults(ctx, &h)
		assert.NoError(t, err) // why context exceeding is not treated as err
		t.Log(i)
	}
}
