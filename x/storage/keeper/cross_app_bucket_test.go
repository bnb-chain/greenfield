package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *TestSuite) TestSynDeleteBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	deleteSynPackage := types.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedSynPackage := deleteSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationDeleteBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: bucket not found
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchBucket)

	// case 2: delete bucket error
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{
		BucketName: "bucket",
	}, true)
	storageKeeper.EXPECT().DeleteBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete error"))
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "delete error")

	// case 3: delete bucket success
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{
		BucketName: "bucket",
		Id:         sdk.NewUint(10),
	}, true)
	storageKeeper.EXPECT().DeleteBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynCreateBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	createSynPackage := types.CreateBucketSynPackage{
		Creator:          sample.RandAccAddress(),
		BucketName:       "bucketname",
		ExtraData:        []byte("extra data"),
		PaymentAddress:   sample.RandAccAddress(),
		PrimarySpAddress: sample.RandAccAddress(),
	}
	serializedSynPackage := createSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: invalid package
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "Invalid type of visibility")

	// case 2: create bucket error
	createSynPackage.Visibility = uint32(types.VISIBILITY_TYPE_PUBLIC_READ)
	serializedSynPackage = createSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().CreateBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.NewUint(1), fmt.Errorf("create error"))
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "create error")

	// case 3: create bucket success
	createSynPackage.Visibility = uint32(types.VISIBILITY_TYPE_PUBLIC_READ)
	serializedSynPackage = createSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().CreateBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.NewUint(1), nil)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynMirrorBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	synPackage := types.MirrorBucketSynPackage{
		Owner: sample.RandAccAddress(),
		Id:    big.NewInt(10),
	}

	serializedSynPack, err := synPackage.Serialize()
	s.Require().NoError(err)
	serializedSynPack = append([]byte{types.OperationMirrorBucket}, serializedSynPack...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1:  normal case
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPack)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckMirrorBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	ackPackage := types.MirrorBucketAckPackage{
		Status: types.StatusSuccess,
		Id:     big.NewInt(10),
	}

	serializedAckPack, err := ackPackage.Serialize()
	s.Require().NoError(err)
	serializedAckPack = append([]byte{types.OperationMirrorBucket}, serializedAckPack...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: bucket not found
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(nil, false)

	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPack)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchBucket)

	// case 2: success case
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{}, true)
	storageKeeper.EXPECT().SetBucketInfo(gomock.Any(), gomock.Any()).Return()

	res = app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPack)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckCreateBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	ackPackage := types.CreateBucketAckPackage{
		Status:    types.StatusSuccess,
		Id:        big.NewInt(10),
		Creator:   sample.RandAccAddress(),
		ExtraData: []byte("extra data"),
	}

	serializedAckPack := ackPackage.MustSerialize()
	serializedAckPack = append([]byte{types.OperationCreateBucket}, serializedAckPack...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1:  normal case
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPack)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckDeleteBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	ackPackage := types.DeleteBucketAckPackage{
		Status:    types.StatusSuccess,
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedAckPack := ackPackage.MustSerialize()
	serializedAckPack = append([]byte{types.OperationDeleteBucket}, serializedAckPack...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1:  normal case
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPack)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckMirrorBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	ackPackage := types.MirrorBucketSynPackage{
		Id:    big.NewInt(10),
		Owner: sample.RandAccAddress(),
	}

	serializedAckPack, err := ackPackage.Serialize()
	s.Require().NoError(err)
	serializedAckPack = append([]byte{types.OperationMirrorBucket}, serializedAckPack...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1:  bucket not found
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{}, false)

	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPack)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchBucket)

	// case 2: normal case
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{}, true)
	storageKeeper.EXPECT().SetBucketInfo(gomock.Any(), gomock.Any()).Return()

	res = app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPack)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckCreateBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	createSynPackage := types.CreateBucketSynPackage{
		Creator:          sample.RandAccAddress(),
		BucketName:       "bucketname",
		ExtraData:        []byte("extra data"),
		PaymentAddress:   sample.RandAccAddress(),
		PrimarySpAddress: sample.RandAccAddress(),
	}
	serializedSynPackage := createSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1:  normal case
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckDeleteBucket() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewBucketApp(storageKeeper)
	deleteSynPackage := types.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedSynPackage := deleteSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationDeleteBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1:  normal case
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}
