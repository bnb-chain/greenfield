package keeper_test

import (
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *TestSuite) TestAckMirrorGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.MirrorGroupAckPackage{
		Status: types.StatusSuccess,
		Id:     big.NewInt(10),
	}

	serializedAckPackage, err := ackPackage.Serialize()
	s.Require().NoError(err)
	serializedAckPackage = append([]byte{types.OperationMirrorGroup}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: mirror group not found
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchGroup)

	// case 2: normal case
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(&types.GroupInfo{}, true)
	storageKeeper.EXPECT().SetGroupInfo(gomock.Any(), gomock.Any()).Return()
	res = app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckCreateGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.CreateGroupAckPackage{
		Status:    types.StatusSuccess,
		Id:        big.NewInt(10),
		Creator:   sample.RandAccAddress(),
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationCreateGroup}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestAckDeleteGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.DeleteGroupAckPackage{
		Status:    types.StatusSuccess,
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationDeleteGroup}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckMirrorGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.MirrorGroupSynPackage{
		Id:    big.NewInt(10),
		Owner: sample.RandAccAddress(),
	}

	serializedAckPackage, err := ackPackage.Serialize()
	s.Require().NoError(err)
	serializedAckPackage = append([]byte{types.OperationMirrorGroup}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: group not found
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchGroup)

	// case 2: normal case
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(&types.GroupInfo{}, true)
	storageKeeper.EXPECT().SetGroupInfo(gomock.Any(), gomock.Any()).Return()
	res = app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckCreateGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.CreateGroupSynPackage{
		Creator:   sample.RandAccAddress(),
		GroupName: "group",
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationCreateGroup}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckDeleteGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationDeleteGroup}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestFailAckUpdateGroupMember() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	ackPackage := types.UpdateGroupMemberSynPackage{
		Operator:  sample.RandAccAddress(),
		GroupId:   big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedAckPackage := ackPackage.MustSerialize()
	serializedAckPackage = append([]byte{types.OperationUpdateGroupMember}, serializedAckPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedAckPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynMirrorGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	synPackage := types.MirrorGroupSynPackage{
		Owner: sample.RandAccAddress(),
		Id:    big.NewInt(10),
	}

	serializedSynPackage, err := synPackage.Serialize()
	s.Require().NoError(err)
	serializedSynPackage = append([]byte{types.OperationMirrorGroup}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: normal case
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynCreateGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	synPackage := types.CreateGroupSynPackage{
		Creator:   sample.RandAccAddress(),
		GroupName: "group",
		ExtraData: []byte("extra data"),
	}

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: invalid group name
	synPackage.GroupName = "g"
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateGroup}, serializedSynPackage...)

	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, gnfderrors.ErrInvalidGroupName)
	s.Require().NotEmpty(res.Payload)

	// case 2: create group error
	synPackage.GroupName = "group"
	serializedSynPackage = synPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationCreateGroup}, serializedSynPackage...)

	storageKeeper.EXPECT().CreateGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(math.NewUint(0), fmt.Errorf("create group error"))
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "create group error")

	// case 3: normal case
	storageKeeper.EXPECT().CreateGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(math.NewUint(10), nil)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynDeleteGroup() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	synPackage := types.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationDeleteGroup}, serializedSynPackage...)

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: group not exist
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchGroup)
	s.Require().NotEmpty(res.Payload)

	// case 2: delete group error
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(&types.GroupInfo{}, true)
	storageKeeper.EXPECT().DeleteGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("delete group error"))
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "delete group error")

	// case 3: normal case
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(&types.GroupInfo{
		Id: sdk.NewUint(10),
	}, true)
	storageKeeper.EXPECT().DeleteGroup(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

func (s *TestSuite) TestSynUpdateGroupMember() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := types.NewMockStorageKeeper(ctrl)
	storageKeeper.EXPECT().Logger(gomock.Any()).Return(s.ctx.Logger()).AnyTimes()

	app := keeper.NewGroupApp(storageKeeper)
	synPackage := types.UpdateGroupMemberSynPackage{
		Operator:  sample.RandAccAddress(),
		GroupId:   big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	storageKeeper.EXPECT().GetSourceTypeByChainId(gomock.Any(), gomock.Any()).Return(types.SOURCE_TYPE_BSC_CROSS_CHAIN, nil).AnyTimes()

	// case 1: invalid package
	synPackage.OperationType = 3
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationUpdateGroupMember}, serializedSynPackage...)

	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, types.ErrInvalidOperationType)
	s.Require().NotEmpty(res.Payload)

	// case 2: group not exist
	synPackage.OperationType = types.OperationAddGroupMember
	serializedSynPackage = synPackage.MustSerialize()
	serializedSynPackage = append([]byte{types.OperationUpdateGroupMember}, serializedSynPackage...)

	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, types.ErrNoSuchGroup)
	s.Require().NotEmpty(res.Payload)

	// case 3: update group member error
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(&types.GroupInfo{}, true)
	storageKeeper.EXPECT().UpdateGroupMember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("update group member error"))
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorContains(res.Err, "update group member error")

	// case 4: normal case
	storageKeeper.EXPECT().GetGroupInfoById(gomock.Any(), gomock.Any()).Return(&types.GroupInfo{
		Id: sdk.NewUint(10),
	}, true)
	storageKeeper.EXPECT().UpdateGroupMember(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	res = app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}
