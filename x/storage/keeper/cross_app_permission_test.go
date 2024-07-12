package keeper_test

import (
	"math/big"
	"math/rand"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	types2 "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	storageTypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *TestSuite) TestSynCreatePolicy() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	resourceIds := []math.Uint{math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64())}
	// policy without expiry
	policy := types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceIds[0],
		Statements:     nil,
		ExpirationTime: nil,
	}

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	data, err := policy.Marshal()
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  sample.RandAccAddress(),
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	// case 1: bucket not found
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(nil, false)
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, storageTypes.ErrNoSuchBucket)
}

func (s *TestSuite) TestSynDeletePolicy() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	synPackage := storageTypes.DeleteBucketSynPackage{
		Operator:  sample.RandAccAddress(),
		Id:        big.NewInt(10),
		ExtraData: []byte("extra data"),
	}

	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationDeletePolicy}, serializedSynPackage...)

	// case 1: No such Policy
	permissionKeeper.EXPECT().GetPolicyByID(gomock.Any(), gomock.Any()).Return(&types.Policy{}, false)
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, storageTypes.ErrNoSuchPolicy)
	s.Require().NotEmpty(res.Payload)
}

func (s *TestSuite) TestSynCreatePolicyByMsgErr() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	resourceIds := []math.Uint{math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64())}
	// policy without expiry
	op := sample.RandAccAddress()
	policy := types.CrossChainPolicy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceIds[0],
		Statements:     nil,
		ExpirationTime: nil,
		XResourceGRN: &types.CrossChainPolicy_ResourceGRN{
			ResourceGRN: types2.NewBucketGRN("test-bucket").String(),
		},
	}

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	data, err := policy.Marshal()
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  op,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	// case 1: bucket not found
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&storageTypes.BucketInfo{
		Owner:      op.String(),
		BucketName: "test-bucket",
	}, false).AnyTimes()
	storageKeeper.EXPECT().GetResourceOwnerAndIdFromGRN(gomock.Any(), gomock.Any()).Return(op, resourceIds[0], storageTypes.ErrNoSuchBucket.Wrapf("bucketName: test-bucket")).AnyTimes()
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, storageTypes.ErrNoSuchBucket)
}

func (s *TestSuite) TestSynCreatePolicyByMsg() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	resourceIds := []math.Uint{math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64())}
	// policy without expiry
	op := sample.RandAccAddress()
	policy := types.CrossChainPolicy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceIds[0],
		Statements:     nil,
		ExpirationTime: nil,
		XResourceGRN: &types.CrossChainPolicy_ResourceGRN{
			ResourceGRN: types2.NewBucketGRN("test-bucket").String(),
		},
	}

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	data, err := policy.Marshal()
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  op,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)
	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&storageTypes.BucketInfo{
		Owner:      op.String(),
		BucketName: "test-bucket",
	}, true)
	storageKeeper.EXPECT().GetResourceOwnerAndIdFromGRN(gomock.Any(), gomock.Any()).Return(op, resourceIds[0], nil).AnyTimes()
	storageKeeper.EXPECT().NormalizePrincipal(gomock.Any(), gomock.Any()).Return().AnyTimes()
	storageKeeper.EXPECT().ValidatePrincipal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	permissionKeeper.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).Return(math.NewUint(1), nil).AnyTimes()
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, nil)
}
