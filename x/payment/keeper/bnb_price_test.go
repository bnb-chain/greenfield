package keeper_test

import (
	"cosmossdk.io/math"
	"reflect"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/payment/keeper"
	"github.com/bnb-chain/bfs/x/payment/types"
)

func createTestBnbPrice(keeper *keeper.Keeper, ctx sdk.Context) types.BnbPrice {
	item := types.BnbPrice{}
	keeper.SetBnbPrice(ctx, item)
	return item
}

func TestBnbPriceGet(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	item := createTestBnbPrice(k, ctx)
	rst, found := k.GetBnbPrice(ctx)
	require.True(t, found)
	require.Equal(t,
		nullify.Fill(&item),
		nullify.Fill(&rst),
	)
}

func TestBnbPriceRemove(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	createTestBnbPrice(k, ctx)
	k.RemoveBnbPrice(ctx)
	_, found := k.GetBnbPrice(ctx)
	require.False(t, found)
}

func TestGetBNBPrice(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	k.SubmitBNBPrice(ctx, 1000, 1000)
	k.SubmitBNBPrice(ctx, 1234, 1234)
	k.SubmitBNBPrice(ctx, 2345, 2345)
}

func TestKeeper_GetBNBPriceByTime(t *testing.T) {
	k, ctx := keepertest.PaymentKeeper(t)
	k.SubmitBNBPrice(ctx, 1000, 1000)
	k.SubmitBNBPrice(ctx, 1234, 1234)
	k.SubmitBNBPrice(ctx, 2345, 2345)
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
		{"test 0", args{0}, math.NewInt(1000), math.NewInt(100000000), false},
		{"test 345", args{345}, math.NewInt(1000), math.NewInt(100000000), false},
		{"test 1001", args{1001}, math.NewInt(1000), math.NewInt(100000000), false},
		{"test 1245", args{1245}, math.NewInt(1234), math.NewInt(100000000), false},
		{"test 2345", args{2345}, math.NewInt(2345), math.NewInt(100000000), false},
		{"test 2346", args{2346}, math.NewInt(2345), math.NewInt(100000000), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNum, gotPrecision, err := k.GetBNBPriceByTime(ctx, tt.args.priceTime)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBNBPriceByTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotNum, tt.wantNum) {
				t.Errorf("GetBNBPriceByTime() gotNum = %v, want %v", gotNum, tt.wantNum)
			}
			if !reflect.DeepEqual(gotPrecision, tt.wantPrecision) {
				t.Errorf("GetBNBPriceByTime() gotPrecision = %v, want %v", gotPrecision, tt.wantPrecision)
			}
		})
	}
}
