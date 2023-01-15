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
	err = keeper.UpdateStreamRecordByAddr(ctx, user, sdkmath.ZeroInt(), sdkmath.ZeroInt(), false)
	require.NoError(t, err)
	userStreamRecord2, _ := keeper.GetStreamRecord(ctx, user)
	t.Logf("stream record after %d seconds: %+v", seconds, userStreamRecord2)
	require.Equal(t, userStreamRecord2.StaticBalance, streamRecord.StaticBalance.Add(rate.Mul(sdkmath.NewInt(seconds))))
	require.Equal(t, userStreamRecord2.BufferBalance, streamRecord.BufferBalance)
	require.Equal(t, userStreamRecord2.NetflowRate, streamRecord.NetflowRate)
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

func TestAutoForceSettle(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	params := keeper.GetParams(ctx)
	var startTime int64 = 100
	ctx = ctx.WithBlockTime(time.Unix(startTime, 0))
	user := "user"
	rate := sdkmath.NewInt(100)
	sp := "sp"
	userInitBalance := sdkmath.NewInt(int64(100*params.ReserveTime) + 1) // just enough for reserve
	// init balance
	streamRecordChanges := []types.StreamRecordChange{
		{user, sdkmath.ZeroInt(), userInitBalance},
	}
	err := keeper.ApplyStreamRecordChanges(ctx, streamRecordChanges)
	require.NoError(t, err)
	userStreamRecord, found := keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	flowChanges := []types.Flow{
		{From: user, To: sp, Rate: rate},
	}
	err = keeper.ApplyUSDFlowChanges(ctx, flowChanges)
	userStreamRecord, found = keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	spStreamRecord, found := keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	require.True(t, found)
	require.Equal(t, spStreamRecord.NetflowRate, rate)
	require.Equal(t, spStreamRecord.StaticBalance, sdkmath.ZeroInt())
	require.Equal(t, spStreamRecord.BufferBalance, sdkmath.ZeroInt())
	// check flows
	flows := keeper.GetAllFlow(ctx)
	t.Logf("flows: %+v", flows)
	require.Equal(t, len(flows), 1)
	require.Equal(t, flows[0].From, user)
	require.Equal(t, flows[0].To, sp)
	require.Equal(t, flows[0].Rate, rate)
	// check auto settle queue
	autoSettleQueue := keeper.GetAllAutoSettleQueue(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue)
	require.Equal(t, len(autoSettleQueue), 1)
	require.Equal(t, autoSettleQueue[0].Addr, user)
	require.Equal(t, autoSettleQueue[0].Timestamp, startTime+int64(params.ReserveTime)-int64(params.ForcedSettleTime))
	// 1 day pass
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(86400) * time.Second))
	// update and deposit to user for extra 100s
	userAddBalance := rate.MulRaw(100)
	err = keeper.UpdateStreamRecordByAddr(ctx, user, sdkmath.ZeroInt(), userAddBalance, false)
	require.NoError(t, err)
	userStreamRecord, found = keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	require.True(t, userStreamRecord.StaticBalance.IsNegative())
	err = keeper.UpdateStreamRecordByAddr(ctx, sp, sdkmath.ZeroInt(), sdkmath.ZeroInt(), false)
	require.NoError(t, err)
	spStreamRecord, found = keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	autoSettleQueue2 := keeper.GetAllAutoSettleQueue(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue2)
	require.Equal(t, autoSettleQueue[0].Timestamp+100, autoSettleQueue2[0].Timestamp)
	// reverve time - forced settle time - 1 day + 101s pass
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(params.ReserveTime-params.ForcedSettleTime-86400+101) * time.Second))
	err = keeper.UpdateStreamRecordByAddr(ctx, user, sdkmath.ZeroInt(), sdkmath.ZeroInt(), false)
	require.NoError(t, err)
	userStreamRecord, found = keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	// user has been force settled
	require.Equal(t, userStreamRecord.StaticBalance, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.BufferBalance, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.NetflowRate, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.Status, int32(types.StreamPaymentAccountStatusFrozen))
	err = keeper.UpdateStreamRecordByAddr(ctx, sp, sdkmath.ZeroInt(), sdkmath.ZeroInt(), false)
	require.NoError(t, err)
	spStreamRecord, found = keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	autoSettleQueue3 := keeper.GetAllAutoSettleQueue(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue3)
	require.Equal(t, len(autoSettleQueue3), 0)
	flows = keeper.GetAllFlow(ctx)
	t.Logf("flows: %+v", flows)
	require.True(t, flows[0].Frozen)
	govStreamRecord, found := keeper.GetStreamRecord(ctx, types.GovernanceAddress.String())
	require.True(t, found)
	t.Logf("gov stream record: %+v", govStreamRecord)
	require.Equal(t, govStreamRecord.StaticBalance.Add(spStreamRecord.StaticBalance), userInitBalance.Add(userAddBalance))
}
