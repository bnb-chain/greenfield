package gensp_test

// todo: fix this
//package gensp_test
//
//import (
//	"encoding/hex"
//	"encoding/json"
//	"fmt"
//	"math/rand"
//	"testing"
//	"time"
//
//	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
//	"github.com/cosmos/cosmos-sdk/simapp"
//	"github.com/cosmos/cosmos-sdk/simapp/helpers"
//	sdk "github.com/cosmos/cosmos-sdk/types"
//	testutil2 "github.com/cosmos/cosmos-sdk/types/module/testutil"
//	"github.com/cosmos/cosmos-sdk/x/bank/testutil"
//	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
//	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
//	"github.com/evmos/ethermint/crypto/ethsecp256k1"
//	"github.com/prysmaticlabs/prysm/crypto/bls"
//	"github.com/stretchr/testify/suite"
//
//	"github.com/bnb-chain/greenfield/x/gensp"
//	gensptypes "github.com/bnb-chain/greenfield/x/gensp/types"
//)
//
//var (
//	priv1, _ = ethsecp256k1.GenerateKey()
//	priv2, _ = ethsecp256k1.GenerateKey()
//	pk1      = priv1.PubKey()
//	pk2      = priv2.PubKey()
//	addr1    = sdk.AccAddress(pk1.Address())
//	addr2    = sdk.AccAddress(pk2.Address())
//	desc     = stakingtypes.NewDescription("testname", "", "", "", "")
//	comm     = stakingtypes.CommissionRates{}
//)
//
//// GenTxTestSuite is a test suite to be used with gentx tests.
//type GenTxTestSuite struct {
//	suite.Suite
//
//	ctx            sdk.Context
//	app            *simapp.SimApp
//	encodingConfig testutil2.TestEncodingConfig
//
//	msg1, msg2 *stakingtypes.MsgCreateValidator
//}
//
//func (suite *GenTxTestSuite) SetupTest() {
//	checkTx := false
//	app := simapp.Setup(suite.T(), checkTx, true)
//	suite.ctx = app.BaseApp.NewContext(checkTx, tmproto.Header{ChainID: simapp.DefaultChainId})
//	suite.app = app
//	suite.encodingConfig = testutil2.MakeTestEncodingConfig()
//
//	var err error
//	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
//	one := sdk.OneInt()
//	blsSecretKey, _ := bls.RandKey()
//	blsPk := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
//	suite.msg1, err = stakingtypes.NewMsgCreateValidator(
//		sdk.AccAddress(pk1.Address()), pk1,
//		amount, desc, comm, one,
//		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()),
//		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()), blsPk)
//	suite.NoError(err)
//	suite.msg2, err = stakingtypes.NewMsgCreateValidator(
//		sdk.AccAddress(pk2.Address()), pk1,
//		amount, desc, comm, one,
//		sdk.AccAddress(pk2.Address()), sdk.AccAddress(pk2.Address()),
//		sdk.AccAddress(pk2.Address()), sdk.AccAddress(pk2.Address()), blsPk)
//	suite.NoError(err)
//}
//
//func (suite *GenTxTestSuite) setAccountBalance(addr sdk.AccAddress, amount int64) json.RawMessage {
//	acc := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, addr)
//	suite.app.AccountKeeper.SetAccount(suite.ctx, acc)
//
//	err := testutil.FundAccount(suite.app.BankKeeper, suite.ctx, addr, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, amount)})
//	suite.Require().NoError(err)
//
//	bankGenesisState := suite.app.BankKeeper.ExportGenesis(suite.ctx)
//	bankGenesis, err := suite.encodingConfig.Amino.MarshalJSON(bankGenesisState) // TODO switch this to use Marshaler
//	suite.Require().NoError(err)
//
//	return bankGenesis
//}
//
//func (suite *GenTxTestSuite) TestSetGenTxsInAppGenesisState() {
//	var (
//		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
//		genTxs    []sdk.Tx
//	)
//
//	testCases := []struct {
//		msg      string
//		malleate func()
//		expPass  bool
//	}{
//		{
//			"one genesis transaction",
//			func() {
//				err := txBuilder.SetMsgs(suite.msg1)
//				suite.Require().NoError(err)
//				tx := txBuilder.GetTx()
//				genTxs = []sdk.Tx{tx}
//			},
//			true,
//		},
//		{
//			"two genesis transactions",
//			func() {
//				err := txBuilder.SetMsgs(suite.msg1, suite.msg2)
//				suite.Require().NoError(err)
//				tx := txBuilder.GetTx()
//				genTxs = []sdk.Tx{tx}
//			},
//			true,
//		},
//	}
//
//	for _, tc := range testCases {
//		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
//			suite.SetupTest()
//			cdc := suite.encodingConfig.Codec
//			txJSONEncoder := suite.encodingConfig.TxConfig.TxJSONEncoder()
//
//			tc.malleate()
//			appGenesisState, err := gensp.SetGenTxsInAppGenesisState(cdc, txJSONEncoder, make(map[string]json.RawMessage), genTxs)
//
//			if tc.expPass {
//				suite.Require().NoError(err)
//				suite.Require().NotNil(appGenesisState[gensptypes.ModuleName])
//
//				var genesisState gensptypes.GenesisState
//				err := cdc.UnmarshalJSON(appGenesisState[gensptypes.ModuleName], &genesisState)
//				suite.Require().NoError(err)
//				suite.Require().NotNil(genesisState.GenspTxs)
//			} else {
//				suite.Require().Error(err)
//			}
//		})
//	}
//}
//
//func (suite *GenTxTestSuite) TestValidateAccountInGenesis() {
//	var (
//		appGenesisState = make(map[string]json.RawMessage)
//		coins           sdk.Coins
//	)
//
//	testCases := []struct {
//		msg      string
//		malleate func()
//		expPass  bool
//	}{
//		{
//			"no accounts",
//			func() {
//				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
//			},
//			false,
//		},
//		{
//			"account without balance in the genesis state",
//			func() {
//				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
//				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr2, 50)
//			},
//			false,
//		},
//		{
//			"account without enough funds of default bond denom",
//			func() {
//				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}
//				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr1, 25)
//			},
//			false,
//		},
//		{
//			"account with enough funds of default bond denom",
//			func() {
//				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
//				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance(addr1, 25)
//			},
//			true,
//		},
//	}
//	for _, tc := range testCases {
//		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
//			suite.SetupTest()
//			cdc := suite.encodingConfig.Codec
//
//			suite.app.StakingKeeper.SetParams(suite.ctx, stakingtypes.DefaultParams())
//			stakingGenesisState := suite.app.StakingKeeper.ExportGenesis(suite.ctx)
//			suite.Require().Equal(stakingGenesisState.Params, stakingtypes.DefaultParams())
//			stakingGenesis, err := cdc.MarshalJSON(stakingGenesisState) // TODO switch this to use Marshaler
//			suite.Require().NoError(err)
//			appGenesisState[stakingtypes.ModuleName] = stakingGenesis
//
//			tc.malleate()
//			err = gensp.ValidateAccountInGenesis(
//				appGenesisState, banktypes.GenesisBalancesIterator{},
//				addr1, coins, cdc,
//			)
//
//			if tc.expPass {
//				suite.Require().NoError(err)
//			} else {
//				suite.Require().Error(err)
//			}
//		})
//	}
//}
//
//func (suite *GenTxTestSuite) TestDeliverGenTxs() {
//	var (
//		genTxs    []json.RawMessage
//		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
//	)
//
//	testCases := []struct {
//		msg      string
//		malleate func()
//		expPass  bool
//	}{
//		{
//			"no signature supplied",
//			func() {
//				err := txBuilder.SetMsgs(suite.msg1)
//				suite.Require().NoError(err)
//
//				genTxs = make([]json.RawMessage, 1)
//				tx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
//				suite.Require().NoError(err)
//				genTxs[0] = tx
//			},
//			false,
//		},
//		{
//			"success",
//			func() {
//				_ = suite.setAccountBalance(addr1, 50)
//				_ = suite.setAccountBalance(addr2, 1)
//
//				r := rand.New(rand.NewSource(time.Now().UnixNano()))
//				msg := banktypes.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)})
//				tx, err := helpers.GenSignedMockTx(
//					r,
//					suite.encodingConfig.TxConfig,
//					[]sdk.Msg{msg},
//					sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)},
//					helpers.DefaultGenTxGas,
//					suite.ctx.ChainID(),
//					[]uint64{7},
//					[]uint64{0},
//					priv1,
//				)
//				suite.Require().NoError(err)
//
//				genTxs = make([]json.RawMessage, 1)
//				genTx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(tx)
//				suite.Require().NoError(err)
//				genTxs[0] = genTx
//			},
//			true,
//		},
//	}
//
//	for _, tc := range testCases {
//		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
//			suite.SetupTest()
//
//			tc.malleate()
//
//			if tc.expPass {
//				suite.Require().NotPanics(func() {
//					// nolint: errcheck
//					gensp.DeliverGenTxs(
//						suite.ctx, genTxs, suite.app.StakingKeeper, suite.app.BaseApp.DeliverTx,
//						suite.encodingConfig.TxConfig,
//					)
//				})
//			} else {
//				_, err := gensp.DeliverGenTxs(
//					suite.ctx, genTxs, suite.app.StakingKeeper, suite.app.BaseApp.DeliverTx,
//					suite.encodingConfig.TxConfig,
//				)
//
//				suite.Require().Error(err)
//			}
//		})
//	}
//}
//
//func TestGenTxTestSuite(t *testing.T) {
//	suite.Run(t, new(GenTxTestSuite))
//}
