package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestApplyFlowChanges(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	user := "user"
	rate := sdkmath.NewInt(100)
	sp := "sp"
	userInitBalance := sdkmath.NewInt(1e10)
	flowChanges := []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr(user).WithStaticBalanceChange(userInitBalance).WithRateChange(rate.Neg()),
		*types.NewDefaultStreamRecordChangeWithAddr(sp).WithRateChange(rate),
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
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithRateChange(rate).WithStaticBalanceChange(staticBalance)
	_, err := keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	// check
	streamRecord, found := keeper.GetStreamRecord(ctx, user)
	require.True(t, found)
	t.Logf("stream record: %+v", streamRecord)
	// 345 seconds pass
	var seconds int64 = 345
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(seconds) * time.Second))
	change = types.NewDefaultStreamRecordChangeWithAddr(user)
	_, err = keeper.UpdateStreamRecordByAddr(ctx, change)
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
		*types.NewDefaultStreamRecordChangeWithAddr("user1").WithRateChange(sdkmath.NewInt(100)).WithStaticBalanceChange(sdkmath.NewInt(1e10)),
		*types.NewDefaultStreamRecordChangeWithAddr("user2").WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
	}
	changes := []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr("user1").WithRateChange(sdkmath.NewInt(100)).WithStaticBalanceChange(sdkmath.NewInt(1e10)),
		*types.NewDefaultStreamRecordChangeWithAddr("user3").WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
	}
	k, _ := keepertest.PaymentKeeper(t)
	k.MergeStreamRecordChanges(&base, changes)
	t.Logf("new base: %+v", base)
	require.Equal(t, len(base), 3)
	require.Equal(t, base, []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr("user1").WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
		*types.NewDefaultStreamRecordChangeWithAddr("user2").WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
		*types.NewDefaultStreamRecordChangeWithAddr("user3").WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
	})
}

func TestAutoForceSettle(t *testing.T) {
	keeper, ctx := keepertest.PaymentKeeper(t)
	params := keeper.GetParams(ctx)
	var startTime int64 = 100
	ctx = ctx.WithBlockTime(time.Unix(startTime, 0))
	user := keepertest.GetRandomAddress()
	rate := sdkmath.NewInt(100)
	sp := keepertest.GetRandomAddress()
	userInitBalance := sdkmath.NewInt(int64(100*params.ReserveTime) + 1) // just enough for reserve
	// init balance
	streamRecordChanges := []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr(user).WithStaticBalanceChange(userInitBalance),
	}
	err := keeper.ApplyStreamRecordChanges(ctx, streamRecordChanges)
	require.NoError(t, err)
	userStreamRecord, found := keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	flowChanges := []types.OutFlow{
		{ToAddress: sp, Rate: rate},
	}
	err = keeper.ApplyFlowChanges(ctx, user, flowChanges)
	require.NoError(t, err)
	userStreamRecord, found = keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	require.Equal(t, 1, len(userStreamRecord.OutFlows))
	require.Equal(t, userStreamRecord.OutFlows[0].ToAddress, sp)
	spStreamRecord, found := keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	require.True(t, found)
	require.Equal(t, spStreamRecord.NetflowRate, rate)
	require.Equal(t, spStreamRecord.StaticBalance, sdkmath.ZeroInt())
	require.Equal(t, spStreamRecord.BufferBalance, sdkmath.ZeroInt())
	// check auto settle queue
	autoSettleQueue := keeper.GetAllAutoSettleRecord(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue)
	require.Equal(t, len(autoSettleQueue), 1)
	require.Equal(t, autoSettleQueue[0].Addr, user)
	require.Equal(t, autoSettleQueue[0].Timestamp, startTime+int64(params.ReserveTime)-int64(params.ForcedSettleTime))
	// 1 day pass
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(86400) * time.Second))
	// update and deposit to user for extra 100s
	userAddBalance := rate.MulRaw(100)
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithStaticBalanceChange(userAddBalance)
	ret, err := keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	userStreamRecord = *ret
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	require.True(t, userStreamRecord.StaticBalance.IsNegative())
	change = types.NewDefaultStreamRecordChangeWithAddr(sp)
	_, err = keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	spStreamRecord, _ = keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	autoSettleQueue2 := keeper.GetAllAutoSettleRecord(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue2)
	require.Equal(t, autoSettleQueue[0].Timestamp+100, autoSettleQueue2[0].Timestamp)
	// reverve time - forced settle time - 1 day + 101s pass
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(params.ReserveTime-params.ForcedSettleTime-86400+101) * time.Second))
	change = types.NewDefaultStreamRecordChangeWithAddr(user)
	_, err = keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	// user has been force settled
	require.Equal(t, userStreamRecord.StaticBalance, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.BufferBalance, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.NetflowRate, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.Status, int32(types.StreamPaymentAccountStatusFrozen))
	change = types.NewDefaultStreamRecordChangeWithAddr(sp)
	_, err = keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	spStreamRecord, _ = keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	autoSettleQueue3 := keeper.GetAllAutoSettleRecord(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue3)
	require.Equal(t, len(autoSettleQueue3), 0)
	govStreamRecord, found := keeper.GetStreamRecord(ctx, types.GovernanceAddress.String())
	require.True(t, found)
	t.Logf("gov stream record: %+v", govStreamRecord)
	require.Equal(t, govStreamRecord.StaticBalance.Add(spStreamRecord.StaticBalance), userInitBalance.Add(userAddBalance))
}
