package gensp_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"testing"
	"time"

	gensptypes "github.com/bnb-chain/greenfield/x/gensp/types"

	"cosmossdk.io/math"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	storetypes "cosmossdk.io/store/types"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	"github.com/cosmos/cosmos-sdk/testutil"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltestutil "github.com/cosmos/cosmos-sdk/x/genutil/testutil"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	"github.com/bnb-chain/greenfield/x/gensp"
)

var (
	priv1, _ = ethsecp256k1.GenPrivKey()
	priv2, _ = ethsecp256k1.GenPrivKey()
	pk1      = priv1.PubKey()
	pk2      = priv2.PubKey()
	addr1    = sdk.AccAddress(pk1.Address())
	addr2    = sdk.AccAddress(pk2.Address())
	desc     = stakingtypes.NewDescription("testname", "", "", "", "")
	comm     = stakingtypes.CommissionRates{}
)

// GenTxTestSuite is a test suite to be used with gentx tests.
type GenTxTestSuite struct {
	suite.Suite

	ctx sdk.Context

	stakingKeeper  *genutiltestutil.MockStakingKeeper
	encodingConfig moduletestutil.TestEncodingConfig
	msg1, msg2     *stakingtypes.MsgCreateValidator
}

func (suite *GenTxTestSuite) SetupTest() {
	suite.encodingConfig = moduletestutil.MakeTestEncodingConfig(genutil.AppModuleBasic{})
	key := storetypes.NewKVStoreKey("a_Store_Key")
	tkey := storetypes.NewTransientStoreKey("a_transient_store")
	suite.ctx = testutil.DefaultContext(key, tkey)

	ctrl := gomock.NewController(suite.T())
	suite.stakingKeeper = genutiltestutil.NewMockStakingKeeper(ctrl)

	stakingtypes.RegisterInterfaces(suite.encodingConfig.InterfaceRegistry)
	banktypes.RegisterInterfaces(suite.encodingConfig.InterfaceRegistry)

	var err error
	amount := sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)
	one := math.OneInt()
	blsSecretKey, _ := bls.RandKey()
	blsPk := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
	suite.msg1, err = stakingtypes.NewMsgCreateValidator(
		sdk.AccAddress(pk1.Address()), pk1,
		amount, desc, comm, one,
		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()),
		sdk.AccAddress(pk1.Address()), sdk.AccAddress(pk1.Address()), blsPk)
	suite.NoError(err)
	suite.msg2, err = stakingtypes.NewMsgCreateValidator(
		sdk.AccAddress(pk2.Address()), pk1,
		amount, desc, comm, one,
		sdk.AccAddress(pk2.Address()), sdk.AccAddress(pk2.Address()),
		sdk.AccAddress(pk2.Address()), sdk.AccAddress(pk1.Address()), blsPk)
	suite.NoError(err)
}

func (suite *GenTxTestSuite) setAccountBalance(balances []banktypes.Balance) json.RawMessage {
	bankGenesisState := banktypes.GenesisState{
		Params: banktypes.Params{DefaultSendEnabled: true},
		Balances: []banktypes.Balance{
			{
				Address: "0x0Ec700c7b488Bf0326FEF647DafB65684371f024",
				Coins:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1000000)},
			},
			{
				Address: "0x93354845030274cD4bf1686Abd60AB28EC52e1a7",
				Coins:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 2059726)},
			},
			{
				Address: "0xEe10332A13816795560dd96a0D922A193Bd08F59",
				Coins:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 100000000000000)},
			},
		},
		Supply: sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)},
	}
	bankGenesisState.Balances = append(bankGenesisState.Balances, balances...)
	for _, balance := range bankGenesisState.Balances {
		bankGenesisState.Supply.Add(balance.Coins...)
	}
	bankGenesis, err := suite.encodingConfig.Amino.MarshalJSON(bankGenesisState) // TODO switch this to use Marshaler
	suite.Require().NoError(err)

	return bankGenesis
}

func (suite *GenTxTestSuite) TestSetGenTxsInAppGenesisState() {
	var (
		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
		genTxs    []sdk.Tx
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"one genesis transaction",
			func() {
				err := txBuilder.SetMsgs(suite.msg1)
				suite.Require().NoError(err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
		{
			"two genesis transactions",
			func() {
				err := txBuilder.SetMsgs(suite.msg1, suite.msg2)
				suite.Require().NoError(err)
				tx := txBuilder.GetTx()
				genTxs = []sdk.Tx{tx}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			cdc := suite.encodingConfig.Codec
			txJSONEncoder := suite.encodingConfig.TxConfig.TxJSONEncoder()

			tc.malleate()
			appGenesisState, err := gensp.SetGenTxsInAppGenesisState(cdc, txJSONEncoder, make(map[string]json.RawMessage), genTxs)

			if tc.expPass {
				suite.Require().NoError(err)
				suite.Require().NotNil(appGenesisState[gensptypes.ModuleName])

				var genesisState gensptypes.GenesisState
				err := cdc.UnmarshalJSON(appGenesisState[gensptypes.ModuleName], &genesisState)
				suite.Require().NoError(err)
				suite.Require().NotNil(genesisState.GenspTxs)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *GenTxTestSuite) TestValidateAccountInGenesis() {
	var (
		appGenesisState = make(map[string]json.RawMessage)
		coins           sdk.Coins
	)

	testCases := []struct {
		msg      string
		malleate func()
		expPass  bool
	}{
		{
			"no accounts",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
			},
			false,
		},
		{
			"account without balance in the genesis state",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 0)}
				balances := banktypes.Balance{
					Address: addr2.String(),
					Coins:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)},
				}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance([]banktypes.Balance{balances})
			},
			false,
		},
		{
			"account without enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 50)}
				balances := banktypes.Balance{
					Address: addr1.String(),
					Coins:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)},
				}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance([]banktypes.Balance{balances})
			},
			false,
		},
		{
			"account with enough funds of default bond denom",
			func() {
				coins = sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)}
				balances := banktypes.Balance{
					Address: addr1.String(),
					Coins:   sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 25)},
				}
				appGenesisState[banktypes.ModuleName] = suite.setAccountBalance([]banktypes.Balance{balances})
			},
			true,
		},
	}
	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()
			cdc := suite.encodingConfig.Codec

			stakingGenesis, err := cdc.MarshalJSON(&stakingtypes.GenesisState{Params: stakingtypes.DefaultParams()}) // TODO switch this to use Marshaler
			suite.Require().NoError(err)
			appGenesisState[stakingtypes.ModuleName] = stakingGenesis

			tc.malleate()
			err = genutil.ValidateAccountInGenesis(
				appGenesisState, banktypes.GenesisBalancesIterator{},
				addr1, coins, cdc,
			)

			if tc.expPass {
				suite.Require().NoError(err)
			} else {
				suite.Require().Error(err)
			}
		})
	}
}

func (suite *GenTxTestSuite) TestDeliverGenTxs() {
	var (
		genTxs    []json.RawMessage
		txBuilder = suite.encodingConfig.TxConfig.NewTxBuilder()
	)

	testCases := []struct {
		msg         string
		malleate    func()
		deliverTxFn func(abci.RequestDeliverTx) abci.ResponseDeliverTx
		expPass     bool
	}{
		{
			"no signature supplied",
			func() {
				err := txBuilder.SetMsgs(suite.msg1)
				suite.Require().NoError(err)

				genTxs = make([]json.RawMessage, 1)
				tx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(txBuilder.GetTx())
				suite.Require().NoError(err)
				genTxs[0] = tx
			},
			func(_ abci.RequestDeliverTx) abci.ResponseDeliverTx {
				return abci.ResponseDeliverTx{
					Code:      sdkerrors.ErrNoSignatures.ABCICode(),
					GasWanted: int64(10000000),
					GasUsed:   int64(41913),
					Log:       "no signatures supplied",
				}
			},
			false,
		},
		{
			"success",
			func() {
				r := rand.New(rand.NewSource(time.Now().UnixNano()))
				msg := banktypes.NewMsgSend(addr1, addr2, sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 1)})
				tx, err := simtestutil.GenSignedMockTx(
					r,
					suite.encodingConfig.TxConfig,
					[]sdk.Msg{msg},
					sdk.Coins{sdk.NewInt64Coin(sdk.DefaultBondDenom, 10)},
					simtestutil.DefaultGenTxGas,
					suite.ctx.ChainID(),
					[]uint64{7},
					[]uint64{0},
					priv1,
				)
				suite.Require().NoError(err)

				genTxs = make([]json.RawMessage, 1)
				genTx, err := suite.encodingConfig.TxConfig.TxJSONEncoder()(tx)
				suite.Require().NoError(err)
				genTxs[0] = genTx
			},
			func(tx abci.RequestDeliverTx) abci.ResponseDeliverTx {
				return abci.ResponseDeliverTx{
					Code:      sdkerrors.ErrUnauthorized.ABCICode(),
					GasWanted: int64(10000000),
					GasUsed:   int64(41353),
					Log:       "signature verification failed; please verify account number (4) and chain-id (): unauthorized",
					Codespace: "sdk",
				}
			},
			true,
		},
	}

	for _, tc := range testCases {
		suite.Run(fmt.Sprintf("Case %s", tc.msg), func() {
			suite.SetupTest()

			tc.malleate()

			if tc.expPass {
				suite.Require().NotPanics(func() {
					gensp.DeliverGenTxs(
						suite.ctx, genTxs, suite.stakingKeeper, tc.deliverTxFn,
						suite.encodingConfig.TxConfig,
					)
				})
			} else {
				_, err := gensp.DeliverGenTxs(
					suite.ctx, genTxs, suite.stakingKeeper, tc.deliverTxFn,
					suite.encodingConfig.TxConfig,
				)

				suite.Require().Error(err)
			}
		})
	}
}

func TestGenTxTestSuite(t *testing.T) {
	suite.Run(t, new(GenTxTestSuite))
}
