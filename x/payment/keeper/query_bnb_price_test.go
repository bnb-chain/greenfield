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

func TestBnbPriceQuerySingle(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBnbPrice(keeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetBnbPriceRequest
		response *types.QueryGetBnbPriceResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetBnbPriceRequest{
				Time: msgs[0].Time,
			},
			response: &types.QueryGetBnbPriceResponse{BnbPrice: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetBnbPriceRequest{
				Time: msgs[1].Time,
			},
			response: &types.QueryGetBnbPriceResponse{BnbPrice: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetBnbPriceRequest{
				Time: 100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := keeper.BnbPrice(wctx, tc.request)
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

func TestBnbPriceQueryPaginated(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNBnbPrice(keeper, ctx, 5)

	request := func(next []byte, offset, limit uint64, total bool) *types.QueryAllBnbPriceRequest {
		return &types.QueryAllBnbPriceRequest{
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
			resp, err := keeper.BnbPriceAll(wctx, request(nil, uint64(i), uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.BnbPrice), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.BnbPrice),
			)
		}
	})
	t.Run("ByKey", func(t *testing.T) {
		step := 2
		var next []byte
		for i := 0; i < len(msgs); i += step {
			resp, err := keeper.BnbPriceAll(wctx, request(next, 0, uint64(step), false))
			require.NoError(t, err)
			require.LessOrEqual(t, len(resp.BnbPrice), step)
			require.Subset(t,
				nullify.Fill(msgs),
				nullify.Fill(resp.BnbPrice),
			)
			next = resp.Pagination.NextKey
		}
	})
	t.Run("Total", func(t *testing.T) {
		resp, err := keeper.BnbPriceAll(wctx, request(nil, 0, 0, true))
		require.NoError(t, err)
		require.Equal(t, len(msgs), int(resp.Pagination.Total))
		require.ElementsMatch(t,
			nullify.Fill(msgs),
			nullify.Fill(resp.BnbPrice),
		)
	})
	t.Run("InvalidRequest", func(t *testing.T) {
		_, err := keeper.BnbPriceAll(wctx, nil)
		require.ErrorIs(t, err, status.Error(codes.InvalidArgument, "invalid request"))
	})
}
