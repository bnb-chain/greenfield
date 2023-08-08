package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/payment/types"
	sp "github.com/bnb-chain/greenfield/x/sp/types"
)

func TestGetStoragePrice(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)

	primaryPrice := sp.SpStoragePrice{
		ReadPrice:     sdk.NewDecWithPrec(2, 2),
		FreeReadQuota: 0,
		StorePrice:    sdk.NewDecWithPrec(5, 1),
	}
	depKeepers.SpKeeper.EXPECT().GetSpStoragePriceByTime(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(primaryPrice, nil).AnyTimes()

	secondaryPrice := sp.SecondarySpStorePrice{
		StorePrice: sdk.NewDecWithPrec(2, 1),
	}
	depKeepers.SpKeeper.EXPECT().GetSecondarySpStorePriceByTime(gomock.Any(), gomock.Any()).
		Return(secondaryPrice, nil).AnyTimes()

	resp, err := keeper.GetStoragePrice(ctx, types.StoragePriceParams{
		PrimarySp: 1,
		PriceTime: 1,
	})
	require.NoError(t, err)
	require.True(t, resp.ReadPrice.Equal(primaryPrice.ReadPrice))
	require.True(t, resp.PrimaryStorePrice.Equal(primaryPrice.StorePrice))
	require.True(t, resp.SecondaryStorePrice.Equal(secondaryPrice.StorePrice))
}
