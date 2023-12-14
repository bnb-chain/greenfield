package ante_test

import (
	"encoding/hex"
	"fmt"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"

	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/sdk/keys"
)

func (suite *AnteTestSuite) TestAnteHandler() {
	privKey, err := keys.NewPrivateKeyManager(test.TEST_PRIVATE_KEY)
	pubKey := privKey.PubKey()
	addr := privKey.GetAddr()
	suite.Require().NoError(err)
	fmt.Println("Sender Private Key: ", privKey.String())
	fmt.Printf("Sender Public Key: %x\n", pubKey.Bytes())
	fmt.Printf("Sender Address: %s\n", addr.String())

	setup := func() {
		suite.SetupTest()

		acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr.Bytes())
		suite.Require().NoError(acc.SetSequence(1))
		suite.app.AccountKeeper.SetAccount(suite.ctx, acc)

		// get coins from validator
		bz, _ := hex.DecodeString(test.TEST_PUBKEY)
		faucetPubKey := &ethsecp256k1.PubKey{Key: bz}
		err := suite.app.BankKeeper.SendCoins(suite.ctx, faucetPubKey.Address().Bytes(), acc.GetAddress(), sdk.Coins{sdk.Coin{
			Denom:  test.TEST_TOKEN_NAME,
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
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgDelegate",
			func() sdk.Tx {
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgDelegate(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgCreateValidator",
			func() sdk.Tx {
				gas := uint64(2e8)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712MsgCreateValidator(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgGrantAllowance",
			func() sdk.Tx {
				gas := uint64(16e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712GrantAllowance(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgEditValidator",
			func() sdk.Tx {
				gas := uint64(2e7)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712MsgEditValidator(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgSubmitProposalV1",
			func() sdk.Tx {
				gas := uint64(2e8)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSubmitProposalV1(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgGrant",
			func() sdk.Tx {
				gas := uint64(16e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgGrant(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"success - DeliverTx EIP712 signed Cosmos Tx MsgCreateBucket",
			func() sdk.Tx {
				gas := uint64(16e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgCreateBucket(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, true,
		},
		{
			"fails - DeliverTx legacy msg MsgSubmitProposal v1beta",
			func() sdk.Tx {
				gas := uint64(2000000000)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				deposit := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(20)))
				txBuilder := suite.CreateTestEIP712SubmitProposal(addr, privKey, test.TEST_CHAIN_ID, gas, fee, deposit)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx unregistered msg type MsgSubmitEvidence",
			func() sdk.Tx {
				gas := uint64(2000000000)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712MsgSubmitEvidence(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with wrong Chain ID",
			func() sdk.Tx {
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(addr, privKey, "ethermint_9002-1", gas, fee)
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with different gas fees",
			func() sdk.Tx {
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				txBuilder.SetGasLimit(uint64(300000))
				txBuilder.SetFeeAmount(sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdkmath.NewInt(30))))
				return txBuilder.GetTx()
			}, true, false, false,
		},
		{
			"fails - DeliverTx EIP712 signed Cosmos Tx with empty signature",
			func() sdk.Tx {
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
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
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, addr)
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
				gas := uint64(12e3)
				fee := sdk.NewCoins(sdk.NewCoin(test.TEST_TOKEN_NAME, sdk.NewIntFromUint64(gas)))
				txBuilder := suite.CreateTestEIP712TxBuilderMsgSend(addr, privKey, test.TEST_CHAIN_ID, gas, fee)
				nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, addr)
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
