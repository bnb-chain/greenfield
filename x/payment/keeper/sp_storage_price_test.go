package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
	"reflect"
	"testing"
	"time"
)

func TestGetSpStoragePriceByTime(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	spStoragePrice := types.SpStoragePrice{
		SpAddress:      "sp",
		UpdateTime:     1,
		ReadQuotaPrice: sdkmath.NewInt(100),
		StorePrice:     sdkmath.NewInt(100),
	}
	keeper.SetSpStoragePrice(ctx, spStoragePrice)
	spStoragePrice2 := types.SpStoragePrice{
		SpAddress:      "sp",
		UpdateTime:     100,
		ReadQuotaPrice: sdkmath.NewInt(200),
		StorePrice:     sdkmath.NewInt(200),
	}
	keeper.SetSpStoragePrice(ctx, spStoragePrice2)
	type args struct {
		time int64
	}
	tests := []struct {
		name    string
		args    args
		wantVal types.SpStoragePrice
		wantErr bool
	}{
		{"test 0", args{time: 0}, types.SpStoragePrice{}, true},
		{"test 1", args{time: 1}, spStoragePrice, false},
		{"test 2", args{time: 2}, spStoragePrice, false},
		{"test 100", args{time: 100}, spStoragePrice2, false},
		{"test 101", args{time: 101}, spStoragePrice2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := keeper.GetSpStoragePriceByTime(ctx, "sp", tt.args.time)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetSpStoragePriceByTime() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotVal, tt.wantVal) {
				t.Errorf("GetSpStoragePriceByTime() gotVal = %v, want %v", gotVal, tt.wantVal)
			}
		})
	}
}
