package ante_test

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdkcryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
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
	"github.com/evmos/ethermint/crypto/ethsecp256k1"
	"github.com/evmos/ethermint/tests"
	evmtypes "github.com/evmos/ethermint/x/evm/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	"github.com/bnb-chain/greenfield/app"
	"github.com/bnb-chain/greenfield/app/ante"
	"github.com/bnb-chain/greenfield/app/params"
)

const genesisAccountPrivateKeyForTest = "02DCA3F2C6CDF541934FA043A0ADBD891968EC7B948691ABA0C3CACA59A5DAC753"

type AnteTestSuite struct {
	suite.Suite

	ctx             sdk.Context
	app             *app.App
	clientCtx       client.Context
	anteHandler     sdk.AnteHandler
	enableFeeMarket bool
	// nolint:unused
	evmParamsOption func(*evmtypes.Params)
}

func TestAnteTestSuite(t *testing.T) {
	suite.Run(t, &AnteTestSuite{})
}

func (suite *AnteTestSuite) SetupTest() {
	simapp.FlagEnabledValue = true
	simapp.FlagCommitValue = true

	var encCfg params.EncodingConfig
	suite.app, encCfg, _ = NewApp()

	suite.ctx = suite.app.NewContext(false, tmproto.Header{Height: 2, ChainID: "greenfield_9000-1", Time: time.Now().UTC()})
	suite.ctx = suite.ctx.WithMinGasPrices(sdk.NewDecCoins(sdk.NewDecCoin(sdk.DefaultBondDenom, sdk.OneInt())))
	suite.ctx = suite.ctx.WithBlockGasMeter(sdk.NewGasMeter(1000000000000000000))

	infCtx := suite.ctx.WithGasMeter(sdk.NewInfiniteGasMeter())
	suite.app.AccountKeeper.SetParams(infCtx, authtypes.DefaultParams())

	// We're using TestMsg amino encoding in some tests, so register it here.
	encCfg.Amino.RegisterConcrete(&testdata.TestMsg{}, "testdata.TestMsg", nil)

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

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgSend(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	// Build MsgSend
	recipient := tests.GenerateAddress().Bytes()
	msgSend := banktypes.NewMsgSend(from, recipient, sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(1))))
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSend)
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgDelegate(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	// Build MsgSend
	valEthAddr := tests.GenerateAddress()
	valAddr := sdk.AccAddress(valEthAddr.Bytes())
	msgSend := stakingtypes.NewMsgDelegate(from, valAddr, sdk.NewCoin(sdk.DefaultBondDenom, sdkmath.NewInt(20)))
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSend)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgCreateValidator(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	// Build MsgCreateValidator
	valAddr := sdk.AccAddress(from.Bytes())
	privEd := ed25519.GenPrivKey()
	addr1 := sdk.AccAddress(from.Bytes())
	blsSecretKey, _ := bls.RandKey()
	blsPk := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr,
		privEd.PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20)),
		stakingtypes.NewDescription("moniker", "indentity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(),
		addr1,
		addr1,
		addr1,
		blsPk,
		addr1,
	)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgCreate)
}

func (suite *AnteTestSuite) CreateTestEIP712SubmitProposal(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins, deposit sdk.Coins) client.TxBuilder {
	proposal, ok := govtypes.ContentFromProposalType("My proposal", "My description", govtypes.ProposalTypeText)
	suite.Require().True(ok)
	msgSubmit, err := govtypes.NewMsgSubmitProposal(proposal, deposit, from)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSubmit)
}

func (suite *AnteTestSuite) CreateTestEIP712GrantAllowance(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	spendLimit := sdk.NewCoins(sdk.NewInt64Coin(sdk.DefaultBondDenom, 10))
	threeHours := time.Now().Add(3 * time.Hour)
	basic := &feegrant.BasicAllowance{
		SpendLimit: spendLimit,
		Expiration: &threeHours,
	}
	granted := tests.GenerateAddress()
	grantedAddr := suite.app.AccountKeeper.NewAccountWithAddress(suite.ctx, granted.Bytes())
	msgGrant, err := feegrant.NewMsgGrantAllowance(basic, from, grantedAddr.GetAddress())
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgGrant)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgEditValidator(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	valAddr := sdk.AccAddress(from.Bytes())
	blsSecretKey, _ := bls.RandKey()
	blsPk := hex.EncodeToString(blsSecretKey.PublicKey().Marshal())
	msgEdit := stakingtypes.NewMsgEditValidator(
		valAddr,
		stakingtypes.NewDescription("moniker", "identity", "website", "security_contract", "details"),
		nil,
		nil,
		sdk.AccAddress(priv.PubKey().Address()), blsPk,
	)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgEdit)
}

func (suite *AnteTestSuite) CreateTestEIP712MsgSubmitEvidence(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
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

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgSubmitProposalV1(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	valAddr := sdk.AccAddress(tests.GenerateAddress().Bytes())
	privEd := ed25519.GenPrivKey()
	msgCreate, err := stakingtypes.NewMsgCreateValidator(
		valAddr,
		privEd.PubKey(),
		sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20)),
		stakingtypes.NewDescription("moniker", "indentity", "website", "security_contract", "details"),
		stakingtypes.NewCommissionRates(sdk.OneDec(), sdk.OneDec(), sdk.OneDec()),
		sdk.OneInt(), valAddr, valAddr, valAddr, "test", valAddr,
	)
	suite.Require().NoError(err)
	msgSubmitProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgCreate},
		sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(20))},
		sdk.AccAddress(from.Bytes()).String(),
		"test",
	)
	suite.Require().NoError(err)
	return suite.CreateTestEIP712CosmosTxBuilder(from, priv, chainId, gas, gasAmount, msgSubmitProposal)
}

func (suite *AnteTestSuite) CreateTestEIP712TxBuilderMsgGrant(from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins) client.TxBuilder {
	allowed := tests.GenerateAddress().Bytes()
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
	from sdk.AccAddress, priv cryptotypes.PrivKey, chainId string, gas uint64, gasAmount sdk.Coins, msg sdk.Msg,
) client.TxBuilder {
	var err error

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
	fmt.Println("SigHash:", hex.EncodeToString(sigHash))

	// Sign typedData
	keyringSigner := tests.NewSigner(priv)
	signature, pubkey, err := keyringSigner.SignByAddress(from, sigHash)
	suite.Require().NoError(err)
	signature[crypto.RecoveryIDOffset] += 27 // Transform V from 0/1 to 27/28 according to the yellow paper
	fmt.Println("Signature:", hex.EncodeToString(signature))

	sigsV2 := signingtypes.SignatureV2{
		PubKey: pubkey,
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

var _ sdk.Tx = &invalidTx{}

type invalidTx struct{}

func (invalidTx) GetMsgs() []sdk.Msg   { return []sdk.Msg{nil} }
func (invalidTx) ValidateBasic() error { return nil }

func genesisStateWithValSet(
	app *app.App, genesisState app.GenesisState,
	valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (app.GenesisState, error) {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = app.AppCodec().MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := sdkcryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, err
		}
		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, err
		}
		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		delegations = append(delegations, stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec()))

	}
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stakingtypes.DefaultParams(), validators, delegations)
	genesisState[stakingtypes.ModuleName] = app.AppCodec().MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(sdk.DefaultBondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(sdk.DefaultBondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(banktypes.DefaultGenesisState().Params, balances, totalSupply, []banktypes.Metadata{})
	genesisState[banktypes.ModuleName] = app.AppCodec().MustMarshalJSON(bankGenesis)

	return genesisState, nil
}

func NewApp(options ...func(baseApp *baseapp.BaseApp)) (*app.App, params.EncodingConfig, error) {
	// create public key
	privVal := mock.NewPV()
	pubKey, _ := privVal.GetPubKey()

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	bz, _ := hex.DecodeString(genesisAccountPrivateKeyForTest)
	senderPubKey := &ethsecp256k1.PubKey{Key: bz}

	acc := authtypes.NewBaseAccount(senderPubKey.Address().Bytes(), senderPubKey, 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
	}

	logger := log.NewNopLogger()
	db := dbm.NewMemDB()
	encCfg := app.MakeEncodingConfig()

	nApp := app.New(
		logger, db, nil, true, app.DefaultNodeHome, 0, encCfg, &app.AppConfig{CrossChain: app.NewDefaultAppConfig().CrossChain}, simapp.EmptyAppOptions{}, options...)

	genesisState := app.NewDefaultGenesisState(encCfg.Marshaler)
	genesisState, _ = genesisStateWithValSet(nApp, genesisState, valSet, []authtypes.GenesisAccount{acc}, balance)

	stateBytes, _ := json.MarshalIndent(genesisState, "", "  ")

	// Initialize the chain
	nApp.InitChain(
		abci.RequestInitChain{
			ChainId:       "greenfield_9000-1",
			Validators:    []abci.ValidatorUpdate{},
			AppStateBytes: stateBytes,
		},
	)
	// nApp.Commit()

	return nApp, encCfg, nil
}
