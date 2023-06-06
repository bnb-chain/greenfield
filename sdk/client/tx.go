package client

import (
	"context"

	"github.com/bnb-chain/greenfield/sdk/keys"

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
	GetNonce() (uint64, error)
	GetNonceByAddr(addr sdk.AccAddress) (uint64, error)
	GetAccountByAddr(addr sdk.AccAddress) (authtypes.AccountI, error)
}

// BroadcastTx signs and broadcasts a tx with simulated gas(if not provided in txOpt)
func (c *GreenfieldClient) BroadcastTx(ctx context.Context, msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.BroadcastTxResponse, error) {
	txConfig := authtx.NewTxConfig(c.codec, []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()

	// txBuilder holds tx info
	if err := c.constructTxWithGasInfo(ctx, msgs, txOpt, txConfig, txBuilder); err != nil {
		return nil, err
	}

	// sign a tx
	txSignedBytes, err := c.signTx(ctx, txConfig, txBuilder, txOpt)
	if err != nil {
		return nil, err
	}

	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	if txOpt != nil && txOpt.Mode != nil {
		mode = *txOpt.Mode
	}
	txRes, err := c.TxClient.BroadcastTx(
		ctx,
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
func (c *GreenfieldClient) SimulateTx(ctx context.Context, msgs []sdk.Msg, txOpt *types.TxOption, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
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
	simulateResponse, err := c.simulateTx(ctx, txBytes, opts...)
	if err != nil {
		return nil, err
	}
	return simulateResponse, nil
}

func (c *GreenfieldClient) simulateTx(ctx context.Context, txBytes []byte, opts ...grpc.CallOption) (*tx.SimulateResponse, error) {
	simulateResponse, err := c.TxClient.Simulate(
		ctx,
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
func (c *GreenfieldClient) SignTx(ctx context.Context, msgs []sdk.Msg, txOpt *types.TxOption) ([]byte, error) {
	txConfig := authtx.NewTxConfig(c.codec, []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	txBuilder := txConfig.NewTxBuilder()
	if err := c.constructTxWithGasInfo(ctx, msgs, txOpt, txConfig, txBuilder); err != nil {
		return nil, err
	}
	return c.signTx(ctx, txConfig, txBuilder, txOpt)
}

func (c *GreenfieldClient) signTx(ctx context.Context, txConfig client.TxConfig, txBuilder client.TxBuilder, txOpt *types.TxOption) ([]byte, error) {

	var km keys.KeyManager
	var err error

	if txOpt != nil && txOpt.OverrideKeyManager != nil {
		km = *txOpt.OverrideKeyManager
	} else {
		km, err = c.GetKeyManager()
		if err != nil {
			return nil, err
		}
	}

	account, err := c.GetAccountByAddr(km.GetAddr())
	if err != nil {
		return nil, err
	}
	nonce := account.GetSequence()
	if txOpt != nil && txOpt.Nonce != 0 {
		nonce = txOpt.Nonce
	}

	signerData := xauthsigning.SignerData{
		ChainID:       c.chainId,
		AccountNumber: account.GetAccountNumber(),
		Sequence:      nonce,
	}
	sig, err := clitx.SignWithPrivKey(
		signing.SignMode_SIGN_MODE_EIP_712,
		signerData,
		txBuilder,
		km,
		txConfig,
		nonce,
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
func (c *GreenfieldClient) setSingerInfo(txBuilder client.TxBuilder, txOpt *types.TxOption) error {

	var km keys.KeyManager
	var err error
	if txOpt != nil && txOpt.OverrideKeyManager != nil {
		km = *txOpt.OverrideKeyManager
	} else {
		km, err = c.GetKeyManager()
		if err != nil {
			return err
		}
	}
	account, err := c.GetAccountByAddr(km.GetAddr())
	if err != nil {
		return err
	}
	nonce := account.GetSequence()
	if txOpt != nil && txOpt.Nonce != 0 {
		nonce = txOpt.Nonce
	}
	sig := signing.SignatureV2{
		PubKey: km.PubKey(),
		Data: &signing.SingleSignatureData{
			SignMode: signing.SignMode_SIGN_MODE_EIP_712,
		},
		Sequence: nonce,
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
		if !txOpt.FeeGranter.Empty() {
			txBuilder.SetFeeGranter(txOpt.FeeGranter)
		}
		if txOpt.Tip != nil {
			txBuilder.SetTip(txOpt.Tip)
		}
	}
	// inject signer info into txBuilder, it is needed for simulating and signing
	return c.setSingerInfo(txBuilder, txOpt)
}

func (c *GreenfieldClient) constructTxWithGasInfo(ctx context.Context, msgs []sdk.Msg, txOpt *types.TxOption, txConfig client.TxConfig, txBuilder client.TxBuilder) error {
	// construct a tx with txOpt excluding GasLimit and
	if err := c.constructTx(msgs, txOpt, txBuilder); err != nil {
		return err
	}
	txBytes, err := txConfig.TxEncoder()(txBuilder.GetTx())
	if err != nil {
		return err
	}

	if txOpt != nil && txOpt.NoSimulate {
		isFeeAmtZero, err := isFeeAmountZero(txOpt.FeeAmount)
		if err != nil {
			return err
		}
		if txOpt.GasLimit == 0 || isFeeAmtZero {
			return types.GasInfoNotProvidedError
		}
		txBuilder.SetGasLimit(txOpt.GasLimit)
		txBuilder.SetFeeAmount(txOpt.FeeAmount)
		return nil
	}

	simulateRes, err := c.simulateTx(ctx, txBytes)
	if err != nil {
		return err
	}
	gasLimit := simulateRes.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateRes.GasInfo.GetMinGasPrice())
	if err != nil {
		return err
	}
	if gasPrice.IsNil() || gasPrice.IsZero() {
		return types.SimulatedGasPriceError
	}
	feeAmount := sdk.NewCoins(
		sdk.NewCoin(gasPrice.Denom, gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit)))), // gasPrice * gasLimit
	)
	txBuilder.SetGasLimit(gasLimit)
	txBuilder.SetFeeAmount(feeAmount)
	return nil
}

func (c *GreenfieldClient) GetNonce() (uint64, error) {
	account, err := c.GetAccountByAddr(c.keyManager.GetAddr())
	if err != nil {
		return 0, err
	}
	return account.GetSequence(), nil
}

func (c *GreenfieldClient) GetNonceByAddr(addr sdk.AccAddress) (uint64, error) {
	account, err := c.GetAccountByAddr(addr)
	if err != nil {
		return 0, err
	}
	return account.GetSequence(), nil
}

func (c *GreenfieldClient) GetAccountByAddr(addr sdk.AccAddress) (authtypes.AccountI, error) {
	acct, err := c.AuthQueryClient.Account(context.Background(), &authtypes.QueryAccountRequest{Address: addr.String()})
	if err != nil {
		return nil, err
	}
	var account authtypes.AccountI
	if err := c.codec.InterfaceRegistry().UnpackAny(acct.Account, &account); err != nil {
		return nil, err
	}
	return account, nil
}

func isFeeAmountZero(feeAmount sdk.Coins) (bool, error) {
	if len(feeAmount) == 0 {
		return true, nil
	}
	if len(feeAmount) != 1 {
		return false, types.FeeAmountNotValidError
	}
	if feeAmount[0].Amount.IsNil() {
		return false, types.FeeAmountNotValidError
	}
	return feeAmount[0].IsZero(), nil
}
