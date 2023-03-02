package keeper_test

import (
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func TestGetSpStoragePriceByTime(t *testing.T) {
	keeper, ctx := keepertest.SpKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	spStoragePrice := types.SpStoragePrice{
		SpAddress:  "sp",
		UpdateTime: 1,
		ReadPrice:  sdk.NewDec(100),
		StorePrice: sdk.NewDec(100),
	}
	keeper.SetSpStoragePrice(ctx, spStoragePrice)
	spStoragePrice2 := types.SpStoragePrice{
		SpAddress:  "sp",
		UpdateTime: 100,
		ReadPrice:  sdk.NewDec(200),
		StorePrice: sdk.NewDec(200),
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

func TestKeeper_GetSecondarySpStorePriceByTime(t *testing.T) {
	keeper, ctx := keepertest.SpKeeper(t)
	secondarySpStorePrice := types.SecondarySpStorePrice{
		UpdateTime: 1,
		StorePrice: sdk.NewDec(100),
	}
	keeper.SetSecondarySpStorePrice(ctx, secondarySpStorePrice)
	secondarySpStorePrice2 := types.SecondarySpStorePrice{
		UpdateTime: 100,
		StorePrice: sdk.NewDec(200),
	}
	keeper.SetSecondarySpStorePrice(ctx, secondarySpStorePrice2)
	type args struct {
		time int64
	}
	tests := []struct {
		name    string
		args    args
		wantVal types.SecondarySpStorePrice
		wantErr bool
	}{
		{"test 0", args{time: 0}, types.SecondarySpStorePrice{}, true},
		{"test 1", args{time: 1}, secondarySpStorePrice, false},
		{"test 2", args{time: 2}, secondarySpStorePrice, false},
		{"test 100", args{time: 100}, secondarySpStorePrice2, false},
		{"test 101", args{time: 101}, secondarySpStorePrice2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotVal, err := keeper.GetSecondarySpStorePriceByTime(ctx, tt.args.time)
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
