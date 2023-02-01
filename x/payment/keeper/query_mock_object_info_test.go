package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestMockObjectInfoQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNMockObjectInfo(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetMockObjectInfoRequest
		response *types.QueryGetMockObjectInfoResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetMockObjectInfoRequest{
				BucketName: msgs[0].BucketName,
				ObjectName: msgs[0].ObjectName,
			},
			response: &types.QueryGetMockObjectInfoResponse{MockObjectInfo: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetMockObjectInfoRequest{
				BucketName: msgs[1].BucketName,
				ObjectName: msgs[1].ObjectName,
			},
			response: &types.QueryGetMockObjectInfoResponse{MockObjectInfo: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetMockObjectInfoRequest{
				BucketName: strconv.Itoa(100000),
				ObjectName: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.MockObjectInfo(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.NoError(t, err)
				require.Equal(t,
					nullify.Fill(tc.response),
					nullify.Fill(response),
				)
			}
		})
	}
}

func TestMockObjectInfoQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNMockObjectInfo(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllMockObjectInfoRequest {
		return &types.QueryAllMockObjectInfoRequest{
			Pagination: &query.PageRequest{
				Key:        next,
				Offset:     offset,
				Limit:      limit,
				CountTotal: total,
			},
		}
	}
	t.Run("ByOffset", func(t *testing.T) {
		step := 2
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MockObjectInfoAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.MockObjectInfo), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.MockObjectInfo),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.MockObjectInfoAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.MockObjectInfo), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.MockObjectInfo),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.MockObjectInfoAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.MockObjectInfo),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.MockObjectInfoAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
