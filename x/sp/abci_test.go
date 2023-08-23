package sp_test

import (
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/testutil/sample"
	spmodule "github.com/bnb-chain/greenfield/x/sp"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

type TestSuite struct {
	suite.Suite

	cdc      codec.Codec
	spKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	accountKeeper *types.MockAccountKeeper
	authzKeeper   *types.MockAuthzKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))
	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	accountKeeper := types.NewMockAccountKeeper(ctrl)
	authzKeeper := types.NewMockAuthzKeeper(ctrl)

	s.spKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		accountKeeper,
		bankKeeper,
		authzKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec

	s.bankKeeper = bankKeeper
	s.accountKeeper = accountKeeper
	s.authzKeeper = authzKeeper

	err := s.spKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.spKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.spKeeper)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestEndBlocker_NoGlobalPrice() {
	s.ctx = s.ctx.WithBlockTime(time.Now())
	sp := &types.StorageProvider{
		Id:              1,
		Status:          types.STATUS_IN_SERVICE,
		OperatorAddress: sample.RandAccAddress().String(),
	}
	s.spKeeper.SetStorageProvider(s.ctx, sp)
	spPrice := types.SpStoragePrice{
		SpId:          1,
		UpdateTimeSec: 1024,
		ReadPrice:     sdk.NewDecWithPrec(100, 0),
		FreeReadQuota: 0,
		StorePrice:    sdk.NewDecWithPrec(200, 0),
	}
	s.spKeeper.SetSpStoragePrice(s.ctx, spPrice)

	spmodule.EndBlocker(s.ctx, *s.spKeeper)

	// new global price
	globalPrice, err := s.spKeeper.GetGlobalSpStorePriceByTime(s.ctx, s.ctx.BlockTime().Unix()+1)
	s.Require().NoError(err)
	s.Require().Equal(globalPrice.PrimaryStorePrice, spPrice.StorePrice)
	s.Require().Equal(globalPrice.ReadPrice, spPrice.ReadPrice)
}

func (s *TestSuite) TestEndBlocker_WithUpdateInterval() {
	preTime := int64(1691648908)
	s.ctx = s.ctx.WithBlockTime(time.Unix(preTime, 0))
	globalPrice := types.GlobalSpStorePrice{
		UpdateTimeSec:       0,
		ReadPrice:           sdk.NewDecWithPrec(1, 0),
		PrimaryStorePrice:   sdk.NewDecWithPrec(1, 0),
		SecondaryStorePrice: sdk.NewDecWithPrec(12, 2),
	}
	s.spKeeper.SetGlobalSpStorePrice(s.ctx, globalPrice)

	params := s.spKeeper.GetParams(s.ctx)
	params.UpdateGlobalPriceInterval = 100
	_ = s.spKeeper.SetParams(s.ctx, params)

	newTime := preTime + 3
	s.ctx = s.ctx.WithBlockTime(time.Unix(newTime, 0))
	sp := &types.StorageProvider{
		Id:              1,
		Status:          types.STATUS_IN_SERVICE,
		OperatorAddress: sample.RandAccAddress().String(),
	}
	s.spKeeper.SetStorageProvider(s.ctx, sp)
	spPrice := types.SpStoragePrice{
		SpId:          1,
		UpdateTimeSec: 1024,
		ReadPrice:     sdk.NewDecWithPrec(100, 0),
		FreeReadQuota: 0,
		StorePrice:    sdk.NewDecWithPrec(200, 0),
	}
	s.spKeeper.SetSpStoragePrice(s.ctx, spPrice)

	spmodule.EndBlocker(s.ctx, *s.spKeeper)
	// global price will not be updated for not reaching the interval
	globalPriceAfter, err := s.spKeeper.GetGlobalSpStorePriceByTime(s.ctx, s.ctx.BlockTime().Unix())
	s.Require().NoError(err)
	s.Require().Equal(globalPrice.PrimaryStorePrice, globalPriceAfter.PrimaryStorePrice)
	s.Require().Equal(globalPrice.ReadPrice, globalPriceAfter.ReadPrice)

	newTime = preTime + 11
	s.ctx = s.ctx.WithBlockTime(time.Unix(newTime, 0))
	spmodule.EndBlocker(s.ctx, *s.spKeeper)
	// new global price
	globalPriceAfter, err = s.spKeeper.GetGlobalSpStorePriceByTime(s.ctx, s.ctx.BlockTime().Unix()+1)
	s.Require().NoError(err)
	s.Require().Equal(globalPriceAfter.PrimaryStorePrice, spPrice.StorePrice)
	s.Require().Equal(globalPriceAfter.ReadPrice, spPrice.ReadPrice)
}

func (s *TestSuite) TestEndBlocker_WithoutUpdateInterval() {
	preTime := int64(1691648908)
	s.ctx = s.ctx.WithBlockTime(time.Unix(preTime, 0))
	globalPrice := types.GlobalSpStorePrice{
		UpdateTimeSec:       0,
		ReadPrice:           sdk.NewDecWithPrec(1, 0),
		PrimaryStorePrice:   sdk.NewDecWithPrec(1, 0),
		SecondaryStorePrice: sdk.NewDecWithPrec(12, 2),
	}
	s.spKeeper.SetGlobalSpStorePrice(s.ctx, globalPrice)

	params := s.spKeeper.GetParams(s.ctx)
	params.UpdateGlobalPriceInterval = 0
	_ = s.spKeeper.SetParams(s.ctx, params)

	newTime := preTime + 3
	s.ctx = s.ctx.WithBlockTime(time.Unix(newTime, 0))
	sp := &types.StorageProvider{
		Id:              1,
		Status:          types.STATUS_IN_SERVICE,
		OperatorAddress: sample.RandAccAddress().String(),
	}
	s.spKeeper.SetStorageProvider(s.ctx, sp)
	spPrice := types.SpStoragePrice{
		SpId:          1,
		UpdateTimeSec: 1024,
		ReadPrice:     sdk.NewDecWithPrec(100, 0),
		FreeReadQuota: 0,
		StorePrice:    sdk.NewDecWithPrec(200, 0),
	}
	s.spKeeper.SetSpStoragePrice(s.ctx, spPrice)

	// in the same month
	s.ctx = s.ctx.WithBlockTime(time.Unix(newTime+2, 0))
	spmodule.EndBlocker(s.ctx, *s.spKeeper)
	// global price will not be updated
	globalPriceAfter, err := s.spKeeper.GetGlobalSpStorePriceByTime(s.ctx, s.ctx.BlockTime().Unix())
	s.Require().NoError(err)
	s.Require().Equal(globalPrice.PrimaryStorePrice, globalPriceAfter.PrimaryStorePrice)
	s.Require().Equal(globalPrice.ReadPrice, globalPriceAfter.ReadPrice)

	// a new month
	t := time.Unix(newTime+10, 0).UTC()
	year, month, _ := t.Date()
	location := t.Location()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, location)
	s.ctx = s.ctx.WithBlockTime(nextMonth)
	spmodule.EndBlocker(s.ctx, *s.spKeeper)
	// new global price
	globalPriceAfter, err = s.spKeeper.GetGlobalSpStorePriceByTime(s.ctx, s.ctx.BlockTime().Unix()+1)
	s.Require().NoError(err)
	s.Require().Equal(globalPriceAfter.PrimaryStorePrice, spPrice.StorePrice)
	s.Require().Equal(globalPriceAfter.ReadPrice, spPrice.ReadPrice)
}
