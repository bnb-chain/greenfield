package keeper_test

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	vgtypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (s *TestSuite) TestUpdateObjectContent_IBIConsistency() {
	owner := sample.RandAccAddress()
	minChargeSize := uint64(1024 * 1024) // 1MB

	gvgFamily := &vgtypes.GlobalVirtualGroupFamily{
		Id:                    1,
		PrimarySpId:           1,
		GlobalVirtualGroupIds: []uint32{1},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}

	gvg := &vgtypes.GlobalVirtualGroup{
		Id:                    1,
		PrimarySpId:           1,
		SecondarySpIds:        []uint32{2, 3, 4, 5, 6, 7},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}

	sp := &sptypes.StorageProvider{
		Id:              1,
		OperatorAddress: owner.String(),
		Status:          sptypes.STATUS_IN_SERVICE,
	}

	price := sptypes.GlobalSpStorePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}

	paymentParams := paymenttypes.VersionedParams{
		ReserveTime:      10000,
		ValidatorTaxRate: sdk.NewDecWithPrec(1, 2), // 1%
	}

	bucketName := "test-bucket"
	objectName := "test-object"
	originalPayloadSize := uint64(10 * 1024 * 1024) // 10MB

	bucketInfo := &types.BucketInfo{
		Owner:                      owner.String(),
		BucketName:                 bucketName,
		Id:                         sdk.NewUint(1),
		PaymentAddress:             owner.String(),
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
		ChargedReadQuota:           0,
		BucketStatus:               types.BUCKET_STATUS_CREATED,
	}

	objectInfo := &types.ObjectInfo{
		Id:                  sdk.NewUint(1),
		BucketName:          bucketName,
		ObjectName:          objectName,
		Owner:               owner.String(),
		PayloadSize:         originalPayloadSize,
		ObjectStatus:        types.OBJECT_STATUS_SEALED,
		LocalVirtualGroupId: 1,
		Visibility:          types.VISIBILITY_TYPE_PRIVATE,
		CreateAt:            100,
	}

	makeIBI := func() *types.InternalBucketInfo {
		return &types.InternalBucketInfo{
			TotalChargeSize: originalPayloadSize,
			LocalVirtualGroups: []*types.LocalVirtualGroup{
				{
					Id:                   1,
					GlobalVirtualGroupId: gvg.Id,
					StoredSize:           originalPayloadSize,
					TotalChargeSize:      originalPayloadSize,
				},
			},
		}
	}

	setupMocks := func() {
		s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gvgFamily.Id).
			Return(gvgFamily, true).AnyTimes()
		s.virtualGroupKeeper.EXPECT().GetGVG(gomock.Any(), gvg.Id).
			Return(gvg, true).AnyTimes()
		s.virtualGroupKeeper.EXPECT().GetGlobalVirtualGroupIfAvailable(gomock.Any(), gvg.Id, gomock.Any()).
			Return(gvg, nil).AnyTimes()
		s.virtualGroupKeeper.EXPECT().SetGVGAndEmitUpdateEvent(gomock.Any(), gomock.Any()).
			Return(nil).AnyTimes()
		s.spKeeper.EXPECT().MustGetStorageProvider(gomock.Any(), sp.Id).
			Return(sp).AnyTimes()
		s.spKeeper.EXPECT().GetGlobalSpStorePriceByTime(gomock.Any(), gomock.Any()).
			Return(price, nil).AnyTimes()
		s.paymentKeeper.EXPECT().GetVersionedParamsWithTs(gomock.Any(), gomock.Any()).
			Return(paymentParams, nil).AnyTimes()
		s.paymentKeeper.EXPECT().ApplyUserFlowsList(gomock.Any(), gomock.Any()).
			Return(nil).AnyTimes()
		s.paymentKeeper.EXPECT().UpdateStreamRecordByAddr(gomock.Any(), gomock.Any()).
			Return(&paymenttypes.StreamRecord{
				StaticBalance: sdk.NewInt(1000000000),
			}, nil).AnyTimes()
		s.paymentKeeper.EXPECT().MergeOutFlows(gomock.Any()).DoAndReturn(
			func(flows []paymenttypes.OutFlow) []paymenttypes.OutFlow {
				return flows
			}).AnyTimes()
	}

	s.Run("before Taiga upgrade - IBI is stale", func() {
		s.SetupTest()
		// store versioned params at t=1 so GetObjectChargeSize can find them for object with CreateAt=1
		s.ctx = s.ctx.WithBlockTime(time.Unix(1, 0))
		err := s.storageKeeper.SetVersionedParamsWithTs(s.ctx, types.VersionedParams{MinChargeSize: minChargeSize})
		s.Require().NoError(err)
		s.ctx = s.ctx.WithBlockTime(time.Unix(100, 0))

		setupMocks()

		s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)
		s.storageKeeper.StoreObjectInfo(s.ctx, objectInfo)
		s.storageKeeper.SetInternalBucketInfo(s.ctx, bucketInfo.Id, makeIBI())

		err = s.storageKeeper.UpdateObjectContent(s.ctx, owner, bucketName, objectName, 0, types.UpdateObjectOptions{
			Checksums: make([][]byte, 1+len(gvg.SecondarySpIds)),
		})
		s.Require().NoError(err)

		ibi := s.storageKeeper.MustGetInternalBucketInfo(s.ctx, bucketInfo.Id)

		// Before Taiga: TotalChargeSize is stale — the original 10MB was NOT reduced by UnChargeObjectStoreFee
		// (it modified a separate copy), then SealEmptyObjectOnVirtualGroup added min_charge_size on top.
		// Result: 10MB + 1MB = 11MB instead of correct 1MB.
		s.Require().Greater(ibi.TotalChargeSize, minChargeSize,
			"before Taiga, TotalChargeSize should be inflated due to stale IBI bug")
	})

	s.Run("after Taiga upgrade - IBI is consistent", func() {
		s.SetupTest()
		upgradeChecker := func(ctx sdk.Context, name string) bool {
			return name == upgradetypes.Serengeti || name == upgradetypes.Taiga
		}
		header := s.ctx.BlockHeader()
		header.Time = time.Unix(1, 0)
		s.ctx = sdk.NewContext(s.ctx.MultiStore(), header, false, upgradeChecker, s.ctx.Logger())

		err := s.storageKeeper.SetParams(s.ctx, types.DefaultParams())
		s.Require().NoError(err)
		err = s.storageKeeper.SetVersionedParamsWithTs(s.ctx, types.VersionedParams{MinChargeSize: minChargeSize})
		s.Require().NoError(err)
		// advance block time past the params timestamp
		header.Time = time.Unix(100, 0)
		s.ctx = sdk.NewContext(s.ctx.MultiStore(), header, false, upgradeChecker, s.ctx.Logger())

		setupMocks()

		s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)
		s.storageKeeper.StoreObjectInfo(s.ctx, objectInfo)
		s.storageKeeper.SetInternalBucketInfo(s.ctx, bucketInfo.Id, makeIBI())

		err = s.storageKeeper.UpdateObjectContent(s.ctx, owner, bucketName, objectName, 0, types.UpdateObjectOptions{
			Checksums: make([][]byte, 1+len(gvg.SecondarySpIds)),
		})
		s.Require().NoError(err)

		ibi := s.storageKeeper.MustGetInternalBucketInfo(s.ctx, bucketInfo.Id)

		// After Taiga: TotalChargeSize should be exactly min_charge_size (1MB).
		// UnChargeObjectStoreFee correctly reduced 10MB on the same IBI copy,
		// then SealEmptyObjectOnVirtualGroup added min_charge_size for the empty object.
		s.Require().Equal(minChargeSize, ibi.TotalChargeSize,
			"after Taiga, TotalChargeSize should equal min_charge_size")

		hasNonZeroLVG := false
		for _, lvg := range ibi.LocalVirtualGroups {
			if lvg.TotalChargeSize > 0 {
				s.Require().Equal(minChargeSize, lvg.TotalChargeSize)
				hasNonZeroLVG = true
			}
		}
		s.Require().True(hasNonZeroLVG, "should have exactly one LVG with min_charge_size")
	})
}
