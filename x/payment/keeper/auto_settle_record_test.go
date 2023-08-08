package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestAutoSettleRecord(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)

	addr1 := sample.RandAccAddress()
	record1 := &types.AutoSettleRecord{
		Addr:      addr1.String(),
		Timestamp: 100,
	}

	addr2 := sample.RandAccAddress()
	record2 := &types.AutoSettleRecord{
		Addr:      addr2.String(),
		Timestamp: 200,
	}

	// set
	keeper.SetAutoSettleRecord(ctx, record1)
	keeper.SetAutoSettleRecord(ctx, record2)

	// update to new time
	keeper.UpdateAutoSettleRecord(ctx, addr1, record1.Timestamp, 110)

	// update to remove
	keeper.UpdateAutoSettleRecord(ctx, addr2, record2.Timestamp, 0)

	// get all
	records := keeper.GetAllAutoSettleRecord(ctx)
	require.True(t, len(records) == 1)
	require.True(t, records[0].Addr == addr1.String())
	require.True(t, records[0].Timestamp == 110)
}
