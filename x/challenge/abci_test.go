package challenge_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/x/challenge"
	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type TestSuite struct {
	suite.Suite

	cdc             codec.Codec
	challengeKeeper *keeper.Keeper

	bankKeeper    *types.MockBankKeeper
	storageKeeper *types.MockStorageKeeper
	spKeeper      *types.MockSpKeeper
	stakingKeeper *types.MockStakingKeeper
	paymentKeeper *types.MockPaymentKeeper

	ctx         sdk.Context
	queryClient types.QueryClient
	msgServer   types.MsgServer
}

func (s *TestSuite) SetupTest() {
	encCfg := moduletestutil.MakeTestEncodingConfig(challenge.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_test"))

	// set mock randao mix
	randaoMix := sdk.Keccak256([]byte{1})
	randaoMix = append(randaoMix, sdk.Keccak256([]byte{2})...)
	header := testCtx.Ctx.BlockHeader()
	header.RandaoMix = randaoMix
	testCtx = testutil.TestContext{
		Ctx: sdk.NewContext(testCtx.CMS, header, false, nil, testCtx.Ctx.Logger()),
		DB:  testCtx.DB,
		CMS: testCtx.CMS,
	}

	s.ctx = testCtx.Ctx

	ctrl := gomock.NewController(s.T())

	bankKeeper := types.NewMockBankKeeper(ctrl)
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	spKeeper := types.NewMockSpKeeper(ctrl)
	stakingKeeper := types.NewMockStakingKeeper(ctrl)
	paymentKeeper := types.NewMockPaymentKeeper(ctrl)

	s.challengeKeeper = keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		bankKeeper,
		storageKeeper,
		spKeeper,
		stakingKeeper,
		paymentKeeper,
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	s.cdc = encCfg.Codec
	s.bankKeeper = bankKeeper
	s.storageKeeper = storageKeeper
	s.spKeeper = spKeeper
	s.stakingKeeper = stakingKeeper
	s.paymentKeeper = paymentKeeper

	err := s.challengeKeeper.SetParams(s.ctx, types.DefaultParams())
	s.Require().NoError(err)

	queryHelper := baseapp.NewQueryServerTestHelper(testCtx.Ctx, encCfg.InterfaceRegistry)
	types.RegisterQueryServer(queryHelper, s.challengeKeeper)

	s.queryClient = types.NewQueryClient(queryHelper)
	s.msgServer = keeper.NewMsgServerImpl(*s.challengeKeeper)
}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}

func (s *TestSuite) TestBeginBlocker_RemoveExpiredChallenge() {
	s.challengeKeeper.SaveChallenge(s.ctx, types.Challenge{
		Id:            100,
		ExpiredHeight: 100,
	})
	s.challengeKeeper.SaveChallenge(s.ctx, types.Challenge{
		Id:            200,
		ExpiredHeight: 300,
	})

	s.ctx = s.ctx.WithBlockHeight(101)
	challenge.BeginBlocker(s.ctx, *s.challengeKeeper)
	s.Require().False(s.challengeKeeper.ExistsChallenge(s.ctx, 100))
	s.Require().True(s.challengeKeeper.ExistsChallenge(s.ctx, 200))
}

func (s *TestSuite) TestBeginBlocker_RemoveSlash() {
	s.challengeKeeper.SaveSlash(s.ctx, types.Slash{
		SpId:     100,
		ObjectId: sdk.NewUint(100),
		Height:   100,
	})
	s.challengeKeeper.SaveSlash(s.ctx, types.Slash{
		SpId:     200,
		ObjectId: sdk.NewUint(200),
		Height:   200,
	})

	params := s.challengeKeeper.GetParams(s.ctx)
	params.SlashCoolingOffPeriod = 10
	_ = s.challengeKeeper.SetParams(s.ctx, params)

	s.ctx = s.ctx.WithBlockHeight(101)
	challenge.BeginBlocker(s.ctx, *s.challengeKeeper)
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, 100, sdk.NewUint(100)))
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, 200, sdk.NewUint(200)))

	s.ctx = s.ctx.WithBlockHeight(111)
	challenge.BeginBlocker(s.ctx, *s.challengeKeeper)
	s.Require().False(s.challengeKeeper.ExistsSlash(s.ctx, 100, sdk.NewUint(100)))
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, 200, sdk.NewUint(200)))

	s.ctx = s.ctx.WithBlockHeight(211)
	challenge.BeginBlocker(s.ctx, *s.challengeKeeper)
	s.Require().False(s.challengeKeeper.ExistsSlash(s.ctx, 100, sdk.NewUint(100)))
	s.Require().False(s.challengeKeeper.ExistsSlash(s.ctx, 200, sdk.NewUint(200)))
}

func (s *TestSuite) TestBeginBlocker_RemoveSpSlashAmount() {
	s.challengeKeeper.SetSpSlashAmount(s.ctx, 100, sdk.NewInt(100))
	s.challengeKeeper.SetSpSlashAmount(s.ctx, 200, sdk.NewInt(200))

	params := s.challengeKeeper.GetParams(s.ctx)
	params.SpSlashCountingWindow = 10
	_ = s.challengeKeeper.SetParams(s.ctx, params)

	s.ctx = s.ctx.WithBlockHeight(101)
	challenge.BeginBlocker(s.ctx, *s.challengeKeeper)
	s.Require().True(s.challengeKeeper.GetSpSlashAmount(s.ctx, 100).Int64() == 100)
	s.Require().True(s.challengeKeeper.GetSpSlashAmount(s.ctx, 200).Int64() == 200)

	s.ctx = s.ctx.WithBlockHeight(100)
	challenge.BeginBlocker(s.ctx, *s.challengeKeeper)
	s.Require().False(s.challengeKeeper.GetSpSlashAmount(s.ctx, 100).Int64() == 100)
	s.Require().False(s.challengeKeeper.GetSpSlashAmount(s.ctx, 200).Int64() == 200)
}

func (s *TestSuite) TestEndBlocker_NoRandomChallenge() {
	preChallengeId := s.challengeKeeper.GetChallengeId(s.ctx)

	params := s.challengeKeeper.GetParams(s.ctx)
	params.ChallengeCountPerBlock = 0
	_ = s.challengeKeeper.SetParams(s.ctx, params)

	challenge.EndBlocker(s.ctx, *s.challengeKeeper)
	afterChallengeId := s.challengeKeeper.GetChallengeId(s.ctx)
	s.Require().True(preChallengeId == afterChallengeId)
}

func (s *TestSuite) TestEndBlocker_ObjectNotExists() {
	s.storageKeeper.EXPECT().GetObjectInfoCount(gomock.Any()).Return(sdk.NewUint(0))

	preChallengeId := s.challengeKeeper.GetChallengeId(s.ctx)
	challenge.EndBlocker(s.ctx, *s.challengeKeeper)
	afterChallengeId := s.challengeKeeper.GetChallengeId(s.ctx)
	s.Require().True(preChallengeId == afterChallengeId)
}

func (s *TestSuite) TestEndBlocker_SuccessRandomChallenge() {
	s.storageKeeper.EXPECT().GetObjectInfoCount(gomock.Any()).Return(sdk.NewUint(100))
	s.storageKeeper.EXPECT().MaxSegmentSize(gomock.Any(), gomock.Any()).Return(uint64(10000), nil).AnyTimes()

	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(64),
		BucketName:   "bucketname",
		ObjectName:   "objectname",
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(existObject.Id)).
		Return(existObject, true).AnyTimes()

	existBucket := &storagetypes.BucketInfo{
		BucketName: existObject.BucketName,
		Id:         math.NewUint(10),
	}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(existBucket.BucketName)).
		Return(existBucket, true).AnyTimes()

	gvg := &virtualgrouptypes.GlobalVirtualGroup{PrimarySpId: 100, SecondarySpIds: []uint32{
		1,
	}}
	s.storageKeeper.EXPECT().GetObjectGVG(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(gvg, true).AnyTimes()

	sp := &sptypes.StorageProvider{Id: 1, Status: sptypes.STATUS_IN_SERVICE}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Any()).
		Return(sp, true).AnyTimes()

	preChallengeId := s.challengeKeeper.GetChallengeId(s.ctx)
	challenge.EndBlocker(s.ctx, *s.challengeKeeper)
	afterChallengeId := s.challengeKeeper.GetChallengeId(s.ctx)
	s.Require().True(preChallengeId == afterChallengeId-1)
}
