package sp_test

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/sp"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func TestGenesis(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	ctrl := gomock.NewController(t)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	bankKeeper := types.NewMockBankKeeper(ctrl)

	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		&types.MockAuthzKeeper{},
		"",
	)

	accountKeeperAcc := authtypes.NewEmptyModuleAccount(types.ModuleName)
	accountKeeper.EXPECT().GetModuleAccount(gomock.Any(), types.ModuleName).Return(accountKeeperAcc)
	bankKeeper.EXPECT().GetAllBalances(gomock.Any(), gomock.Any()).Return(sdk.NewCoins(sdk.NewCoin("BNB", sdkmath.NewInt(100000000000000))))

	operatorAddr, _, err := testutil.GenerateCoinKey(hd.Secp256k1, encCfg.Codec)
	require.NoError(t, err)

	genesisState := types.GenesisState{
		Params: types.DefaultParams(),
		StorageProviders: []types.StorageProvider{
			{
				OperatorAddress: operatorAddr.String(),
				TotalDeposit:    sdkmath.NewInt(100000000000000),
				Status:          types.STATUS_IN_SERVICE,
			},
		},
	}

	ctx := testCtx.Ctx

	sp.InitGenesis(ctx, *k, genesisState)
	got := sp.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)
}
