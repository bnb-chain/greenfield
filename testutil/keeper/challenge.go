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
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschainkeeper "github.com/cosmos/cosmos-sdk/x/crosschain/keeper"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	paymentkeeper "github.com/bnb-chain/greenfield/x/payment/keeper"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permissionmodulekeeper "github.com/bnb-chain/greenfield/x/permission/keeper"
	permissionmoduletypes "github.com/bnb-chain/greenfield/x/permission/types"
	spkeeper "github.com/bnb-chain/greenfield/x/sp/keeper"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagekeeper "github.com/bnb-chain/greenfield/x/storage/keeper"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func ChallengeKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKeys := storetypes.NewKVStoreKeys(paramstypes.StoreKey, authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey,
		stakingtypes.StoreKey, storagetypes.StoreKey, paymenttypes.StoreKey, crosschaintypes.StoreKey)

	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey, types.TStoreKey)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	db := tmdb.NewMemDB()
	stateStore := store.NewCommitMultiStore(db, log.NewNopLogger(), metrics.NoOpMetrics{})
	stateStore.MountStoreWithDB(storeKey, storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(memStoreKey, storetypes.StoreTypeMemory, nil)
	stateStore.MountStoreWithDB(storeKeys[paramstypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[authtypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[banktypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[authz.ModuleName], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[stakingtypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[storagetypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[paymenttypes.StoreKey], storetypes.StoreTypeIAVL, db)
	stateStore.MountStoreWithDB(storeKeys[crosschaintypes.StoreKey], storetypes.StoreTypeIAVL, db)

	stateStore.MountStoreWithDB(tkeys[paramstypes.TStoreKey], storetypes.StoreTypeTransient, nil)

	require.NoError(t, stateStore.LoadLatestVersion())

	registry := codectypes.NewInterfaceRegistry()
	cdc := codec.NewProtoCodec(registry)

	paramKeeper := paramskeeper.NewKeeper(cdc, types.Amino, storeKeys[paramstypes.StoreKey], tkeys[paramstypes.TStoreKey])

	paramKeeper.Subspace(paramstypes.ModuleName)
	paramKeeper.Subspace(authtypes.ModuleName)
	paramKeeper.Subspace(banktypes.ModuleName)
	paramKeeper.Subspace(authz.ModuleName)
	paramKeeper.Subspace(storagetypes.ModuleName)
	paramKeeper.Subspace(sptypes.ModuleName)
	paramKeeper.Subspace(stakingtypes.ModuleName)
	paramKeeper.Subspace(paymenttypes.ModuleName)
	paramKeeper.Subspace(permissionmoduletypes.ModuleName)
	paramKeeper.Subspace(crosschaintypes.ModuleName)

	paramsSubspace := paramstypes.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"ChallengeParams",
	)

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		storeKeys[authtypes.StoreKey],
		GetSubspace(paramKeeper, authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
	)

	authzKeeper := authzkeeper.NewKeeper(
		storeKeys[authz.ModuleName],
		cdc,
		baseapp.NewMsgServiceRouter(),
		accountKeeper,
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc,
		storeKeys[banktypes.StoreKey],
		accountKeeper,
		GetSubspace(paramKeeper, banktypes.ModuleName),
		nil,
	)

	spKeeper := spkeeper.NewKeeper(
		cdc,
		storeKeys[sptypes.StoreKey],
		memStoreKey,
		GetSubspace(paramKeeper, sptypes.ModuleName),
		accountKeeper,
		bankKeeper,
		authzKeeper,
	)

	paymentKeeper := paymentkeeper.NewKeeper(
		cdc,
		storeKeys[paymenttypes.StoreKey],
		memStoreKey,
		GetSubspace(paramKeeper, paymenttypes.ModuleName),
		bankKeeper,
		accountKeeper,
		spKeeper,
	)
	permissionKeeper := permissionmodulekeeper.NewKeeper(
		cdc,
		storeKeys[permissionmoduletypes.ModuleName],
		storeKeys[permissionmoduletypes.MemStoreKey],
		GetSubspace(paramKeeper, permissionmoduletypes.ModuleName),
		accountKeeper,
	)
	crossChainKeeper := crosschainkeeper.NewKeeper(
		cdc,
		storeKeys[crosschaintypes.StoreKey],
		GetSubspace(paramKeeper, crosschaintypes.ModuleName),
	)

	storageKeeper := storagekeeper.NewKeeper(
		cdc,
		storeKeys[storagetypes.StoreKey],
		memStoreKey,
		GetSubspace(paramKeeper, storagetypes.ModuleName),
		accountKeeper,
		spKeeper,
		paymentKeeper,
		permissionKeeper,
		crossChainKeeper,
	)

	stakingKeeper := stakingkeeper.NewKeeper(
		cdc,
		storeKeys[stakingtypes.StoreKey],
		accountKeeper,
		authzKeeper,
		bankKeeper,
		GetSubspace(paramKeeper, stakingtypes.ModuleName),
	)

	k := keeper.NewKeeper(cdc,
		storeKey,
		memStoreKey,
		tkeys[types.TStoreKey],
		paramsSubspace,
		bankKeeper,
		storageKeeper,
		spKeeper,
		stakingKeeper,
		paymentKeeper,
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil, log.NewNopLogger())

	accountKeeper.SetParams(ctx, authtypes.DefaultParams())
	spKeeper.SetParams(ctx, sptypes.DefaultParams())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	err := bankKeeper.MintCoins(ctx, authtypes.FeeCollectorName, sdk.Coins{sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(1000000000),
	}})
	if err != nil {
		panic("mint coins error")
	}

	return k, ctx
}
