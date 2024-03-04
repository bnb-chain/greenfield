package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (s *TestSuite) TestSetBucketFlowRateLimit() {
	operatorAddress := sample.RandAccAddress()
	bucketOwner := sample.RandAccAddress()
	paymentAccount := sample.RandAccAddress()
	bucketName := string(sample.RandStr(10))

	// case 1: operator is not owner of payment account
	s.paymentKeeper.EXPECT().IsPaymentAccountOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(false)

	err := s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, sdkmath.NewInt(1))
	s.Require().ErrorContains(err, "not payment account owner")

	// case 2: bucket is not found
	s.paymentKeeper.EXPECT().IsPaymentAccountOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, sdkmath.NewInt(1))
	s.Require().ErrorContains(err, "No such bucket")

	bucketInfo := &types.BucketInfo{
		Owner:            bucketOwner.String(),
		BucketName:       bucketName,
		Id:               sdk.NewUint(1),
		PaymentAddress:   paymentAccount.String(),
		ChargedReadQuota: 0,
		BucketStatus:     types.BUCKET_STATUS_CREATED,
	}
	s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)

	// case 3: different bucket owner
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, sample.RandAccAddress(), paymentAccount, bucketName, sdkmath.NewInt(1))
	s.Require().ErrorContains(err, "invalid bucket owner")

	// case 4: bucket does not use the payment account
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, sample.RandAccAddress(), bucketName, sdkmath.NewInt(1))
	s.Require().NoError(err)
}

func (s *TestSuite) TestSetZeroBucketFlowRateLimit() {
	operatorAddress := sample.RandAccAddress()
	bucketOwner := sample.RandAccAddress()
	paymentAccount := sample.RandAccAddress()
	bucketName := string(sample.RandStr(10))

	bucketInfo := &types.BucketInfo{
		Owner:            bucketOwner.String(),
		BucketName:       bucketName,
		Id:               sdk.NewUint(1),
		PaymentAddress:   paymentAccount.String(),
		ChargedReadQuota: 100,
	}
	prepareReadStoreBill(s, bucketInfo)

	s.paymentKeeper.EXPECT().IsPaymentAccountOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	s.paymentKeeper.EXPECT().ApplyUserFlowsList(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	err := s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, sdkmath.NewInt(0))
	s.Require().NoError(err)
}

func (s *TestSuite) TestSetNonZeroBucketFlowRateLimit() {
	operatorAddress := sample.RandAccAddress()
	bucketOwner := sample.RandAccAddress()
	paymentAccount := sample.RandAccAddress()
	bucketName := string(sample.RandStr(10))

	bucketInfo := &types.BucketInfo{
		Owner:            bucketOwner.String(),
		BucketName:       bucketName,
		Id:               sdk.NewUint(1),
		PaymentAddress:   paymentAccount.String(),
		ChargedReadQuota: 100,
	}
	prepareReadStoreBill(s, bucketInfo)

	s.paymentKeeper.EXPECT().IsPaymentAccountOwner(gomock.Any(), gomock.Any(), gomock.Any()).Return(true).AnyTimes()
	s.paymentKeeper.EXPECT().ApplyUserFlowsList(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	internalBucketInfo := s.storageKeeper.MustGetInternalBucketInfo(s.ctx, bucketInfo.Id)
	bill, err := s.storageKeeper.GetBucketReadStoreBill(s.ctx, bucketInfo, internalBucketInfo)
	s.Require().NoError(err)
	totalOutFlowRate := getTotalOutFlowRate(bill.Flows)

	// case 1: rate limit does not exist before and is less than total out flow rate
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, totalOutFlowRate.Sub(sdkmath.NewInt(1)))
	s.Require().ErrorContains(err, "greater than the new rate limit")

	// case 2: rate limit does not exist before and is equal to total out flow rate
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, totalOutFlowRate)
	s.Require().NoError(err)

	// case 3: rate limit does not exist before and is greater than total out flow rate
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, totalOutFlowRate.Add(sdkmath.NewInt(1)))
	s.Require().NoError(err)

	// case 4: rate limit exists before and is equal than total out flow rate
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, totalOutFlowRate)
	s.Require().NoError(err)

	// case 5: rate limit exists before and is 0, and the new rate limit is less than total out flow rate
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, sdk.NewInt(0))
	s.Require().NoError(err)
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, totalOutFlowRate.Sub(sdkmath.NewInt(1)))
	s.Require().ErrorContains(err, "greater than the new rate limit")

	// case 6: rate limit exists before and is 0, and the new rate limit is larger than the total out flow rate
	err = s.storageKeeper.SetBucketFlowRateLimit(s.ctx, operatorAddress, bucketOwner, paymentAccount, bucketName, totalOutFlowRate.Add(sdkmath.NewInt(1)))
	s.Require().NoError(err)
}

func getTotalOutFlowRate(flows []paymenttypes.OutFlow) sdkmath.Int {
	totalFlowRate := sdkmath.ZeroInt()
	for _, flow := range flows {
		totalFlowRate = totalFlowRate.Add(flow.Rate)
	}
	return totalFlowRate
}

func prepareReadStoreBill(s *TestSuite, bucketInfo *types.BucketInfo) {
	gvgFamily := &virtualgroupmoduletypes.GlobalVirtualGroupFamily{
		Id:                    1,
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gomock.Any()).
		Return(gvgFamily, true).AnyTimes()

	bucketInfo.GlobalVirtualGroupFamilyId = gvgFamily.Id

	primarySp := &sptypes.StorageProvider{
		Status:          sptypes.STATUS_IN_SERVICE,
		Id:              100,
		OperatorAddress: sample.RandAccAddress().String(),
		FundingAddress:  sample.RandAccAddress().String(),
	}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(primarySp.Id)).
		Return(primarySp, true).AnyTimes()

	price := sptypes.GlobalSpStorePrice{
		ReadPrice:           sdk.NewDec(100),
		PrimaryStorePrice:   sdk.NewDec(1000),
		SecondaryStorePrice: sdk.NewDec(500),
	}
	s.spKeeper.EXPECT().GetGlobalSpStorePriceByTime(gomock.Any(), gomock.Any()).
		Return(price, nil).AnyTimes()
	params := paymenttypes.DefaultParams()
	s.paymentKeeper.EXPECT().GetVersionedParamsWithTs(gomock.Any(), gomock.Any()).
		Return(params.VersionedParams, nil).AnyTimes()

	// none empty bucket
	lvg1 := &types.LocalVirtualGroup{
		Id:                   1,
		TotalChargeSize:      100,
		GlobalVirtualGroupId: 1,
	}
	lvg2 := &types.LocalVirtualGroup{
		Id:                   2,
		TotalChargeSize:      200,
		GlobalVirtualGroupId: 2,
	}
	internalBucketInfo := &types.InternalBucketInfo{
		TotalChargeSize: 300,
		LocalVirtualGroups: []*types.LocalVirtualGroup{
			lvg1, lvg2,
		},
	}

	gvg1 := &virtualgroupmoduletypes.GlobalVirtualGroup{
		Id:                    1,
		PrimarySpId:           primarySp.Id,
		SecondarySpIds:        []uint32{101, 102, 103, 104, 105, 106},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	gvg2 := &virtualgroupmoduletypes.GlobalVirtualGroup{
		Id:                    2,
		PrimarySpId:           primarySp.Id,
		SecondarySpIds:        []uint32{201, 202, 203, 204, 205, 206},
		VirtualPaymentAddress: sample.RandAccAddress().String(),
	}
	s.virtualGroupKeeper.EXPECT().GetGVG(gomock.Any(), gvg1.Id).
		Return(gvg1, true).AnyTimes()
	s.virtualGroupKeeper.EXPECT().GetGVG(gomock.Any(), gvg2.Id).
		Return(gvg2, true).AnyTimes()

	s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)
	s.storageKeeper.SetInternalBucketInfo(s.ctx, bucketInfo.Id, internalBucketInfo)
}
