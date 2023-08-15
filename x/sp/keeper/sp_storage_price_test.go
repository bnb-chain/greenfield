package keeper_test

import (
	"reflect"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *KeeperTestSuite) TestGetSpStoragePriceByTime() {
	ctx := s.ctx.WithBlockTime(time.Unix(100, 0))
	spId := uint32(10)

	_, found := s.spKeeper.GetSpStoragePrice(ctx, spId)
	s.Require().True(!found)

	spStoragePrice := types.SpStoragePrice{
		SpId:          spId,
		UpdateTimeSec: 1,
		ReadPrice:     sdk.NewDec(100),
		StorePrice:    sdk.NewDec(100),
	}
	s.spKeeper.SetSpStoragePrice(ctx, spStoragePrice)

	price, found := s.spKeeper.GetSpStoragePrice(ctx, spId)
	s.Require().True(found)
	s.Require().True(reflect.DeepEqual(price, spStoragePrice))

	spStoragePrice2 := types.SpStoragePrice{
		SpId:          spId,
		UpdateTimeSec: 100,
		ReadPrice:     sdk.NewDec(200),
		StorePrice:    sdk.NewDec(200),
	}
	s.spKeeper.SetSpStoragePrice(ctx, spStoragePrice2)

	price, found = s.spKeeper.GetSpStoragePrice(ctx, spId)
	s.Require().True(found)
	s.Require().True(reflect.DeepEqual(price, spStoragePrice2))
}

func (s *KeeperTestSuite) TestGetGlobalSpStorePriceByTime() {
	keeper := s.spKeeper
	ctx := s.ctx
	secondarySpStorePrice := types.GlobalSpStorePrice{
		UpdateTimeSec:       1,
		PrimaryStorePrice:   sdk.NewDec(100),
		SecondaryStorePrice: sdk.NewDec(40),
		ReadPrice:           sdk.NewDec(80),
	}
	keeper.SetGlobalSpStorePrice(ctx, secondarySpStorePrice)
	secondarySpStorePrice2 := types.GlobalSpStorePrice{
		UpdateTimeSec:       100,
		PrimaryStorePrice:   sdk.NewDec(200),
		SecondaryStorePrice: sdk.NewDec(70),
		ReadPrice:           sdk.NewDec(90),
	}
	keeper.SetGlobalSpStorePrice(ctx, secondarySpStorePrice2)
	type args struct {
		time int64
	}
	tests := []struct {
		name    string
		args    args
		wantVal types.GlobalSpStorePrice
		wantErr bool
	}{
		{"test 0", args{time: 0}, types.GlobalSpStorePrice{}, true},
		{"test 1", args{time: 1}, types.GlobalSpStorePrice{}, true},
		{"test 2", args{time: 2}, secondarySpStorePrice, false},
		{"test 100", args{time: 100}, secondarySpStorePrice, false},
		{"test 101", args{time: 101}, secondarySpStorePrice2, false},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			gotVal, err := keeper.GetGlobalSpStorePriceByTime(ctx, tt.args.time)
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
