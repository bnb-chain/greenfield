package keeper_test

import (
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/payment/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestAutoSettleQueueQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNAutoSettleQueue(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetAutoSettleQueueRequest
		response *types.QueryGetAutoSettleQueueResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetAutoSettleQueueRequest{
				Timestamp: msgs[0].Timestamp,
				User:      msgs[0].Addr,
			},
			response: &types.QueryGetAutoSettleQueueResponse{AutoSettleQueue: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetAutoSettleQueueRequest{
				Timestamp: msgs[1].Timestamp,
				User:      msgs[1].Addr,
			},
			response: &types.QueryGetAutoSettleQueueResponse{AutoSettleQueue: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetAutoSettleQueueRequest{
				Timestamp: 100000,
				User:      strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.AutoSettleQueue(wctx, tc.request)
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

func TestAutoSettleQueueQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNAutoSettleQueue(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllAutoSettleQueueRequest {
		return &types.QueryAllAutoSettleQueueRequest{
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
			resp, err := keeper.AutoSettleQueueAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.AutoSettleQueue), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.AutoSettleQueue),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.AutoSettleQueueAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.AutoSettleQueue), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.AutoSettleQueue),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.AutoSettleQueueAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.AutoSettleQueue),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.AutoSettleQueueAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
