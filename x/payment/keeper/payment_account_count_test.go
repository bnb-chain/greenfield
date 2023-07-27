package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestPaymentAccountCount(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)

	owner1 := sample.RandAccAddress()
	paymentCount1 := &types.PaymentAccountCount{
		Owner: owner1.String(),
		Count: 1,
	}

	owner2 := sample.RandAccAddress()
	paymentCount2 := &types.PaymentAccountCount{
		Owner: owner2.String(),
		Count: 3,
	}

	// set
	keeper.SetPaymentAccountCount(ctx, paymentCount1)
	keeper.SetPaymentAccountCount(ctx, paymentCount2)

	// get
	resp1, _ := keeper.GetPaymentAccountCount(ctx, owner1)
	require.True(t, resp1.Owner == owner1.String())
	require.True(t, resp1.Count == paymentCount1.Count)

	resp2, _ := keeper.GetPaymentAccountCount(ctx, owner2)
	require.True(t, resp2.Owner == owner2.String())
	require.True(t, resp2.Count == paymentCount2.Count)

	_, found := keeper.GetPaymentAccountCount(ctx, sample.RandAccAddress())
	require.True(t, !found)

	// get all
	resp3 := keeper.GetAllPaymentAccountCount(ctx)
	require.True(t, len(resp3) == 2)
}
