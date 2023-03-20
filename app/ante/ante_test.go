package ante_test

import (
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"

	"github.com/bnb-chain/greenfield/sdk/client/test"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
)

func (suite *AnteTestSuite) TestAnteHandler() {
	var acc authtypes.AccountI
	addr, privKey := tests.NewAddrKey()
	fmt.Printf("Sender Private Key: %x\n", privKey.Bytes())
	fmt.Printf("Sender Public Key: %x\n", privKey.PubKey().Bytes())
	fmt.Printf("Sender Address: %s\n", addr.String())

	setup := func() {
		suite.SetupTest()

		acc = suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
		suite.Require().NoError(acc.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

		// get coins from validator
		bz, _ := hex.DecodeString(test.TEST_PUBKEY)
		faucetPubKey := &ethsecp256k1.PubKey{Key: bz}
		err := suite.app.BankKeeper.SendCoins(suite.ctx, faucetPubKey.Address().Bytes(), acc.GetAddress(), sdk.Coins{sdk.Coin{
			Denom:  sdk.DefaultBondDenom,
			Amount: sdk.NewInt(100000000000000),
		}})
		if err != nil {
			panic(err)
		}
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
			"success - DeliverTx EIP712 signed Cosmos Tx MsgSend",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgDelegate",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgDelegate(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgCreateValidator",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(2e8)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712MsgCreateValidator(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgGrantAllowance",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(16e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712GrantAllowance(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgEditValidator",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(2e7)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712MsgEditValidator(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgSubmitProposalV1",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(2e8)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSubmitProposalV1(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgGrant",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(16e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgGrant(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"fails - DeliverTx legacy msg MsgSubmitProposal v1beta",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(2000000000)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				deposit := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(20)))
				txBuilder := suite.CreateTestEIP712SubmitProposal(from, privKey, "greenfield_9000-1", gas, fee, deposit)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx unregistered msg type MsgSubmitEvidence",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(2000000000)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712MsgSubmitEvidence(from, privKey, "greenfield_9000-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "ethermint_9002-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, fee)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with empty signature",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, fee)
				sigsV2 := signing.SignatureV2{
					PubKey: privKey.PubKey(),
				}
				// nolint:errcheck
				txBuilder.SetSignatures(sigsV2)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid sequence",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, fee)
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
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with invalid signMode",
			func() sdk.Tx {
				from := acc.GetAddress()
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(from, privKey, "greenfield_9000-1", gas, fee)
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
			}, true, false, false,
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
