package keeper_test

import (
	"reflect"
	"strconv"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/nullify"
	"github.com/bnb-chain/greenfield/x/payment/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func createNBnbPrice(keeper *keeper.Keeper, ctx sdk.Context, n int) []types.BnbPrice {
	items := make([]types.BnbPrice, n)
	for i := range items {
		items[i].Time = int64(i)

		keeper.SetBnbPrice(ctx, items[i])
	}
	return items
}

func TestBnbPriceGet(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNBnbPrice(keeper, ctx, 10)
	for _, item := range items {
		rst, found := keeper.GetBnbPrice(ctx,
			item.Time,
		)
		require.True(t, found)
		require.Equal(t,
			nullify.Fill(&item),
			nullify.Fill(&rst),
		)
	}
}
func TestBnbPriceRemove(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNBnbPrice(keeper, ctx, 10)
	for _, item := range items {
		keeper.RemoveBnbPrice(ctx,
			item.Time,
		)
		_, found := keeper.GetBnbPrice(ctx,
			item.Time,
		)
		require.False(t, found)
	}
}

func TestBnbPriceGetAll(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	items := createNBnbPrice(keeper, ctx, 10)
	require.ElementsMatch(t,
		nullify.Fill(items),
		nullify.Fill(keeper.GetAllBnbPrice(ctx)),
	)
}

func TestKeeper_GetBNBPriceByTime(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	k.SubmitBNBPrice(ctx, 1, 1)
	k.SubmitBNBPrice(ctx, 3, 3)
	k.SubmitBNBPrice(ctx, 4, 4)
	k.SubmitBNBPrice(ctx, 1000, 1000)
	k.SubmitBNBPrice(ctx, 1234, 1234)
	k.SubmitBNBPrice(ctx, 2345, 2345)
	k.GetAllBnbPrice(ctx)
	type args struct {
		priceTime int64
	}
	tests := []struct {
		name          string
		args          args
		wantNum       math.Int
		wantPrecision math.Int
		wantErr       bool
	}{
		{"test 0", args{0}, math.NewInt(0), math.NewInt(100000000), true},
		{"test 1", args{1}, math.NewInt(1), math.NewInt(100000000), false},
		{"test 2", args{2}, math.NewInt(1), math.NewInt(100000000), false},
		{"test 3", args{3}, math.NewInt(3), math.NewInt(100000000), false},
		{"test 4", args{4}, math.NewInt(4), math.NewInt(100000000), false},
		{"test 5", args{5}, math.NewInt(4), math.NewInt(100000000), false},
		{"test 345", args{345}, math.NewInt(4), math.NewInt(100000000), false},
		{"test 1000", args{1000}, math.NewInt(1000), math.NewInt(100000000), false},
		{"test 1001", args{1001}, math.NewInt(1000), math.NewInt(100000000), false},
		{"test 1245", args{1245}, math.NewInt(1234), math.NewInt(100000000), false},
		{"test 2345", args{2345}, math.NewInt(2345), math.NewInt(100000000), false},
		{"test 2346", args{2346}, math.NewInt(2345), math.NewInt(100000000), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			price, err := k.GetBNBPriceByTime(ctx, tt.args.priceTime)
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetBNBPriceByTime() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("GetBNBPriceByTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(price.Num, tt.wantNum) {
				t.Errorf("GetBNBPriceByTime() gotNum = %v, want %v", price.Num, tt.wantNum)
			}
			if !reflect.DeepEqual(price.Precision, tt.wantPrecision) {
				t.Errorf("GetBNBPriceByTime() gotPrecision = %v, want %v", price.Precision, tt.wantPrecision)
			}
		})
	}
}
