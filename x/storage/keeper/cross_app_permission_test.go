package keeper_test

import (
	"math/big"
	"math/rand"
	"time"

	"cosmossdk.io/math"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
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

// TestSynCreatePolicyTaiga_OperatorIsOwner verifies that after the Taiga upgrade,
// creating a policy succeeds when the operator is the resource owner (legacy path, no GRN).
func (s *TestSuite) TestSynCreatePolicyTaiga_OperatorIsOwner() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	owner := sample.RandAccAddress()
	resourceId := math.NewUint(rand.Uint64())

	policy := types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceId,
		Statements:     nil,
		ExpirationTime: nil,
	}

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	data, err := policy.Marshal()
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  owner,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&storageTypes.BucketInfo{
		Owner:      owner.String(),
		BucketName: "test-bucket",
	}, true)
	storageKeeper.EXPECT().NormalizePrincipal(gomock.Any(), gomock.Any()).Return()
	storageKeeper.EXPECT().ValidatePrincipal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	permissionKeeper.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).Return(math.NewUint(1), nil)

	ctx := s.ctxWithUpgrades(upgradetypes.Serengeti, upgradetypes.Mongolian, upgradetypes.Taiga)
	res := app.ExecuteSynPackage(ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

// TestSynCreatePolicyTaiga_OperatorNotOwner verifies that after the Taiga upgrade,
// creating a policy fails when the operator is NOT the resource owner (legacy path, no GRN).
func (s *TestSuite) TestSynCreatePolicyTaiga_OperatorNotOwner() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	operator := sample.RandAccAddress()
	owner := sample.RandAccAddress()
	resourceId := math.NewUint(rand.Uint64())

	policy := types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceId,
		Statements:     nil,
		ExpirationTime: nil,
	}

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	data, err := policy.Marshal()
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  operator,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&storageTypes.BucketInfo{
		Owner:      owner.String(),
		BucketName: "test-bucket",
	}, true)

	ctx := s.ctxWithUpgrades(upgradetypes.Serengeti, upgradetypes.Mongolian, upgradetypes.Taiga)
	res := app.ExecuteSynPackage(ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, storageTypes.ErrAccessDenied)
	s.Require().NotEmpty(res.Payload)
}

// TestSynCreatePolicyPreTaiga_OperatorNotOwner verifies that before the Taiga upgrade,
// creating a policy succeeds even when operator != owner (no ownership check).
func (s *TestSuite) TestSynCreatePolicyPreTaiga_OperatorNotOwner() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	operator := sample.RandAccAddress()
	owner := sample.RandAccAddress()
	resourceId := math.NewUint(rand.Uint64())

	policy := types.Policy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceId,
		Statements:     nil,
		ExpirationTime: nil,
	}

	app := keeper.NewPermissionApp(storageKeeper, permissionKeeper)
	data, err := policy.Marshal()
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  operator,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	storageKeeper.EXPECT().GetBucketInfoById(gomock.Any(), gomock.Any()).Return(&storageTypes.BucketInfo{
		Owner:      owner.String(),
		BucketName: "test-bucket",
	}, true)
	storageKeeper.EXPECT().NormalizePrincipal(gomock.Any(), gomock.Any()).Return()
	storageKeeper.EXPECT().ValidatePrincipal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	permissionKeeper.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).Return(math.NewUint(1), nil)

	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

// TestSynCreatePolicyTaigaGRN_OperatorIsOwner verifies that after the Taiga upgrade
// with the Mongolian GRN path, creating a policy succeeds when operator == owner.
func (s *TestSuite) TestSynCreatePolicyTaigaGRN_OperatorIsOwner() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	owner := sample.RandAccAddress()
	resourceId := math.NewUint(rand.Uint64())

	policy := types.CrossChainPolicy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceId,
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
		Operator:  owner,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	storageKeeper.EXPECT().GetResourceOwnerAndIdFromGRN(gomock.Any(), gomock.Any()).Return(owner, resourceId, nil)
	storageKeeper.EXPECT().NormalizePrincipal(gomock.Any(), gomock.Any()).Return()
	storageKeeper.EXPECT().ValidatePrincipal(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)
	permissionKeeper.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).Return(math.NewUint(1), nil)

	ctx := s.ctxWithUpgrades(upgradetypes.Serengeti, upgradetypes.Mongolian, upgradetypes.Taiga)
	res := app.ExecuteSynPackage(ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
}

// TestSynCreatePolicyTaigaGRN_OperatorNotOwner verifies that after the Taiga upgrade
// with the Mongolian GRN path, creating a policy fails when operator != owner.
func (s *TestSuite) TestSynCreatePolicyTaigaGRN_OperatorNotOwner() {
	ctrl := gomock.NewController(s.T())
	storageKeeper := storageTypes.NewMockStorageKeeper(ctrl)
	permissionKeeper := storageTypes.NewMockPermissionKeeper(ctrl)

	operator := sample.RandAccAddress()
	owner := sample.RandAccAddress()
	resourceId := math.NewUint(rand.Uint64())

	policy := types.CrossChainPolicy{
		Principal: &types.Principal{
			Type:  types.PRINCIPAL_TYPE_GNFD_ACCOUNT,
			Value: sample.RandAccAddressHex(),
		},
		ResourceType:   1,
		ResourceId:     resourceId,
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
		Operator:  operator,
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	storageKeeper.EXPECT().GetResourceOwnerAndIdFromGRN(gomock.Any(), gomock.Any()).Return(owner, resourceId, nil)

	ctx := s.ctxWithUpgrades(upgradetypes.Serengeti, upgradetypes.Mongolian, upgradetypes.Taiga)
	res := app.ExecuteSynPackage(ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().ErrorIs(res.Err, storageTypes.ErrAccessDenied)
	s.Require().NotEmpty(res.Payload)
}

func (s *TestSuite) ctxWithUpgrades(names ...string) sdk.Context {
	key := storetypes.NewKVStoreKey("test_upgrade_ctx")
	testCtx := testutil.DefaultContextWithDB(s.T(), key, storetypes.NewTransientStoreKey("transient_upgrade_ctx"))
	header := testCtx.Ctx.BlockHeader()
	header.Time = time.Now()
	upgradeSet := make(map[string]bool, len(names))
	for _, n := range names {
		upgradeSet[n] = true
	}
	upgradeChecker := func(_ sdk.Context, name string) bool {
		return upgradeSet[name]
	}
	return sdk.NewContext(testCtx.CMS, header, false, upgradeChecker, testCtx.Ctx.Logger())
}
