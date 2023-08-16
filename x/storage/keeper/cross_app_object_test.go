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

func (s *TestSuite) TestAckMirrorObject() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewObjectApp(storageKeeper)
	ackPackage := types.MirrorObjectAckPackage{
		Status: types.StatusSuccess,
		Id:     big.NewInt(10),
	}

	serializedAckPackage, err := ackPackage.Serialize()
	s.Require().NoError(err)
	serializedAckPackage = append([]byte{types.OperationMirrorObject}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: object not exist
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchObject)

	// case 2: normal case
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(&types.ObjectInfo{}, true)
	storageKeeper.EXPECT().SetObjectInfo(gomock.Any(), gomock.Any()).Return()
	res = app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckDeleteObject() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewObjectApp(storageKeeper)
	ackPackage := types.DeleteObjectAckPackage{
		Status:    types.StatusSuccess,
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationDeleteObject}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckMirrorObject() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewObjectApp(storageKeeper)
	ackPackage := types.MirrorObjectSynPackage{
		Owner: sample.RandAccAddress(),
		Id:    big.NewInt(10),
	}

	serializedAckPackage, err := ackPackage.Serialize()
	s.Require().NoError(err)
	serializedAckPackage = append([]byte{types.OperationMirrorObject}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: object not exist
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchObject)

	// case 2: normal case
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(&types.ObjectInfo{}, true)
	storageKeeper.EXPECT().SetObjectInfo(gomock.Any(), gomock.Any()).Return()
	res = app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckDeleteObject() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewObjectApp(storageKeeper)
	ackPackage := types.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationDeleteObject}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynMirrorObject() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewObjectApp(storageKeeper)
	synPackage := types.MirrorObjectSynPackage{
		Owner: sample.RandAccAddress(),
		Id:    big.NewInt(10),
	}

	serializedSynPackage, err := synPackage.Serialize()
	s.Require().NoError(err)
	serializedSynPackage = append([]byte{types.OperationMirrorObject}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynDeleteObject() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewObjectApp(storageKeeper)
	synPackage := types.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationDeleteObject}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: object not exist
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchObject)

	// case 2: delete object error
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(&types.ObjectInfo{}, true)
	storageKeeper.EXPECT().DeleteObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete object error"))
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "delete object error")

	// case 3: normal case
	storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Any()).Return(&types.ObjectInfo{
		Id: sdk.NewUint(10),
	}, true)
	storageKeeper.EXPECT().DeleteObject(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}
