package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestMergeActiveOutFlows(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)

	addr := sample.RandAccAddress()

	toAddr1 := sample.RandAccAddress()
	toAddr1Rate := math.NewInt(10)
	outFlow1 := types.OutFlow{
		ToAddress: toAddr1.String(),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
		Rate:      toAddr1Rate,
	}
	toAddr2 := sample.RandAccAddress()
	toAddr2Rate := math.NewInt(10)
	outFlow2 := types.OutFlow{
		ToAddress: toAddr2.String(),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
		Rate:      toAddr2Rate,
	}

	// add new flows
	count := keeper.MergeActiveOutFlows(ctx, addr, []types.OutFlow{outFlow1, outFlow2})
	require.Equal(t, 2, count)

	outFlows := keeper.GetOutFlows(ctx, addr)
	require.Equal(t, 2, len(outFlows))

	// update existing flows
	count = keeper.MergeActiveOutFlows(ctx, addr, []types.OutFlow{outFlow1, outFlow2})
	require.Equal(t, 0, count)

	outFlows = keeper.GetOutFlows(ctx, addr)
	require.Equal(t, 2, len(outFlows))

	// add new flow
	outFlow3 := types.OutFlow{
		ToAddress: sample.RandAccAddress().String(),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
		Rate:      toAddr2Rate,
	}

	count = keeper.MergeActiveOutFlows(ctx, addr, []types.OutFlow{outFlow3})
	require.Equal(t, 1, count)

	outFlows = keeper.GetOutFlows(ctx, addr)
	require.Equal(t, 3, len(outFlows))

	// delete a flow
	outFlow4 := types.OutFlow{
		ToAddress: toAddr1.String(),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
		Rate:      toAddr1Rate.MulRaw(2).Neg(),
	}

	count = keeper.MergeActiveOutFlows(ctx, addr, []types.OutFlow{outFlow4})
	require.Equal(t, -1, count)

	outFlows = keeper.GetOutFlows(ctx, addr)
	require.Equal(t, 2, len(outFlows))
}
