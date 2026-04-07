package keeper_test

import (
	"errors"
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/keeper"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type TestSuite struct {
	suite.Suite

	cdc                codec.Codec
	virtualgroupKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	accountKeeper *types.MockAccountKeeper
	spKeeper      *types.MockSpKeeper
	paymentKeeper *types.MockPaymentKeeper

	ctx sdk.Context
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(challenge.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	paymentKeeper := types.NewMockPaymentKeeper(ctrl)

	s.ctx = testCtx.Ctx
	s.virtualgroupKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		spKeeper,
		accountKeeper,
		bankKeeper,
		paymentKeeper,
	)

	s.cdc = encCfg.Codec
	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.spKeeper = spKeeper
	s.paymentKeeper = paymentKeeper

	err := s.virtualgroupKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestSettleAndDistributeGVGFamily() {
	sp := &sptypes.StorageProvider{Id: 1, FundingAddress: sample.RandAccAddress().String()}
	family := &types.GlobalVirtualGroupFamily{Id: 1, VirtualPaymentAddress: sample.RandAccAddress().String()}

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), nil)
	err := s.virtualgroupKeeper.SettleAndDistributeGVGFamily(s.ctx, sp, family)
	require.NoError(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), errors.New("error"))
	err = s.virtualgroupKeeper.SettleAndDistributeGVGFamily(s.ctx, sp, family)
	require.Error(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1024), nil)
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	err = s.virtualgroupKeeper.SettleAndDistributeGVGFamily(s.ctx, sp, family)
	require.NoError(s.T(), err)
}

func (s *TestSuite) TestSettleAndDistributeGVG() {
	gvg := &types.GlobalVirtualGroup{Id: 1,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
		SecondarySpIds:        []uint32{3, 6, 9}}

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), nil)
	err := s.virtualgroupKeeper.SettleAndDistributeGVG(s.ctx, gvg)
	require.NoError(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), errors.New("error"))
	err = s.virtualgroupKeeper.SettleAndDistributeGVG(s.ctx, gvg)
	require.Error(s.T(), err)

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1024), nil).AnyTimes()
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	sp := &sptypes.StorageProvider{Id: 1, FundingAddress: sample.RandAccAddress().String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Any()).
		Return(sp, true).AnyTimes()
	err = s.virtualgroupKeeper.SettleAndDistributeGVG(s.ctx, gvg)
	require.NoError(s.T(), err)
}

// setupGVGForDelete creates a GVG, its family, and the required SP statistics in KVStore.
func (s *TestSuite) setupGVGForDelete(ctx sdk.Context) (
	*sptypes.StorageProvider, *sptypes.StorageProvider, *types.GlobalVirtualGroup, *types.GlobalVirtualGroupFamily,
) {
	primarySp := &sptypes.StorageProvider{
		Id:             1,
		FundingAddress: sample.RandAccAddress().String(),
	}
	secondarySp := &sptypes.StorageProvider{
		Id:             2,
		FundingAddress: sample.RandAccAddress().String(),
	}
	gvg := &types.GlobalVirtualGroup{
		Id:                    1,
		FamilyId:              1,
		PrimarySpId:           1,
		SecondarySpIds:        []uint32{2},
		StoredSize:            0,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
		TotalDeposit:          math.NewInt(1000),
	}
	family := &types.GlobalVirtualGroupFamily{
		Id:                    1,
		PrimarySpId:           1,
		GlobalVirtualGroupIds: []uint32{1},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}

	s.virtualgroupKeeper.SetGVG(ctx, gvg)
	s.virtualgroupKeeper.SetGVGFamily(ctx, family)
	s.virtualgroupKeeper.SetGVGStatisticsWithSP(ctx, &types.GVGStatisticsWithinSP{
		StorageProviderId: 1, PrimaryCount: 1,
	})
	s.virtualgroupKeeper.SetGVGStatisticsWithSP(ctx, &types.GVGStatisticsWithinSP{
		StorageProviderId: 2, SecondaryCount: 1,
	})
	return primarySp, secondarySp, gvg, family
}

// TestDeleteGVG_BeforePrairie verifies that before the Prairie hardfork,
// DeleteGVG succeeds without settling the GVG virtual payment account balance.
// Any undistributed StaticBalance is silently left behind (the bug).
func (s *TestSuite) TestDeleteGVG_BeforePrairie_SkipsSettlement() {
	// s.ctx has nil upgradeChecker → IsUpgraded returns false for all
	primarySp, _, _, _ := s.setupGVGForDelete(s.ctx)

	// IsEmptyNetFlow called twice:
	// 1) GVG payment address check
	// 2) family payment address check (for empty-family cleanup, since Manchurian is also not upgraded)
	s.paymentKeeper.EXPECT().IsEmptyNetFlow(gomock.Any(), gomock.Any()).
		Return(true).Times(2)
	// Deposit refund
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	// NOTE: No QueryDynamicBalance expectation set — if the settlement code path
	// is incorrectly reached, gomock will fail with "unexpected call".

	err := s.virtualgroupKeeper.DeleteGVG(s.ctx, primarySp, 1)
	require.NoError(s.T(), err)

	// GVG should be deleted
	_, found := s.virtualgroupKeeper.GetGVG(s.ctx, 1)
	require.False(s.T(), found)
}

// TestDeleteGVG_AfterPrairie verifies that after the Prairie hardfork,
// DeleteGVG settles the GVG virtual payment account balance before deletion,
// distributing remaining funds to secondary SPs.
func (s *TestSuite) TestDeleteGVG_AfterPrairie_SettlesBalance() {
	// Create context with Prairie (and Manchurian) upgrade enabled
	ctx := sdk.NewContext(
		s.ctx.MultiStore(),
		s.ctx.BlockHeader(),
		false,
		func(_ sdk.Context, name string) bool {
			return name == upgradetypes.Prairie || name == upgradetypes.Manchurian
		},
		s.ctx.Logger(),
	)

	primarySp, secondarySp, _, _ := s.setupGVGForDelete(ctx)

	// 1) IsEmptyNetFlow for GVG — allows deletion to proceed
	s.paymentKeeper.EXPECT().IsEmptyNetFlow(gomock.Any(), gomock.Any()).
		Return(true)

	// 2) Prairie settlement path:
	//    DeleteGVG: QueryDynamicBalance → 1024 (positive, triggers settlement)
	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1024), nil)
	//    SettleAndDistributeGVG: QueryDynamicBalance → 1024
	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1024), nil)
	//    SettleAndDistributeGVG: GetStorageProvider for secondary SP
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), uint32(2)).
		Return(secondarySp, true)
	//    SettleAndDistributeGVG: Withdraw to secondary SP
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), math.NewInt(1024)).
		Return(nil)
	//    DeleteGVG: re-check QueryDynamicBalance → 0 (balance cleared)
	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.ZeroInt(), nil)

	// 3) Deposit refund
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil)
	// Family cleanup condition: len==0 is true, so Go evaluates IsEmptyNetFlow
	// before reaching !IsUpgraded(Manchurian) which is false
	s.paymentKeeper.EXPECT().IsEmptyNetFlow(gomock.Any(), gomock.Any()).
		Return(true)

	err := s.virtualgroupKeeper.DeleteGVG(ctx, primarySp, 1)
	require.NoError(s.T(), err)

	// GVG should be deleted
	_, found := s.virtualgroupKeeper.GetGVG(ctx, 1)
	require.False(s.T(), found)
}
