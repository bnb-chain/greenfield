package client

import (
	"testing"

	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
)

func TestSendTokenSucceedWithSimulatedGas(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID, km)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 12)))
	response, err := gnfdCli.BroadcastTx([]sdk.Msg{transfer}, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
	t.Log(response.TxResponse.String())
}

func TestSendTokenWithTxOptionSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID, km)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 100)))
	payerAddr, err := sdk.AccAddressFromHexUnsafe(km.GetAddr().String())
	assert.NoError(t, err)
	mode := tx.BroadcastMode_BROADCAST_MODE_ASYNC
	txOpt := &types.TxOption{
		Mode:      &mode,
		GasLimit:  123456,
		Memo:      "test",
		FeeAmount: sdk.Coins{{Denom: "bnb", Amount: sdk.NewInt(1)}},
		FeePayer:  payerAddr,
	}
	response, err := gnfdCli.BroadcastTx([]sdk.Msg{transfer}, txOpt)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
	t.Log(response.TxResponse.String())
}

func TestSimulateTx(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli := NewGreenfieldClientWithKeyManager(test.TEST_GRPC_ADDR, test.TEST_CHAIN_ID, km)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin("bnb", 100)))
	simulateRes, err := gnfdCli.SimulateTx([]sdk.Msg{transfer}, nil)
	assert.NoError(t, err)
	t.Log(simulateRes.GasInfo.String())
}
