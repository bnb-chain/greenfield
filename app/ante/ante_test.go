package ante_test

import (
	"encoding/hex"
	"fmt"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
)

func (suite *AnteTestSuite) TestAnteHandler() {
	var acc authtypes.AccountI
	addr, privKey := tests.NewAddrKey()
	fmt.Println("Private Key:", hex.EncodeToString(privKey.Bytes()))
	fmt.Println("Public Key:", hex.EncodeToString(privKey.PubKey().Bytes()))
	fmt.Println("Address:", addr.String())

	setup := func() {
		suite.enableFeeMarket = false
		suite.SetupTest() // reset

		acc = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
		suite.Require().NoError(acc.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

		// get coins from validator
		bz, _ := hex.DecodeString(genesisAccountPrivateKeyForTest)
		senderPubKey := &ethsecp256k1.PubKey{Key: bz}
		_ = suite.app.BankKeeper.SendCoins(suite.ctx, senderPubKey.Address().Bytes(), acc.GetAddress(), sdk.Coins{sdk.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: sdk.NewInt(100000000000),
		}})
	}

	testCases := []struct {
		name      string
		txFn      func() sdk.Tx
		checkTx   bool
		reCheckTx bool
		expPass   bool
	}{
		// ensure all msg type could pass test
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx with MsgDelegate",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas)))
				amount := sdk.NewCoins(coinAmount)
				txBuilder := suite.CreateTestEIP712TxBuilderMsgDelegate(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 create validator",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712MsgCreateValidator(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgGrant",
			func() sdk.Tx {
				from := acc.GetAddress()
				grantee := sdk.AccAddress("_______grantee______")
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				blockTime := time.Date(1, 1, 1, 1, 1, 1, 1, time.UTC)
				expiresAt := blockTime.Add(time.Hour)
				msg, err := authz.NewMsgGrant(
					from, grantee, &banktypes.SendAuthorization{SpendLimit: gasAmount}, &expiresAt,
				)
				suite.Require().NoError(err)
				return suite.CreateTestEIP712CosmosTxBuilder(from, privKey, "greenfield_9000-1", gas, gasAmount, msg).GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgGrantAllowance",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				gasAmount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712GrantAllowance(from, privKey, "greenfield_9000-1", gas, gasAmount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 edit validator",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712MsgEditValidator(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 submit evidence",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712MsgEditValidator(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgSubmitProposalV1",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSubmitProposalV1(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		{
			"success- DeliverTx EIP712 MsgGrant",
			func() sdk.Tx {
				from := acc.GetAddress()
				coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
				amount := sdk.NewCoins(coinAmount)
				gas := uint64(200000)
				txBuilder := suite.CreateTestEIP712TxBuilderMsgGrant(from, privKey, "greenfield_9000-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, true,
		},
		// todo: fix this test after refactoring the gashub module
		//{
		//	"fails- DeliverTx MsgSubmitProposal v1beta",
		//	func() sdk.Tx {
		//		from := acc.GetAddress()
		//		coinAmount := sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))
		//		gasAmount := sdk.NewCoins(coinAmount)
		//		gas := uint64(200000)
		//		// reusing the gasAmount for deposit
		//		deposit := sdk.NewCoins(coinAmount)
		//		txBuilder := suite.CreateTestEIP712SubmitProposal(from, privKey, "greenfield_9000-1", gas, gasAmount, deposit)
		//		return txBuilder.GetTx()
		//	}, false, false, false,
		//},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9002-1", gas, amount)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, amount)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with empty signature",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, amount)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
				}
				// nolint:errcheck
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid sequence",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, amount)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_LEGACY_AMINO_JSON,
					},
					Sequence: nonce - 1,
				}
				// nolint:errcheck
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, false, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid signMode",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(200000)
				amount := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(100*int64(gas))))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, amount)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, acc.GetAddress())
				suite.Require().NoError(err)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
					Data: &signing.SingleSignatureData{
						SignMode: signing.SignMode_SIGN_MODE_UNSPECIFIED,
					},
					Sequence: nonce,
				}
				// nolint:errcheck
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, false, false, false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			setup()

			suite.ctx = suite.ctx.WithIsCheckTx(tc.checkTx).WithIsReCheckTx(tc.reCheckTx)

			_, err := suite.anteHandler(suite.ctx, tc.txFn(), false)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}
