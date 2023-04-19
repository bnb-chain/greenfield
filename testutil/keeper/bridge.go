package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	evidencetypes "cosmossdk.io/x/evidence/types"
	"cosmossdk.io/x/feegrant"
	upgradetypes "cosmossdk.io/x/upgrade/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschainkeeper "github.com/cosmos/cosmos-sdk/x/crosschain/keeper"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/app"
	"github.com/bnb-chain/greenfield/x/bridge/keeper"
	"github.com/bnb-chain/greenfield/x/bridge/types"
	storagemoduletypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     {authtypes.Minter, authtypes.Staking},
		distrtypes.ModuleName:          nil,
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		crosschaintypes.ModuleName:     {authtypes.Minter},
		types.ModuleName:               nil,
	}
)

type BridgeKeeperSuite struct {
	Cdc codec.Codec

	BridgeKeeper *keeper.Keeper

	Ctx sdk.Context

	BankKeeper *bankkeeper.BaseKeeper

	AccountKeeper authkeeper.AccountKeeper
}

func BridgeKeeper(t testing.TB) (*BridgeKeeperSuite, *keeper.Keeper, sdk.Context) {
	keys := storetypes.NewKVStoreKeys(
		authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey,
		group.StoreKey,
		storagemoduletypes.StoreKey,
		crosschaintypes.StoreKey,
		oracletypes.StoreKey,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NoOpMetrics{})
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)

	stateStore.MountStoreWithDB(keys[paramstypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(keys[authtypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(keys[banktypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(keys[stakingtypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(keys[crosschaintypes.StoreKey], storetypes.StoreTypeIAVL, db)

	stateStore.MountStoreWithDB(tkeys[paramstypes.TStoreKey], storetypes.StoreTypeTransient, nil)

	require.NoError(t, stateStore.LoadLatestVersion())

	cdcConfig := app.MakeEncodingConfig()

	cdc := cdcConfig.Marshaler

	paramKeeper := initParamsKeeper(cdc, types.Amino, keys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"BridgeParams",
	)

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		keys[authtypes.StoreKey],
		GetSubspace(paramKeeper, authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)

	authzKeeper := authzkeeper.NewKeeper(
		keys[authz.ModuleName],
		cdc,
		baseapp.NewMsgServiceRouter(),
		accountKeeper,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		keys[banktypes.StoreKey],
		accountKeeper,
		GetSubspace(paramKeeper, banktypes.ModuleName),
		nil,
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		cdc,
		keys[stakingtypes.StoreKey],
		accountKeeper,
		authzKeeper,
		bankKeeper,
		GetSubspace(paramKeeper, stakingtypes.ModuleName),
	)

	crossChainKeeper := crosschainkeeper.NewKeeper(
		cdc,
		keys[crosschaintypes.StoreKey],
		GetSubspace(paramKeeper, crosschaintypes.ModuleName),
	)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		bankKeeper,
		stakingKeeper,
		crossChainKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil, log.NewNopLogger())

	stakingKeeper.SetParams(ctx, stakingtypes.DefaultParams())

	accountKeeper.SetParams(ctx, authtypes.DefaultParams())

	err := bankKeeper.MintCoins(ctx, authtypes.FeeCollectorName, sdk.Coins{sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(10000000000000000),
	}})
	if err != nil {
		panic("mint coins error")
	}
	err = bankKeeper.MintCoins(ctx, crosschaintypes.ModuleName, sdk.Coins{sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(10000000000000000),
	}})
	if err != nil {
		panic("mint coins error")
	}

	crossChainKeeper.SetSrcChainID(sdk.ChainID(1))
	crossChainKeeper.SetDestChainID(sdk.ChainID(2))

	crossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(2), types.TransferOutChannelID, sdk.ChannelAllow)
	crossChainKeeper.SetChannelSendPermission(ctx, sdk.ChainID(2), types.TransferInChannelID, sdk.ChannelAllow)

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return &BridgeKeeperSuite{
		Cdc:           cdc,
		BridgeKeeper:  k,
		Ctx:           sdk.Context{},
		BankKeeper:    &bankKeeper,
		AccountKeeper: accountKeeper,
	}, k, ctx
}

func initParamsKeeper(appCodec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(storagemoduletypes.ModuleName)
	paramsKeeper.Subspace(crosschaintypes.ModuleName)
	paramsKeeper.Subspace(oracletypes.ModuleName)
	paramsKeeper.Subspace(types.ModuleName)
	// this line is used by starport scaffolding # stargate/app/paramSubspace

	return paramsKeeper
}

func GetSubspace(keeper paramskeeper.Keeper, moduleName string) paramstypes.Subspace {
	subspace, _ := keeper.GetSubspace(moduleName)
	return subspace
}
