package client

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
)

func TestSendTokenSucceedWithSimulatedGas(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID, WithKeyManager(km))
	assert.NoError(t, err)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 12)))
	response, err := gnfdCli.BroadcastTx(context.Background(), []sdk.Msg{transfer}, nil)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
	t.Log(response.TxResponse.String())
}

func TestSendTokenWithTxOptionSucceed(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID, WithKeyManager(km))
	assert.NoError(t, err)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 100)))
	payerAddr, err := sdk.AccAddressFromHexUnsafe(km.GetAddr().String())
	assert.NoError(t, err)
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	feeAmt := sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(int64(10000000000000)))) // gasPrice * gasLimit

	txOpt := &types.TxOption{
		Mode:       &mode,
		NoSimulate: true,
		GasLimit:   2000,
		Memo:       "test",
		FeePayer:   payerAddr,
		FeeAmount:  feeAmt, // 2000 * 5000000000
	}
	response, err := gnfdCli.BroadcastTx(context.Background(), []sdk.Msg{transfer}, txOpt)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
	t.Log(response.TxResponse.String())
}

func TestErrorOutWhenGasInfoNotFullProvided(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID, WithKeyManager(km))
	assert.NoError(t, err)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 100)))
	payerAddr, err := sdk.AccAddressFromHexUnsafe(km.GetAddr().String())
	assert.NoError(t, err)
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &types.TxOption{
		Mode:       &mode,
		NoSimulate: true,
		Memo:       "test",
		FeePayer:   payerAddr,
	}
	_, err = gnfdCli.BroadcastTx(context.Background(), []sdk.Msg{transfer}, txOpt)
	assert.Equal(t, err, types.GasInfoNotProvidedError)
}

func TestSimulateTx(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID, WithKeyManager(km))
	assert.NoError(t, err)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 100)))
	simulateRes, err := gnfdCli.SimulateTx(context.Background(), []sdk.Msg{transfer}, nil)
	assert.NoError(t, err)
	t.Log(simulateRes.GasInfo.String())
}

func TestSendTokenWithCustomizedNonce(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient(test.TEST_RPC_ADDR, test.TEST_CHAIN_ID, WithKeyManager(km))
	assert.NoError(t, err)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 100)))
	payerAddr, err := sdk.AccAddressFromHexUnsafe(km.GetAddr().String())
	assert.NoError(t, err)
	nonce, err := gnfdCli.GetNonce()
	assert.NoError(t, err)
	for i := 0; i < 50; i++ {
		txOpt := &types.TxOption{
			GasLimit: 123456,
			Memo:     "test",
			FeePayer: payerAddr,
			Nonce:    nonce,
		}
		response, err := gnfdCli.BroadcastTx(context.Background(), []sdk.Msg{transfer}, txOpt)
		assert.NoError(t, err)
		nonce++
		assert.Equal(t, uint32(0), response.TxResponse.Code)
		t.Log(response.TxResponse.String())
	}
}

func TestSendTxWithGrpcConn(t *testing.T) {
	km, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	assert.NoError(t, err)
	gnfdCli, err := NewGreenfieldClient("", test.TEST_CHAIN_ID, WithKeyManager(km), WithGrpcConnectionAndDialOption(test.TEST_GRPC_ADDR, grpc.WithTransportCredentials(insecure.NewCredentials())))
	assert.NoError(t, err)
	to, err := sdk.AccAddressFromHexUnsafe(test.TEST_ADDR)
	assert.NoError(t, err)
	transfer := banktypes.NewMsgSend(km.GetAddr(), to, sdk.NewCoins(sdk.NewInt64Coin(test.TEST_TOKEN_NAME, 100)))
	payerAddr, err := sdk.AccAddressFromHexUnsafe(km.GetAddr().String())
	assert.NoError(t, err)
	nonce, err := gnfdCli.GetNonce()
	assert.NoError(t, err)
	txOpt := &types.TxOption{
		GasLimit: 123456,
		Memo:     "test",
		FeePayer: payerAddr,
		Nonce:    nonce,
	}
	response, err := gnfdCli.BroadcastTx(context.Background(), []sdk.Msg{transfer}, txOpt)
	assert.NoError(t, err)
	assert.Equal(t, uint32(0), response.TxResponse.Code)
	t.Log(response.TxResponse.String())
}
