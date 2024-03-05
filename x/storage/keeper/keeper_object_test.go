package keeper_test

import (
	"cosmossdk.io/math"
	types2 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/types/common"
	types4 "github.com/bnb-chain/greenfield/x/payment/types"
	types3 "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *TestSuite) TestCreateObject() {
	operatorAddress := sample.RandAccAddress()
	objectName := "objectName"

	bucketInfo := &types.BucketInfo{
		Owner:            operatorAddress.String(),
		BucketName:       "bucketname",
		Id:               sdk.NewUint(1),
		PaymentAddress:   sample.RandAccAddress().String(),
		ChargedReadQuota: 100,
		BucketStatus:     types.BUCKET_STATUS_CREATED,
	}

	// case 1: bucket does not exist
	_, err := s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
		objectName, 100, types.CreateObjectOptions{
			Visibility:        0,
			ContentType:       "",
			SourceType:        0,
			RedundancyType:    0,
			Checksums:         nil,
			PrimarySpApproval: nil,
			ApprovalMsgBytes:  nil,
		})
	s.Require().ErrorContains(err, "No such bucket")

	// case 2: bucket is migrating
	bucketInfo.BucketStatus = types.BUCKET_STATUS_MIGRATING
	s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)
	_, err = s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
		objectName, 100, types.CreateObjectOptions{
			Visibility:        0,
			ContentType:       "",
			SourceType:        0,
			RedundancyType:    0,
			Checksums:         nil,
			PrimarySpApproval: nil,
			ApprovalMsgBytes:  nil,
		})
	s.Require().ErrorContains(err, "the bucket is migrating")

	// case 3: bucket is discontinued
	bucketInfo.BucketStatus = types.BUCKET_STATUS_DISCONTINUED
	s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)
	_, err = s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
		objectName, 100, types.CreateObjectOptions{
			Visibility:        0,
			ContentType:       "",
			SourceType:        0,
			RedundancyType:    0,
			Checksums:         nil,
			PrimarySpApproval: nil,
			ApprovalMsgBytes:  nil,
		})
	s.Require().ErrorContains(err, "the bucket is discontinued")

	// case 4: invalid payload size
	_, err = s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
		objectName, types.DefaultParams().MaxPayloadSize+1, types.CreateObjectOptions{
			Visibility:        0,
			ContentType:       "",
			SourceType:        0,
			RedundancyType:    0,
			Checksums:         nil,
			PrimarySpApproval: nil,
			ApprovalMsgBytes:  nil,
		})
	s.Require().ErrorContains(err, "Object payload size is too large")

	// case 4: gvg family does not exist
	bucketInfo.BucketStatus = types.BUCKET_STATUS_CREATED
	s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gomock.Any()).Return(nil, false)
	s.Require().Panics(func() {
		_, _ = s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
			objectName, 100, types.CreateObjectOptions{
				Visibility:        0,
				ContentType:       "",
				SourceType:        0,
				RedundancyType:    0,
				Checksums:         nil,
				PrimarySpApproval: nil,
				ApprovalMsgBytes:  nil,
			})
	})

	// case 5: object exist
	s.storageKeeper.StoreObjectInfo(s.ctx, &types.ObjectInfo{
		Id:         sdk.NewUint(1),
		BucketName: bucketInfo.BucketName,
		ObjectName: objectName,
	})
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), gomock.Any()).Return(&types2.GlobalVirtualGroupFamily{
		Id:                    0,
		PrimarySpId:           0,
		GlobalVirtualGroupIds: nil,
		VirtualPaymentAddress: "",
	}, true).AnyTimes()
	spAddress, _, _ := sample.RandSignBytes()
	s.spKeeper.EXPECT().MustGetStorageProvider(gomock.Any(), gomock.Any()).Return(&types3.StorageProvider{
		Id:              0,
		OperatorAddress: spAddress.String(),
		FundingAddress:  "",
		SealAddress:     "",
		ApprovalAddress: spAddress.String(),
		GcAddress:       "",
		TotalDeposit:    math.Int{},
		Status:          0,
		Endpoint:        "",
		Description:     types3.Description{},
		BlsKey:          nil,
	}).AnyTimes()
	s.ctx = s.ctx.WithBlockHeight(100)
	_, err = s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
		objectName, 100, types.CreateObjectOptions{
			Visibility:     0,
			ContentType:    "",
			SourceType:     0,
			RedundancyType: 0,
			Checksums:      nil,
		})
	s.Require().ErrorContains(err, "Object already exists")

	// case 6: valid case
	s.storageKeeper.DeleteObjectInfo(s.ctx, &types.ObjectInfo{
		Id:         sdk.NewUint(1),
		BucketName: bucketInfo.BucketName,
		ObjectName: objectName,
	})
	s.spKeeper.EXPECT().GetGlobalSpStorePriceByTime(gomock.Any(), gomock.Any()).Return(types3.GlobalSpStorePrice{
		ReadPrice:           sdk.NewDec(1),
		PrimaryStorePrice:   sdk.NewDec(2),
		SecondaryStorePrice: sdk.NewDec(1),
	}, nil).AnyTimes()
	s.paymentKeeper.EXPECT().GetVersionedParamsWithTs(gomock.Any(), gomock.Any()).Return(types4.VersionedParams{
		ReserveTime:      10000,
		ValidatorTaxRate: sdk.NewDec(1),
	}, nil).AnyTimes()
	s.paymentKeeper.EXPECT().UpdateStreamRecordByAddr(gomock.Any(), gomock.Any()).Return(&types4.StreamRecord{
		Account:           "",
		CrudTimestamp:     0,
		NetflowRate:       math.Int{},
		StaticBalance:     sdk.NewInt(100),
		BufferBalance:     math.Int{},
		LockBalance:       math.Int{},
		Status:            0,
		SettleTimestamp:   0,
		OutFlowCount:      0,
		FrozenNetflowRate: math.Int{},
	}, nil).AnyTimes()
	_, err = s.storageKeeper.CreateObject(s.ctx, operatorAddress, bucketInfo.BucketName,
		objectName, 100, types.CreateObjectOptions{
			Visibility:     0,
			ContentType:    "",
			SourceType:     0,
			RedundancyType: 0,
			Checksums:      nil,
			PrimarySpApproval: &common.Approval{
				ExpiredHeight: uint64(s.ctx.BlockHeight() + 1),
			},
		})

	s.Require().NoError(err)
}
