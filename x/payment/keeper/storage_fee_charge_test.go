package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestApplyFlowChanges(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	user := "user"
	rate := sdkmath.NewInt(100)
	sp := "sp"
	userInitBalance := sdkmath.NewInt(1e10)
	flowChanges := []types.StreamRecordChange{
		{user, rate.Neg(), userInitBalance},
		{sp, rate, sdkmath.NewInt(0)},
	}
	err := keeper.ApplyStreamRecordChanges(ctx, flowChanges)
	require.NoError(t, err)
	userStreamRecord, found := keeper.GetStreamRecord(ctx, user)
	require.True(t, found)
	require.Equal(t, userStreamRecord.StaticBalance.Add(userStreamRecord.BufferBalance), userInitBalance)
	require.Equal(t, userStreamRecord.NetflowRate, rate.Neg())
	t.Logf("user stream record: %+v", userStreamRecord)
	spStreamRecord, found := keeper.GetStreamRecord(ctx, sp)
	require.Equal(t, spStreamRecord.NetflowRate, rate)
	require.Equal(t, spStreamRecord.StaticBalance, sdkmath.ZeroInt())
	require.Equal(t, spStreamRecord.BufferBalance, sdkmath.ZeroInt())
	require.True(t, found)
	t.Logf("sp stream record: %+v", spStreamRecord)
}

func TestSettleStreamRecord(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	user := "user"
	rate := sdkmath.NewInt(-100)
	staticBalance := sdkmath.NewInt(1e10)
	err := keeper.UpdateStreamRecordByAddr(ctx, user, rate, staticBalance, false)
	require.NoError(t, err)
	// check
	streamRecord, found := keeper.GetStreamRecord(ctx, user)
	require.True(t, found)
	t.Logf("stream record: %+v", streamRecord)
	// 345 seconds pass
	var seconds int64 = 345
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(seconds) * time.Second))
	err = keeper.SettleStreamRecord(ctx, user)
	require.NoError(t, err)
	userStreamRecord2, _ := keeper.GetStreamRecord(ctx, user)
	t.Logf("stream record after %d seconds: %+v", seconds, userStreamRecord2)
	require.Equal(t, userStreamRecord2.StaticBalance, streamRecord.StaticBalance.Add(rate.Mul(sdkmath.NewInt(seconds))))
	require.Equal(t, userStreamRecord2.BufferBalance, streamRecord.BufferBalance)
	require.Equal(t, userStreamRecord2.NetflowRate, streamRecord.NetflowRate)
	require.Equal(t, userStreamRecord2.FrozenNetflowRate, streamRecord.FrozenNetflowRate)
	require.Equal(t, userStreamRecord2.CrudTimestamp, streamRecord.CrudTimestamp+seconds)
}

func TestMergeStreamRecordChanges(t *testing.T) {
	base := []types.StreamRecordChange{
		{"user1", sdkmath.NewInt(100), sdkmath.NewInt(1e10)},
		{"user2", sdkmath.NewInt(200), sdkmath.NewInt(2e10)},
	}
	changes := []types.StreamRecordChange{
		{"user1", sdkmath.NewInt(100), sdkmath.NewInt(1e10)},
		{"user3", sdkmath.NewInt(200), sdkmath.NewInt(2e10)},
	}
	k, _ := keepertest.PaymentKeeper(t)
	k.MergeStreamRecordChanges(&base, changes)
	t.Logf("new base: %+v", base)
	require.Equal(t, len(base), 3)
	require.Equal(t, base, []types.StreamRecordChange{
		{"user1", sdkmath.NewInt(200), sdkmath.NewInt(2e10)},
		{"user2", sdkmath.NewInt(200), sdkmath.NewInt(2e10)},
		{"user3", sdkmath.NewInt(200), sdkmath.NewInt(2e10)},
	})
}
