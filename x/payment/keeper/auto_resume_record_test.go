package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestAutoResumeRecord(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)

	addr1 := sample.RandAccAddress()
	record1 := &types.AutoResumeRecord{
		Addr:      addr1.String(),
		Timestamp: 100,
	}

	addr2 := sample.RandAccAddress()
	record2 := &types.AutoResumeRecord{
		Addr:      addr2.String(),
		Timestamp: 200,
	}

	// set
	keeper.SetAutoResumeRecord(ctx, record1)
	keeper.SetAutoResumeRecord(ctx, record2)

	// exits
	// before the timestamp
	exist := keeper.ExistsAutoResumeRecord(ctx, 90, addr1)
	require.True(t, !exist)
	exist = keeper.ExistsAutoResumeRecord(ctx, 101, addr1)
	require.True(t, exist)

	// at any time
	exist = keeper.ExistsAutoResumeRecord(ctx, 0, addr1)
	require.True(t, exist)
	exist = keeper.ExistsAutoResumeRecord(ctx, 0, addr2)
	require.True(t, exist)

	// remove
	keeper.RemoveAutoResumeRecord(ctx, record1.Timestamp, addr1)
	keeper.RemoveAutoResumeRecord(ctx, record2.Timestamp, addr2)

	exist = keeper.ExistsAutoResumeRecord(ctx, 0, addr1)
	require.True(t, !exist)
	exist = keeper.ExistsAutoResumeRecord(ctx, 0, addr2)
	require.True(t, !exist)
}
