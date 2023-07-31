package keeper_test

import (
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

func TestParamsQuery(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := k.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestVersionedParamsQuery(t *testing.T) {
	k, ctx := makeKeeper(t)
	params := types.DefaultParams()
	params.VersionedParams.MaxSegmentSize = 1
	blockTimeT1 := ctx.BlockTime().Unix()
	paramsT1 := params
	err := k.SetParams(ctx, params)
	require.NoError(t, err)

	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(1 * time.Hour))
	blockTimeT2 := ctx.BlockTime().Unix()
	params.VersionedParams.MaxSegmentSize = 2
	paramsT2 := params
	err = k.SetParams(ctx, params)
	require.NoError(t, err)

	responseT1, err := k.QueryParamsByTimestamp(ctx, &types.QueryParamsByTimestampRequest{Timestamp: blockTimeT1})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsByTimestampResponse{Params: paramsT1}, responseT1)
	getParams := responseT1.GetParams()
	require.EqualValues(t, getParams.GetMaxSegmentSize(), 1)

	responseT2, err := k.QueryParamsByTimestamp(ctx, &types.QueryParamsByTimestampRequest{Timestamp: blockTimeT2})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsByTimestampResponse{Params: paramsT2}, responseT2)
	p := responseT2.GetParams()
	require.EqualValues(t, p.GetMaxSegmentSize(), 2)

	responseT3, err := k.QueryParamsByTimestamp(ctx, &types.QueryParamsByTimestampRequest{Timestamp: 0})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsByTimestampResponse{Params: paramsT2}, responseT3)
	p = responseT2.GetParams()
	require.EqualValues(t, p.GetMaxSegmentSize(), 2)
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
	_, err := k.ListGroup(ctx, nil)
	require.ErrorContains(t, err, "invalid request")

	_, err = k.ListGroup(ctx, &types.QueryListGroupRequest{
		Pagination: &query.PageRequest{
			Limit: types.MaxPaginationLimit + 1,
		},
	})
	require.ErrorContains(t, err, "exceed pagination limit")

	_, err = k.ListGroup(ctx, &types.QueryListGroupRequest{
		GroupOwner: "xxx",
	})
	require.ErrorContains(t, err, "invalid address hex length")
}

func TestHeadGroupMember(t *testing.T) {
	// invalid argument
	k, ctx := makeKeeper(t)
	_, err := k.ListGroup(ctx, nil)
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
