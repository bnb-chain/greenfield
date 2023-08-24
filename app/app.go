package app

import (
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	tmjson "github.com/cometbft/cometbft/libs/json"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/iavl"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/consensus"
	consensusparamkeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamtypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	"github.com/cosmos/cosmos-sdk/x/crosschain"
	crosschainkeeper "github.com/cosmos/cosmos-sdk/x/crosschain/keeper"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/gashub"
	gashubkeeper "github.com/cosmos/cosmos-sdk/x/gashub/keeper"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/oracle"
	oraclekeeper "github.com/cosmos/cosmos-sdk/x/oracle/keeper"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	"github.com/bnb-chain/greenfield/app/ante"
	appparams "github.com/bnb-chain/greenfield/app/params"
	docs "github.com/bnb-chain/greenfield/swagger"
	"github.com/bnb-chain/greenfield/types/openapiutil"
	"github.com/bnb-chain/greenfield/version"
	bridgemodule "github.com/bnb-chain/greenfield/x/bridge"
	bridgemodulekeeper "github.com/bnb-chain/greenfield/x/bridge/keeper"
	bridgemoduletypes "github.com/bnb-chain/greenfield/x/bridge/types"
	challengemodule "github.com/bnb-chain/greenfield/x/challenge"
	challengemodulekeeper "github.com/bnb-chain/greenfield/x/challenge/keeper"
	challengemoduletypes "github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/bnb-chain/greenfield/x/gensp"
	gensptypes "github.com/bnb-chain/greenfield/x/gensp/types"
	paymentmodule "github.com/bnb-chain/greenfield/x/payment"
	paymentmodulekeeper "github.com/bnb-chain/greenfield/x/payment/keeper"
	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissionmodule "github.com/bnb-chain/greenfield/x/permission"
	permissionmodulekeeper "github.com/bnb-chain/greenfield/x/permission/keeper"
	permissionmoduletypes "github.com/bnb-chain/greenfield/x/permission/types"
	spmodule "github.com/bnb-chain/greenfield/x/sp"
	spmodulekeeper "github.com/bnb-chain/greenfield/x/sp/keeper"
	spmoduletypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagemodule "github.com/bnb-chain/greenfield/x/storage"
	storagemodulekeeper "github.com/bnb-chain/greenfield/x/storage/keeper"
	storagemoduletypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmodule "github.com/bnb-chain/greenfield/x/virtualgroup"
	virtualgroupmodulekeeper "github.com/bnb-chain/greenfield/x/virtualgroup/keeper"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

const (
	Name          = "greenfield"
	ShortName     = "gnfd"
	EIP155ChainID = "9000"
	Epoch         = "1"

	// CoinType is the ETH coin type as defined in SLIP44 (https://github.com/satoshilabs/slips/blob/master/slip-0044.md)
	// In order to keep consistent with bnb smart chain
	CoinType = 60
)

func getGovProposalHandlers() []govclient.ProposalHandler {
	var govProposalHandlers []govclient.ProposalHandler

	govProposalHandlers = append(govProposalHandlers,
		paramsclient.ProposalHandler,
	)

	return govProposalHandlers
}

var (
	// DefaultNodeHome default home directories for the application daemon
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		gensp.AppModuleBasic{},
		bank.AppModuleBasic{},
		staking.AppModuleBasic{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		consensus.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		crosschain.AppModuleBasic{},
		oracle.AppModuleBasic{},
		bridgemodule.AppModuleBasic{},
		gashub.AppModuleBasic{},
		spmodule.AppModuleBasic{},
		paymentmodule.AppModuleBasic{},
		permissionmodule.AppModuleBasic{},
		virtualgroupmodule.AppModuleBasic{},
		storagemodule.AppModuleBasic{},
		challengemodule.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:         nil,
		distrtypes.ModuleName:              nil,
		stakingtypes.BondedPoolName:        {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName:     {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:                {authtypes.Burner},
		paymentmoduletypes.ModuleName:      {authtypes.Burner, authtypes.Staking},
		crosschaintypes.ModuleName:         {authtypes.Minter},
		permissionmoduletypes.ModuleName:   nil,
		bridgemoduletypes.ModuleName:       nil,
		spmoduletypes.ModuleName:           {authtypes.Staking},
		virtualgroupmoduletypes.ModuleName: nil,
	}
)

var (
	_ servertypes.Application = (*App)(nil)
	_ runtime.AppI            = (*App)(nil)
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, "."+ShortName)

	sdk.DefaultPowerReduction = sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil))
}

// App extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type App struct {
	*baseapp.BaseApp

	cdc               *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper         authkeeper.AccountKeeper
	AuthzKeeper           authzkeeper.Keeper
	BankKeeper            bankkeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	CrossChainKeeper      crosschainkeeper.Keeper
	OracleKeeper          oraclekeeper.Keeper
	GashubKeeper          gashubkeeper.Keeper
	ConsensusParamsKeeper consensusparamkeeper.Keeper

	BridgeKeeper           bridgemodulekeeper.Keeper
	SpKeeper               spmodulekeeper.Keeper
	PaymentKeeper          paymentmodulekeeper.Keeper
	ChallengeKeeper        challengemodulekeeper.Keeper
	PermissionmoduleKeeper permissionmodulekeeper.Keeper
	VirtualgroupKeeper     virtualgroupmodulekeeper.Keeper
	StorageKeeper          storagemodulekeeper.Keeper

	// mm is the module manager
	mm *module.Manager

	// sm is the simulation manager
	sm           *module.SimulationManager
	configurator module.Configurator

	// app config

	appConfig *AppConfig
}

// New returns a reference to an initialized blockchain app
func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	customAppConfig *AppConfig,
	appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *App {
	appCodec := encodingConfig.Marshaler
	cdc := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry
	legacyAmino := encodingConfig.Amino

	bApp := baseapp.NewBaseApp(
		Name,
		logger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.AppVersion)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey,
		consensusparamtypes.StoreKey,
		group.StoreKey,
		crosschaintypes.StoreKey,
		oracletypes.StoreKey,
		bridgemoduletypes.StoreKey,
		gashubtypes.StoreKey,
		spmoduletypes.StoreKey,
		virtualgroupmoduletypes.StoreKey,
		paymentmoduletypes.StoreKey,
		permissionmoduletypes.StoreKey,
		storagemoduletypes.StoreKey,
		challengemoduletypes.StoreKey,
		reconStoreKey,
	)
	tKeys := sdk.NewTransientStoreKeys(challengemoduletypes.TStoreKey, storagemoduletypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(challengemoduletypes.MemStoreKey)

	app := &App{
		BaseApp:           bApp,
		cdc:               cdc,
		appCodec:          appCodec,
		appConfig:         customAppConfig,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tKeys,
		memKeys:           memKeys,
	}

	// set the BaseApp's parameter store
	app.ConsensusParamsKeeper = consensusparamkeeper.NewKeeper(appCodec, keys[consensusparamtypes.StoreKey], authtypes.NewModuleAddress(govtypes.ModuleName).String())
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authz.ModuleName],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.BlockedModuleAccountAddrs(),
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.AuthzKeeper,
		app.BankKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.CrossChainKeeper = crosschainkeeper.NewKeeper(
		appCodec,
		keys[crosschaintypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.StakingKeeper,
		app.BankKeeper,
	)

	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec,
		keys[distrtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec,
		legacyAmino,
		keys[slashingtypes.StoreKey],
		app.StakingKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	app.OracleKeeper = oraclekeeper.NewKeeper(
		appCodec,
		keys[crosschaintypes.StoreKey],
		authtypes.FeeCollectorName,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.CrossChainKeeper,
		app.BankKeeper,
		app.StakingKeeper,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec,
		keys[feegrant.StoreKey],
		app.AccountKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	// ... other modules keepers

	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler)
	govConfig := govtypes.DefaultConfig()

	govKeeper := govkeeper.NewKeeper(
		appCodec,
		keys[govtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		app.CrossChainKeeper,
		app.MsgServiceRouter(),
		govConfig,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	govKeeper.SetLegacyRouter(govRouter)

	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	// Register the upgrade keeper
	// todo: init upgrade keeper
	app.UpgradeKeeper = upgradekeeper.NewKeeper(keys[upgradetypes.StoreKey], appCodec, homePath, app.BaseApp)

	app.BridgeKeeper = *bridgemodulekeeper.NewKeeper(
		appCodec,
		keys[bridgemoduletypes.StoreKey],
		app.BankKeeper,
		app.StakingKeeper,
		app.CrossChainKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	bridgeModule := bridgemodule.NewAppModule(appCodec, app.BridgeKeeper, app.AccountKeeper, app.BankKeeper)

	app.GashubKeeper = gashubkeeper.NewKeeper(
		appCodec,
		keys[gashubtypes.StoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	gashubModule := gashub.NewAppModule(app.GashubKeeper)

	app.SpKeeper = *spmodulekeeper.NewKeeper(
		appCodec,
		keys[spmoduletypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.AuthzKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	spModule := spmodule.NewAppModule(appCodec, app.SpKeeper, app.AccountKeeper, app.BankKeeper)

	app.PaymentKeeper = *paymentmodulekeeper.NewKeeper(
		appCodec,
		keys[paymentmoduletypes.StoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	paymentModule := paymentmodule.NewAppModule(appCodec, app.PaymentKeeper, app.AccountKeeper, app.BankKeeper)

	app.VirtualgroupKeeper = *virtualgroupmodulekeeper.NewKeeper(
		appCodec,
		keys[virtualgroupmoduletypes.StoreKey],
		tKeys[virtualgroupmoduletypes.TStoreKey],
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		app.SpKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.PaymentKeeper,
	)
	virtualgroupModule := virtualgroupmodule.NewAppModule(appCodec, app.VirtualgroupKeeper, app.SpKeeper)

	app.PermissionmoduleKeeper = *permissionmodulekeeper.NewKeeper(
		appCodec,
		keys[permissionmoduletypes.StoreKey],
		app.AccountKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	permissionModule := permissionmodule.NewAppModule(appCodec, app.PermissionmoduleKeeper, app.AccountKeeper, app.BankKeeper)

	app.StorageKeeper = *storagemodulekeeper.NewKeeper(
		appCodec,
		keys[storagemoduletypes.StoreKey],
		tKeys[storagemoduletypes.TStoreKey],
		app.AccountKeeper,
		app.SpKeeper,
		app.PaymentKeeper,
		app.PermissionmoduleKeeper,
		app.CrossChainKeeper,
		app.VirtualgroupKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	storageModule := storagemodule.NewAppModule(appCodec, app.StorageKeeper, app.AccountKeeper, app.BankKeeper, app.SpKeeper)

	app.ChallengeKeeper = *challengemodulekeeper.NewKeeper(
		appCodec,
		keys[challengemoduletypes.StoreKey],
		tKeys[challengemoduletypes.TStoreKey],
		app.BankKeeper,
		app.StorageKeeper,
		app.SpKeeper,
		app.StakingKeeper,
		app.PaymentKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)
	challengeModule := challengemodule.NewAppModule(appCodec, app.ChallengeKeeper, app.AccountKeeper, app.BankKeeper)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.

	app.mm = module.NewManager(
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		gensp.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil, nil),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, nil),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, &app.GovKeeper, app.AccountKeeper, app.BankKeeper, nil),
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, nil),
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, nil),
		upgrade.NewAppModule(app.UpgradeKeeper),
		crosschain.NewAppModule(app.CrossChainKeeper, app.BankKeeper, app.StakingKeeper),
		oracle.NewAppModule(app.OracleKeeper),
		consensus.NewAppModule(appCodec, app.ConsensusParamsKeeper),
		bridgeModule,
		gashubModule,
		spModule,
		virtualgroupModule,
		paymentModule,
		permissionModule,
		storageModule,
		challengeModule,
	)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	app.mm.SetOrderBeginBlockers(
		// upgrades should be run first
		upgradetypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		consensusparamtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		crosschaintypes.ModuleName,
		oracletypes.ModuleName,
		bridgemoduletypes.ModuleName,
		gashubtypes.ModuleName,
		spmoduletypes.ModuleName,
		virtualgroupmoduletypes.ModuleName,
		paymentmoduletypes.ModuleName,
		permissionmoduletypes.ModuleName,
		storagemoduletypes.ModuleName,
		gensptypes.ModuleName,
		challengemoduletypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		consensusparamtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		upgradetypes.ModuleName,
		crosschaintypes.ModuleName,
		oracletypes.ModuleName,
		bridgemoduletypes.ModuleName,
		gashubtypes.ModuleName,
		spmoduletypes.ModuleName,
		virtualgroupmoduletypes.ModuleName,
		paymentmoduletypes.ModuleName,
		permissionmoduletypes.ModuleName,
		storagemoduletypes.ModuleName,
		gensptypes.ModuleName,
		challengemoduletypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	app.mm.SetOrderInitGenesis(
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		gashubtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		consensusparamtypes.ModuleName,
		upgradetypes.ModuleName,
		crosschaintypes.ModuleName,
		oracletypes.ModuleName,
		bridgemoduletypes.ModuleName,
		spmoduletypes.ModuleName,
		virtualgroupmoduletypes.ModuleName,
		paymentmoduletypes.ModuleName,
		permissionmoduletypes.ModuleName,
		storagemoduletypes.ModuleName,
		gensptypes.ModuleName,
		challengemoduletypes.ModuleName,
	)

	// Uncomment if you want to set a custom migration order here.
	// app.mm.SetOrderMigrations(custom order)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// create the simulation manager and define the order of the modules for deterministic simulations
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, nil),
	}
	app.sm = module.NewSimulationManagerFromAppModules(app.mm.Modules, overrideModules)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tKeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	anteHandler, err := ante.NewAnteHandler(
		ante.HandlerOptions{
			AccountKeeper:   app.AccountKeeper,
			BankKeeper:      app.BankKeeper,
			SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
			FeegrantKeeper:  app.FeeGrantKeeper,
			GashubKeeper:    app.GashubKeeper,
		},
	)
	if err != nil {
		panic(fmt.Errorf("failed to create AnteHandler: %s", err))
	}

	app.SetAnteHandler(anteHandler)
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetUpgradeChecker(app.UpgradeKeeper.IsUpgraded)

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	err = app.RegisterUpgradeHandlers(app.ChainID(), &app.appConfig.Config)
	if err != nil {
		panic(err)
	}

	ms := app.CommitMultiStore()
	ctx := sdk.NewContext(ms, tmproto.Header{ChainID: app.ChainID(), Height: app.LastBlockHeight()}, true, app.UpgradeKeeper.IsUpgraded, app.Logger())
	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}

		// Execute the upgraded register, such as the newly added Msg type
		// ex.
		// app.GovKeeper.Router().RegisterService(...)
		err = app.UpgradeKeeper.InitUpgraded(ctx)
		if err != nil {
			panic(err)
		}
	}

	if app.IsIavlStore() {
		//enable diff for reconciliation
		bankIavl, ok := ms.GetCommitStore(keys[banktypes.StoreKey]).(*iavl.Store)
		if !ok {
			tmos.Exit("cannot convert bank store to ival store")
		}
		bankIavl.EnableDiff()
		paymentIavl, ok := ms.GetCommitStore(keys[paymentmoduletypes.StoreKey]).(*iavl.Store)
		if !ok {
			tmos.Exit("cannot convert payment store to ival store")
		}
		paymentIavl.EnableDiff()
	}

	app.initModules(ctx)

	// add eth query router
	ethRouter := app.BaseApp.EthQueryRouter()
	ethRouter.RegisterConstHandler()
	ethRouter.RegisterEthQueryBalanceHandler(app.BankKeeper, bankkeeper.EthQueryBalanceHandlerGen)

	return app
}

func (app *App) initModules(ctx sdk.Context) {
	app.initCrossChain()

	app.initBridge()
	app.initStorage()
	app.initGov()
}

func (app *App) initCrossChain() {
	app.CrossChainKeeper.SetSrcChainID(sdk.ChainID(app.appConfig.CrossChain.SrcChainId))
	app.CrossChainKeeper.SetDestBscChainID(sdk.ChainID(app.appConfig.CrossChain.DestBscChainId))
}

func (app *App) initBridge() {
	bridgemodulekeeper.RegisterCrossApps(app.BridgeKeeper)
}

func (app *App) initStorage() {
	storagemodulekeeper.RegisterCrossApps(app.StorageKeeper)
}

func (app *App) initGov() {
	err := app.GovKeeper.RegisterCrossChainSyncParamsApp()
	if err != nil {
		panic(err)
	}
}

// Name returns the name of the App
func (app *App) Name() string { return app.BaseApp.Name() }

// BeginBlocker application updates every begin block
func (app *App) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates every end block
func (app *App) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	resp := app.mm.EndBlock(ctx, req)
	if app.IsIavlStore() {
		bankIavl, _ := app.CommitMultiStore().GetCommitStore(sdk.NewKVStoreKey(banktypes.StoreKey)).(*iavl.Store)
		paymentIavl, _ := app.CommitMultiStore().GetCommitStore(sdk.NewKVStoreKey(paymentmoduletypes.StoreKey)).(*iavl.Store)

		reconCtx, _ := ctx.CacheContext()
		reconCtx = reconCtx.WithGasMeter(sdk.NewInfiniteGasMeter())
		app.reconcile(reconCtx, bankIavl, paymentIavl)
	}
	return resp
}

// InitChainer application update at chain initialization
func (app *App) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	// init cross chain channel permissions
	app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestBscChainId), bridgemoduletypes.TransferOutChannelID, sdk.ChannelAllow)
	app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestBscChainId), bridgemoduletypes.TransferInChannelID, sdk.ChannelAllow)
	app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestBscChainId), bridgemoduletypes.SyncParamsChannelID, sdk.ChannelAllow)
	app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestBscChainId), storagemoduletypes.BucketChannelId, sdk.ChannelAllow)
	app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestBscChainId), storagemoduletypes.ObjectChannelId, sdk.ChannelAllow)
	app.CrossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(app.appConfig.CrossChain.DestBscChainId), storagemoduletypes.GroupChannelId, sdk.ChannelAllow)

	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height
func (app *App) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all the app's module account addresses.
func (app *App) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// BlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
func (app *App) BlockedModuleAccountAddrs() map[string]bool {
	modAccAddrs := app.ModuleAccountAddrs()
	delete(modAccAddrs, authtypes.NewModuleAddress(govtypes.ModuleName).String())
	delete(modAccAddrs, authtypes.NewModuleAddress(distrtypes.ModuleName).String())

	return modAccAddrs
}

// LegacyAmino returns SimApp's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) LegacyAmino() *codec.LegacyAmino {
	return app.cdc
}

// AppCodec returns an app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *App) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns an InterfaceRegistry
func (app *App) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *App) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *App) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *App) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register app's OpenAPI routes.
	apiSvr.Router.Handle("/static/swagger.yaml", http.FileServer(http.FS(docs.Docs)))
	apiSvr.Router.HandleFunc("/", openapiutil.Handler(Name, "/static/swagger.yaml"))
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *App) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *App) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

// GetMaccPerms returns a copy of the module account permissions
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// SimulationManager implements the SimulationApp interface
func (app *App) SimulationManager() *module.SimulationManager {
	return app.sm
}

func (app *App) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}
