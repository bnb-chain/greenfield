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
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/group"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/stretchr/testify/require"

	paymentmoduletypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/permission/keeper"
	"github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

var (
	permissionMaccPerms = map[string][]string{
		authtypes.FeeCollectorName:     {authtypes.Minter, authtypes.Staking},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		sptypes.ModuleName:             {authtypes.Staking},
		types.ModuleName:               {authtypes.Staking},
	}
)

func PermissionKeeper(t testing.TB) (*keeper.Keeper, sdk.Context) {
	storeKeys := storetypes.NewKVStoreKeys(authtypes.StoreKey, authz.ModuleName, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey, govtypes.StoreKey,
		paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey, evidencetypes.StoreKey,
		group.StoreKey,
		crosschaintypes.StoreKey,
		paymentmoduletypes.StoreKey,
		oracletypes.StoreKey, types.StoreKey)

	memStoreKey := storetypes.NewMemoryStoreKey(types.MemStoreKey)

	tkeys := storetypes.NewTransientStoreKeys(paramstypes.TStoreKey)

	storeKey := storetypes.NewKVStoreKey(types.StoreKey)

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
	paramKeeper.Subspace(paymentmoduletypes.ModuleName)

	paramsSubspace := paramstypes.NewSubspace(cdc,
		types.Amino,
		storeKey,
		memStoreKey,
		"PermissionParams",
	)

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc,
		runtime.NewKVStoreService(storeKeys[authtypes.StoreKey]),
		authtypes.ProtoBaseAccount,
		maccPerms,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	k := keeper.NewKeeper(
		cdc,
		storeKey,
		memStoreKey,
		paramsSubspace,
		accountKeeper,
	)
	ctx := sdk.NewContext(stateStore, tmproto.Header{}, false, nil, log.NewNopLogger())

	// Initialize params
	k.SetParams(ctx, types.DefaultParams())

	accountKeeper.SetParams(ctx, authtypes.DefaultParams())

	return k, ctx

}
