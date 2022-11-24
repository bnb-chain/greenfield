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

func TestPaymentAccountCountQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPaymentAccountCount(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetPaymentAccountCountRequest
		response *types.QueryGetPaymentAccountCountResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetPaymentAccountCountRequest{
				Owner: msgs[0].Owner,
			},
			response: &types.QueryGetPaymentAccountCountResponse{PaymentAccountCount: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetPaymentAccountCountRequest{
				Owner: msgs[1].Owner,
			},
			response: &types.QueryGetPaymentAccountCountResponse{PaymentAccountCount: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetPaymentAccountCountRequest{
				Owner: strconv.Itoa(100000),
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.PaymentAccountCount(wctx, tc.request)
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

func TestPaymentAccountCountQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNPaymentAccountCount(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllPaymentAccountCountRequest {
		return &types.QueryAllPaymentAccountCountRequest{
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
			resp, err := keeper.PaymentAccountCountAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PaymentAccountCount), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PaymentAccountCount),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.PaymentAccountCountAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.PaymentAccountCount), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.PaymentAccountCount),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.PaymentAccountCountAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.PaymentAccountCount),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.PaymentAccountCountAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
