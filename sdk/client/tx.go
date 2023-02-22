package client

import (
	"context"

	"github.com/cosmos/cosmos-sdk/client"
	clitx "github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	xauthsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield/sdk/types"
)

type TransactionClient interface {
	BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error)
	SimulateTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error)
	SignTx(msgs []sdk.Msg, txOpt *types.TxOption) ([]byte, error)
}

// BroadcastTx signs and broadcasts a tx with simulated gas(if not provided in txOpt)
func (c *GreenfieldClient) BroadcastTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {
	txConfig := authtx.NewTxConfig(c.codec, []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()

	// txBuilder holds tx info
	if err := c.constructTxWithGasInfo(msgs, txOpt, txConfig, txBuilder); err != nil {
		return nil, err
	}

	// sign a tx
	txSignedBytes, err := c.signTx(txConfig, txBuilder)
	if err != nil {
		return nil, err
	}

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	if txOpt != nil && txOpt.Mode != nil {
		mode = *txOpt.Mode
	}
	txRes, err := c.TxClient.BroadcastTx(
		context.Background(),
		&tx.BroadcastTxRequest{
			Mode:    mode,
			TxBytes: txSignedBytes,
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return txRes, nil
}

// SimulateTx simulates a tx and gets Gas info
func (c *GreenfieldClient) SimulateTx(msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	txConfig := authtx.NewTxConfig(c.codec, []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()
	err := c.constructTx(msgs, txOpt, txBuilder)
	if err != nil {
		return nil, err
	}
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	simulateResponse, err := c.simulateTx(txBytes, opts...)
	if err != nil {
		return nil, err
	}
	return simulateResponse, nil
}

func (c *GreenfieldClient) simulateTx(txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	simulateResponse, err := c.TxClient.Simulate(
		context.Background(),
		&tx.SimulateRequest{
			TxBytes: txBytes,
		},
		opts...,
	)
	if err != nil {
		return nil, err
	}
	return simulateResponse, nil
}

// SignTx signs the tx with private key and returns bytes
func (c *GreenfieldClient) SignTx(msgs []sdk.Msg, txOpt *types.TxOption) ([]byte, error) {
	txConfig := authtx.NewTxConfig(c.codec, []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()
	if err := c.constructTxWithGasInfo(msgs, txOpt, txConfig, txBuilder); err != nil {
		return nil, err
	}
	return c.signTx(txConfig, txBuilder)
}

func (c *GreenfieldClient) signTx(txConfig client.TxConfig, txBuilder client.TxBuilder) ([]byte, error) {
	km, err := c.GetKeyManager()
	if err != nil {
		return nil, err
	}
	account, err := c.getAccount()
	if err != nil {
		return nil, err
	}
	signerData := xauthsigning.SignerData{
		ChainID:       c.chainId,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      account.GetSequence(),
	}
	sig, err := clitx.SignWithPrivKey(signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		km.GetPrivKey(),
		txConfig,
		account.GetSequence(),
	)
	if err != nil {
		return nil, err
	}
	err = txBuilder.SetSignatures(sig)
	if err != nil {
		return nil, err
	}
	txSignedBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return nil, err
	}
	return txSignedBytes, nil
}

// setSingerInfo gathers the signer info by doing "empty signature" hack, and inject it into txBuilder
func (c *GreenfieldClient) setSingerInfo(txBuilder client.TxBuilder) error {
	km, err := c.GetKeyManager()
	if err != nil {
		return err
	}
	account, err := c.getAccount()
	if err != nil {
		return err
	}
	sig := signing.SignatureV2{
		PubKey: km.GetPrivKey().PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_EIP_712,
		},
		Sequence: account.GetSequence(),
	}
	if err := txBuilder.SetSignatures(sig); err != nil {
		return err
	}
	return nil
}

func (c *GreenfieldClient) constructTx(msgs []sdk.Msg, txOpt *types.TxOption, txBuilder client.TxBuilder) error {
	for _, m := range msgs {
		if err := m.ValidateBasic(); err != nil {
			return err
		}
	}

	if err := txBuilder.SetMsgs(msgs...); err != nil {
		return err
	}
	if txOpt != nil {
		if txOpt.Memo != "" {
			txBuilder.SetMemo(txOpt.Memo)
		}
		if !txOpt.FeePayer.Empty() {
			txBuilder.SetFeePayer(txOpt.FeePayer)
		}
	}
	// inject signer info into txBuilder, it is needed for simulating and signing
	return c.setSingerInfo(txBuilder)
}

func (c *GreenfieldClient) constructTxWithGasInfo(msgs []sdk.Msg, txOpt *types.TxOption, txConfig client.TxConfig, txBuilder client.TxBuilder) error {
	// construct a tx with txOpt excluding GasLimit and
	if err := c.constructTx(msgs, txOpt, txBuilder); err != nil {
		return err
	}
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}
	simulateRes, err := c.simulateTx(txBytes)
	if err != nil {
		return err
	}

	gasLimit := simulateRes.GasInfo.GetGasUsed()
	if txOpt != nil && txOpt.GasLimit != 0 {
		gasLimit = txOpt.GasLimit
	}
	gasPrice, err := sdk.ParseCoinsNormalized(simulateRes.GasInfo.GetMinGasPrices())
	if err != nil {
		return err
	}
	if gasPrice.IsZero() {
		return types.SimulatedGasPriceError
	}
	feeAmount := sdk.NewCoins(
		sdk.NewInt64Coin(
			types.Denom,
			sdk.NewInt(int64(gasLimit)).Mul(gasPrice[0].Amount).Int64()),
	)
	if txOpt != nil && !txOpt.FeeAmount.IsZero() {
		feeAmount = txOpt.FeeAmount
	}
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(feeAmount)
	return nil
}

func (c *GreenfieldClient) getAccount() (authtypes.AccountI, error) {
	km, err := c.GetKeyManager()
	if err != nil {
		return nil, err
	}
	acct, err := c.AuthQueryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: km.GetAddr().String()})
	if err != nil {
		return nil, err
	}
	var account authtypes.AccountI
	if err := c.codec.InterfaceRegistry().UnpackAny(acct.Account, &account); err != nil {
		return nil, err
	}
	return account, nil
}
