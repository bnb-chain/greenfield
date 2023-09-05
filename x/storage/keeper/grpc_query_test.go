package keeper_test

import (
	"context"
	"math/rand"
	"strconv"
	"testing"
	"time"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func makeKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	tStorekey := storetypes.NewTransientStoreKey(types.TStoreKey)

	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		tStorekey,
		&types.MockAccountKeeper{},
		&types.MockSpKeeper{},
		&types.MockPaymentKeeper{},
		&types.MockPermissionKeeper{},
		&types.MockCrossChainKeeper{},
		&types.MockVirtualGroupKeeper{},
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	return k, testCtx.Ctx
}

func (s *TestSuite) TestQueryParams() {
	res, err := s.queryClient.Params(context.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(s.storageKeeper.GetParams(s.ctx), res.GetParams())
}

func (s *TestSuite) TestQueryVersionedParams() {
	params := types.DefaultParams()
	params.VersionedParams.MaxSegmentSize = 1
	blockTimeT1 := s.ctx.BlockTime().Unix()
	paramsT1 := params
	err := s.storageKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	s.ctx = s.ctx.WithBlockTime(s.ctx.BlockTime().Add(1 * time.Hour))
	blockTimeT2 := s.ctx.BlockTime().Unix()
	params.VersionedParams.MaxSegmentSize = 2
	paramsT2 := params
	err = s.storageKeeper.SetParams(s.ctx, params)
	s.Require().NoError(err)

	responseT1, err := s.storageKeeper.QueryParamsByTimestamp(s.ctx, &types.QueryParamsByTimestampRequest{Timestamp: blockTimeT1})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsByTimestampResponse{Params: paramsT1}, responseT1)
	getParams := responseT1.GetParams()
	s.Require().Equal(getParams.GetMaxSegmentSize(), uint64(1))

	responseT2, err := s.storageKeeper.QueryParamsByTimestamp(s.ctx, &types.QueryParamsByTimestampRequest{Timestamp: blockTimeT2})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsByTimestampResponse{Params: paramsT2}, responseT2)
	p := responseT2.GetParams()
	s.Require().Equal(p.GetMaxSegmentSize(), uint64(2))

	responseT3, err := s.storageKeeper.QueryParamsByTimestamp(s.ctx, &types.QueryParamsByTimestampRequest{Timestamp: 0})
	s.Require().NoError(err)
	s.Require().Equal(&types.QueryParamsByTimestampResponse{Params: paramsT2}, responseT3)
	p = responseT2.GetParams()
	s.Require().Equal(p.GetMaxSegmentSize(), uint64(2))
}

func (s *TestSuite) TestQueryGroupMembersExist() {
	groupId := rand.Intn(1000)
	members := make([]string, 3)
	exists := make(map[string]bool)
	for i := 0; i < 3; i++ {
		members[i] = sample.RandAccAddressHex()
		exist := rand.Intn(2)
		if exist == 0 {
			exists[members[i]] = false
			s.permissionKeeper.EXPECT().GetGroupMember(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, false).Times(1)
		} else {
			exists[members[i]] = true
			s.permissionKeeper.EXPECT().GetGroupMember(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, true).Times(1)
		}
	}

	req := &types.QueryGroupMembersExistRequest{
		GroupId: strconv.Itoa(groupId),
		Members: members,
	}
	res, err := s.queryClient.QueryGroupMembersExist(context.Background(), req)
	s.Require().NoError(err)
	s.Require().Equal(exists, res.GetExists())
}

func (s *TestSuite) TestQueryGroupsExist() {
	groupOwner := sample.RandAccAddress()
	groupNames := make([]string, 3)
	exists := make(map[string]bool)
	for i := 0; i < 3; i++ {
		groupNames[i] = string(sample.RandStr(10))
		exist := rand.Intn(2)
		if exist == 0 {
			exists[groupNames[i]] = false
		} else {
			exists[groupNames[i]] = true
			_, err := s.storageKeeper.CreateGroup(s.ctx, groupOwner, groupNames[i], types.CreateGroupOptions{})
			s.Require().NoError(err)
		}
	}

	req := &types.QueryGroupsExistRequest{
		GroupOwner: groupOwner.String(),
		GroupNames: groupNames,
	}
	res, err := s.queryClient.QueryGroupsExist(context.Background(), req)
	s.Require().NoError(err)
	s.Require().Equal(exists, res.GetExists())
}

func (s *TestSuite) TestQueryGroupsExistById() {
	groupIds := make([]string, 3)
	exists := make(map[string]bool)
	for i := 0; i < 3; i++ {
		groupIds[i] = strconv.Itoa(rand.Intn(1000) + 10) // make sure there's no conflict
		exist := rand.Intn(2)
		if exist == 0 {
			exists[groupIds[i]] = false
		} else {
			id, err := s.storageKeeper.CreateGroup(s.ctx, sample.RandAccAddress(), string(sample.RandStr(10)), types.CreateGroupOptions{})
			s.Require().NoError(err)
			groupIds[i] = id.String()
			exists[groupIds[i]] = true
		}
	}

	req := &types.QueryGroupsExistByIdRequest{
		GroupIds: groupIds,
	}
	res, err := s.queryClient.QueryGroupsExistById(context.Background(), req)
	s.Require().NoError(err)
	s.Require().Equal(exists, res.GetExists())
}

func TestHeadBucket(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadBucket(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadBucket(ctx, &types.QueryHeadBucketRequest{
		BucketName: "bucket",
	})
	require.ErrorIs(t, err, types.ErrNoSuchBucket)
}

func TestHeadGroupNFT(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadGroupNFT(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadGroupNFT(ctx, &types.QueryNFTRequest{
		TokenId: "xxx",
	})
	require.ErrorContains(t, err, "invalid token id")

	// group not exist
	_, err = k.HeadGroupNFT(ctx, &types.QueryNFTRequest{
		TokenId: "0",
	})
	require.ErrorIs(t, err, types.ErrNoSuchGroup)
}

func TestHeadObjectNFT(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadObjectNFT(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadObjectNFT(ctx, &types.QueryNFTRequest{
		TokenId: "xxx",
	})
	require.ErrorContains(t, err, "invalid token id")

	// object not exist
	_, err = k.HeadObjectNFT(ctx, &types.QueryNFTRequest{
		TokenId: "0",
	})
	require.ErrorIs(t, err, types.ErrNoSuchObject)
}

func TestHeadBucketNFT(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadBucketNFT(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadBucketNFT(ctx, &types.QueryNFTRequest{
		TokenId: "xxx",
	})
	require.ErrorContains(t, err, "invalid token id")

	// bucket not exist
	_, err = k.HeadBucketNFT(ctx, &types.QueryNFTRequest{
		TokenId: "0",
	})
	require.ErrorIs(t, err, types.ErrNoSuchBucket)
}

func TestHeadBucketById(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadBucketById(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadBucketById(ctx, &types.QueryHeadBucketByIdRequest{
		BucketId: "xxx",
	})
	require.ErrorContains(t, err, "invalid bucket id")

	// bucket not exist
	_, err = k.HeadBucketById(ctx, &types.QueryHeadBucketByIdRequest{
		BucketId: "0",
	})
	require.ErrorIs(t, err, types.ErrNoSuchBucket)
}

func TestHeadObject(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadObject(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	// object not exist
	_, err = k.HeadObject(ctx, &types.QueryHeadObjectRequest{
		BucketName: "bucket",
		ObjectName: "object",
	})
	require.ErrorIs(t, err, types.ErrNoSuchObject)
}

func TestHeadObjectById(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadBucketById(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadObjectById(ctx, &types.QueryHeadObjectByIdRequest{
		ObjectId: "xxx",
	})
	require.ErrorContains(t, err, "invalid object id")

	// bucket not exist
	_, err = k.HeadObjectById(ctx, &types.QueryHeadObjectByIdRequest{
		ObjectId: "1",
	})
	require.ErrorIs(t, err, types.ErrNoSuchObject)
}

func TestListBuckets(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.ListBuckets(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.ListBuckets(ctx, &types.QueryListBucketsRequest{
		Pagination: &query.PageRequest{
			Limit: types.MaxPaginationLimit + 1,
		},
	})
	require.ErrorContains(t, err, "exceed pagination limit")
}

func TestListObjects(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.ListObjects(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.ListObjects(ctx, &types.QueryListObjectsRequest{
		Pagination: &query.PageRequest{
			Limit: types.MaxPaginationLimit,
		},
	})
	require.ErrorContains(t, err, "bucket name should not be empty")

	_, err = k.ListObjects(ctx, &types.QueryListObjectsRequest{
		BucketName: "abc",
		Pagination: &query.PageRequest{
			Limit: types.MaxPaginationLimit + 1,
		},
	})
	require.ErrorContains(t, err, "exceed pagination limit")
}

func TestListObjectsByBucketId(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.ListObjectsByBucketId(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.ListObjectsByBucketId(ctx, &types.QueryListObjectsByBucketIdRequest{
		Pagination: &query.PageRequest{
			Limit: types.MaxPaginationLimit + 1,
		},
	})
	require.ErrorContains(t, err, "exceed pagination limit")

	_, err = k.ListObjectsByBucketId(ctx, &types.QueryListObjectsByBucketIdRequest{
		BucketId: "xxx",
	})
	require.ErrorContains(t, err, "invalid bucket id")

	_, err = k.ListObjectsByBucketId(ctx, &types.QueryListObjectsByBucketIdRequest{
		BucketId: "0",
	})
	require.ErrorIs(t, err, types.ErrNoSuchBucket)
}

func TestQueryPolicyForAccount(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.QueryPolicyForAccount(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.QueryPolicyForAccount(ctx, &types.QueryPolicyForAccountRequest{
		PrincipalAddress: "xxxx",
	})
	require.ErrorContains(t, err, "invalid address hex length")

	_, err = k.QueryPolicyForAccount(ctx, &types.QueryPolicyForAccountRequest{
		PrincipalAddress: sample.RandAccAddressHex(),
		Resource:         "xxx",
	})
	require.ErrorContains(t, err, "regex match error")
}

func TestQueryPolicyForGroup(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.QueryPolicyForGroup(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.QueryPolicyForGroup(ctx, &types.QueryPolicyForGroupRequest{
		PrincipalGroupId: "xxx",
	})
	require.ErrorContains(t, err, "invalid group id")

	_, err = k.QueryPolicyForGroup(ctx, &types.QueryPolicyForGroupRequest{
		PrincipalGroupId: "10",
		Resource:         "xxx",
	})
	require.ErrorContains(t, err, "regex match error")
}

func TestVerifyPermission(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.VerifyPermission(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.VerifyPermission(ctx, &types.QueryVerifyPermissionRequest{
		Operator: "xxx",
	})
	require.ErrorContains(t, err, "invalid operator address")

	_, err = k.VerifyPermission(ctx, &types.QueryVerifyPermissionRequest{
		Operator:   sample.RandAccAddressHex(),
		BucketName: "",
	})
	require.ErrorContains(t, err, "No bucket specified")

	_, err = k.VerifyPermission(ctx, &types.QueryVerifyPermissionRequest{
		Operator:   sample.RandAccAddressHex(),
		BucketName: "bucket",
	})
	require.ErrorIs(t, err, types.ErrNoSuchBucket)
}

func TestHeadGroup(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.HeadGroup(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadGroup(ctx, &types.QueryHeadGroupRequest{
		GroupOwner: "xxx",
	})
	require.ErrorContains(t, err, "invalid address hex length")

	_, err = k.HeadGroup(ctx, &types.QueryHeadGroupRequest{
		GroupOwner: sample.RandAccAddressHex(),
		GroupName:  "group",
	})
	require.ErrorIs(t, err, types.ErrNoSuchGroup)
}

func TestListGroup(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.ListGroups(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.ListGroups(ctx, &types.QueryListGroupsRequest{
		Pagination: &query.PageRequest{
			Limit: types.MaxPaginationLimit + 1,
		},
	})
	require.ErrorContains(t, err, "exceed pagination limit")

	_, err = k.ListGroups(ctx, &types.QueryListGroupsRequest{
		GroupOwner: "xxx",
	})
	require.ErrorContains(t, err, "invalid address hex length")
}

func TestHeadGroupMember(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.ListGroups(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.HeadGroupMember(ctx, &types.QueryHeadGroupMemberRequest{
		Member: "xxx",
	})
	require.ErrorContains(t, err, "invalid address hex length")

	_, err = k.HeadGroupMember(ctx, &types.QueryHeadGroupMemberRequest{
		Member:     sample.RandAccAddressHex(),
		GroupOwner: "xxx",
	})
	require.ErrorContains(t, err, "invalid address hex length")

	_, err = k.HeadGroupMember(ctx, &types.QueryHeadGroupMemberRequest{
		Member:     sample.RandAccAddressHex(),
		GroupOwner: sample.RandAccAddressHex(),
		GroupName:  "group",
	})
	require.ErrorIs(t, err, types.ErrNoSuchGroup)
}

func TestQueryPolicyById(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.QueryPolicyById(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.QueryPolicyById(ctx, &types.QueryPolicyByIdRequest{
		PolicyId: "xxx",
	})
	require.ErrorContains(t, err, "invalid policy id")
}

func TestQueryQuotaUpdateTime(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.QueryQuotaUpdateTime(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	// bucket not exist
	_, err = k.QueryQuotaUpdateTime(ctx, &types.QueryQuoteUpdateTimeRequest{
		BucketName: "xxx",
	})
	require.ErrorIs(t, err, types.ErrNoSuchBucket)
}
