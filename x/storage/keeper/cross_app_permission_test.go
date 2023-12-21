package keeper_test

import (
	"encoding/json"
	"math/big"
	"math/rand"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
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
	data, err := json.Marshal(&policy)
	s.NoError(err)

	synPackage := storageTypes.CreatePolicySynPackage{
		Operator:  sample.RandAccAddress(),
		Data:      data,
		ExtraData: []byte("extra data"),
	}
	serializedSynPackage := synPackage.MustSerialize()
	serializedSynPackage = append([]byte{storageTypes.OperationCreatePolicy}, serializedSynPackage...)

	// normal case
	permissionKeeper.EXPECT().PutPolicy(gomock.Any(), gomock.Any()).Return(math.NewUint(0), nil)
	res := app.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{}, serializedSynPackage)
	s.Require().NoError(res.Err)
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
