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
	pack := types.DeleteBucketAckPackage{
		Status:    1,
		Id:        big.NewInt(10),
		ExtraData: []byte("x"),
	}
	pack.MustSerialize()
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
