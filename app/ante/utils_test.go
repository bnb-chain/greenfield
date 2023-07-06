package ante_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmtypes "github.com/cometbft/cometbft/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	simtestutil "github.com/cosmos/cosmos-sdk/testutil/sims"
	sdk "github.com/cosmos/cosmos-sdk/types"
	signingtypes "github.com/cosmos/cosmos-sdk/types/tx/signing"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
	"github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	evtypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/app"
	"github.com/bnb-chain/greenfield/app/ante"
	"github.com/bnb-chain/greenfield/app/params"
	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/sdk/keys"
)

type AnteTestSuite struct {
	suite.Suite

	ctx         sdk.Context
	app         *app.App
	clientCtx   client.Context
	anteHandler sdk.AnteHandler
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{})
}

func (suite *AnteTestSuite) SetupTest() {
	var encCfg params.EncodingConfig
	suite.app, encCfg, _ = NewApp(baseapp.SetChainID("greenfield_9000-1"))

	suite.ctx = suite.app.NewContext(false, tmproto.Header{Height: 2, ChainID: "greenfield_9000-1", Time: time.Now().UTC()})
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.OneInt()))) // set to 1 stake

	infCtx := suite.ctx.WithGasMeter(storetypes.NewInfiniteGasMeter())
	err := suite.app.AccountKeeper.SetParams(infCtx, authtypes.DefaultParams())
	suite.Require().NoError(err)

	suite.clientCtx = client.Context{}.WithTxConfig(encCfg.TxConfig)

	anteHandler, _ := ante.NewAnteHandler(ante.HandlerOptions{
		AccountKeeper:   suite.app.AccountKeeper,
		BankKeeper:      suite.app.BankKeeper,
		FeegrantKeeper:  suite.app.FeeGrantKeeper,
		SignModeHandler: encCfg.TxConfig.SignModeHandler(),
		GashubKeeper:    suite.app.GashubKeeper,
	})
	suite.anteHandler = anteHandler
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgSend(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	recipient := core.GenRandomAddr()
	msgSend := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1))))
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSend)
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgDelegate(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	validator := core.GenRandomAddr()
	msgSend := stakingtypes.NewMsgDelegate(from, validator, sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(20)))
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSend)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgCreateValidator(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	privEd := ed25519.GenPrivKey()
	blsSecretKey, _ := bls.RandKey()
	blsPubkey := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		from,
		privEd.PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20)),
		stakingtypes.NewDescription("moniker", "identity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
		from,
		from,
		from,
		from,
		blsPubkey,
	)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgCreate)
}

func (suite *AnteTestSuite) CreateTestEIP712SubmitProposal(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount, deposit sdk.Coins) client.TxBuilder {
	proposal, ok := govtypes.ContentFromProposalType("proposal", "description", govtypes.ProposalTypeText)
	suite.Require().True(ok)
	msgSubmit, err := govtypes.NewMsgSubmitProposal(proposal, deposit, from)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSubmit)
}

func (suite *AnteTestSuite) CreateTestEIP712GrantAllowance(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	spendLimit := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))
	threeHours := time.Now().Add(3 * time.Hour)
	basic := &feegrant.BasicAllowance{
		SpendLimit: spendLimit,
		Expiration: &threeHours,
	}
	granted := core.GenRandomAddr()
	grantedAddr := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, granted.Bytes())
	msgGrant, err := feegrant.NewMsgGrantAllowance(basic, from, grantedAddr.GetAddress())
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgGrant)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgEditValidator(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	newRelayerAddr := core.GenRandomAddr()
	newChallengerAddr := core.GenRandomAddr()
	blsSecretKey, _ := bls.RandKey()
	blsPk := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
	msgEdit := stakingtypes.NewMsgEditValidator(
		from,
		stakingtypes.NewDescription("moniker", "identity", "website", "security_contract", "details"),
		nil,
		nil,
		newRelayerAddr,
		newChallengerAddr,
		blsPk,
		nil,
	)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgEdit)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgSubmitEvidence(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	pk := ed25519.GenPrivKey()
	msgEvidence, err := evtypes.NewMsgSubmitEvidence(from, &evtypes.Equivocation{
		Height:           11,
		Time:             time.Now().UTC(),
		Power:            100,
		ConsensusAddress: pk.PubKey().Address().String(),
	})
	suite.Require().NoError(err)

	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgEvidence)
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgSubmitProposalV1(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		from,
		privEd.PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20)),
		stakingtypes.NewDescription("moniker", "indentity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
		from,
		from,
		from,
		from,
		"test",
	)
	suite.Require().NoError(err)
	msgSubmitProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgCreate},
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))},
		from.String(),
		"test", "test", "test",
	)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSubmitProposal)
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgGrant(from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	allowed := core.GenRandomAddr()
	stakeAuthorization, err := stakingtypes.NewStakeAuthorization(
		[]sdk.AccAddress{allowed},
		nil,
		1,
		nil,
	)
	suite.Require().NoError(err)
	msgGrant, err := authz.NewMsgGrant(from, allowed, stakeAuthorization, nil)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgGrant)
}

func (suite *AnteTestSuite) CreateTestEIP712CosmosTxBuilder(
	from sdk.AccAddress, priv keys.KeyManager, chainId string, gas uint64, gasAmount sdk.Coins, msg sdk.Msg,
) client.TxBuilder {
	nonce, err := suite.app.AccountKeeper.GetSequence(suite.ctx, from)
	suite.Require().NoError(err)
	acc, err := authante.GetSignerAcc(suite.ctx, suite.app.AccountKeeper, from)
	suite.Require().NoError(err)

	txBuilder := suite.clientCtx.TxConfig.NewTxBuilder()

	txBuilder.SetFeeAmount(gasAmount)
	txBuilder.SetGasLimit(gas)

	err = txBuilder.SetMsgs(msg)
	suite.Require().NoError(err)

	signerData := signing.SignerData{
		Address:       from.String(),
		ChainID:       chainId,
		AccountNumber: acc.GetAccountNumber(),
		Sequence:      nonce,
		PubKey:        acc.GetPubKey(),
	}

	msgTypes, signDoc, err := tx.GetMsgTypes(signerData, txBuilder.GetTx(), big.NewInt(9000))
	suite.Require().NoError(err)

	typedData, err := tx.WrapTxToTypedData(9000, signDoc, msgTypes)
	suite.Require().NoError(err)

	typedDataJson, _ := json.MarshalIndent(typedData, "", "  ")
	fmt.Println("Typed data:\n", string(typedDataJson))

	sigHash, err := suite.clientCtx.TxConfig.SignModeHandler().GetSignBytes(signingtypes.SignMode_SIGN_MODE_EIP_712, signerData, txBuilder.GetTx())
	suite.Require().NoError(err)
	fmt.Printf("SigHash: %x\n", sigHash)

	// Sign typedData
	signature, err := priv.Sign(sigHash)
	suite.Require().NoError(err)
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	fmt.Printf("Signature: %x\n", signature)

	sigsV2 := signingtypes.SignatureV2{
		PubKey: priv.PubKey(),
		Data: &signingtypes.SingleSignatureData{
			SignMode:  signingtypes.SignMode_SIGN_MODE_EIP_712,
			Signature: signature,
		},
		Sequence: nonce,
	}

	err = txBuilder.SetSignatures(sigsV2)
	suite.Require().NoError(err)
	return txBuilder
}

func NewApp(options ...func(baseApp *baseapp.BaseApp)) (*app.App, params.EncodingConfig, error) {
	// create public key
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	bz, _ := hex.DecodeString(test.TEST_PUBKEY)
	faucetPubKey := &ethsecp256k1.PubKey{Key: bz}

	acc := authtypes.NewBaseAccount(faucetPubKey.Address().Bytes(), faucetPubKey, 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	encCfg := app.MakeEncodingConfig()

	nApp := app.New(
		logger,
		db,
		nil,
		true,
		app.DefaultNodeHome,
		0,
		encCfg,
		&app.AppConfig{CrossChain: app.NewDefaultAppConfig().CrossChain},
		simtestutil.EmptyAppOptions{},
		options...,
	)

	genesisState := app.NewDefaultGenesisState(encCfg.Marshaler)
	genesisState, _ = simtestutil.GenesisStateWithValSet(nApp.AppCodec(), genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)

	stateBytes, _ := json.MarshalIndent(genesisState, "", "  ")

	// Initialize the chain
	nApp.InitChain(
		abci.RequestInitChain{
			ChainId:       "greenfield_9000-1",
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)

	return nApp, encCfg, nil
}
