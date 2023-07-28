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

	// case 1: bucket not found
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteSynPackage(s.ctx, nil, serializedSynPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchBucket)

	// case 2: delete bucket error
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{
		BucketName: "bucket",
	}, true)
	storageKeeper.EXPECT().DeleteBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete error"))
	res = app.ExecuteSynPackage(s.ctx, nil, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "delete error")

	// case 3: delete bucket success
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&types.BucketInfo{
		BucketName: "bucket",
		Id:         sdk.NewUint(10),
	}, true)
	storageKeeper.EXPECT().DeleteBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	res = app.ExecuteSynPackage(s.ctx, nil, serializedSynPackage)
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

	// case 1: invalid package
	res := app.ExecuteSynPackage(s.ctx, nil, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "Invalid type of visibility")

	// case 2: create bucket error
	createSynPackage.Visibility = uint32(types.VISIBILITY_TYPE_PUBLIC_READ)
	serializedSynPackage = createSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().CreateBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.NewUint(1), fmt.Errorf("create error"))
	res = app.ExecuteSynPackage(s.ctx, nil, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "create error")

	// case 3: create bucket success
	createSynPackage.Visibility = uint32(types.VISIBILITY_TYPE_PUBLIC_READ)
	serializedSynPackage = createSynPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateBucket}, serializedSynPackage...)

	storageKeeper.EXPECT().CreateBucket(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(sdk.NewUint(1), nil)
	res = app.ExecuteSynPackage(s.ctx, nil, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckMirrorBucket() {

}
