package keeper

import (
	"testing"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/metrics"
	storetypes "cosmossdk.io/store/types"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	tmdb "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	typesparams "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/payment/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
	spkeeper "github.com/bnb-chain/greenfield/x/sp/keeper"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

func PaymentKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKeys := storetypes.NewKVStoreKeys(
		banktypes.StoreKey,
		authtypes.StoreKey,
		paramstypes.StoreKey,
		sptypes.StoreKey,
		authz.ModuleName,
	)
	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NoOpMetrics{})
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(storeKeys[paramstypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[authtypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[banktypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(tkeys[paramstypes.TStoreKey], storetypes.StoreTypeTransient, nil)
	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramKeeper := paramskeeper.NewKeeper(cdc, types.Amino, storeKeys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	paramKeeper.Subspace(authtypes.ModuleName)
	paramKeeper.Subspace(banktypes.ModuleName)
	paramKeeper.Subspace(authz.ModuleName)
	paramKeeper.Subspace(sptypes.ModuleName)

	paramsSubspace := typesparams.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"PaymentParams",
	)
	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(storeKeys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		spMaccPerms,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		storeKeys[banktypes.StoreKey],
		accountKeeper,
		GetSubspace(paramKeeper, banktypes.ModuleName),
		nil,
	)
	authzKeeper := authzkeeper.NewKeeper(
		storeKeys[authz.ModuleName],
		cdc,
		baseapp.NewMsgServiceRouter(),
		accountKeeper,
	)
	spKeeper := spkeeper.NewKeeper(
		cdc,
		storeKeys[sptypes.ModuleName],
		storeKeys[sptypes.MemStoreKey],
		GetSubspace(paramKeeper, sptypes.ModuleName),
		accountKeeper,
		bankKeeper,
		authzKeeper,
	)
	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		bankKeeper,
		accountKeeper,
		spKeeper,
	)

	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	return k, ctx
}
