package keeper_test

import (
	"sort"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestApplyFlowChanges(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	user := sample.RandAccAddress()
	rate := sdkmath.NewInt(100)
	sp := sample.RandAccAddress()
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
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	user := sample.RandAccAddress()
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
	users := []sdk.AccAddress{
		sample.RandAccAddress(),
		sample.RandAccAddress(),
		sample.RandAccAddress(),
	}
	sort.Slice(users, func(i, j int) bool {
		return users[i].String() < users[j].String()
	})
	user1 := users[0]
	user2 := users[1]
	user3 := users[2]
	base := []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr(user1).WithRateChange(sdkmath.NewInt(100)).WithStaticBalanceChange(sdkmath.NewInt(1e10)),
		*types.NewDefaultStreamRecordChangeWithAddr(user2).WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
	}
	changes := []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr(user1).WithRateChange(sdkmath.NewInt(100)).WithStaticBalanceChange(sdkmath.NewInt(1e10)),
		*types.NewDefaultStreamRecordChangeWithAddr(user3).WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
	}
	k, _, _ := makePaymentKeeper(t)
	merged := k.MergeStreamRecordChanges(append(base, changes...))
	t.Logf("merged: %+v", merged)
	require.Equal(t, len(merged), 3)
	require.Equal(t, merged, []types.StreamRecordChange{
		*types.NewDefaultStreamRecordChangeWithAddr(user1).WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
		*types.NewDefaultStreamRecordChangeWithAddr(user2).WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
		*types.NewDefaultStreamRecordChangeWithAddr(user3).WithRateChange(sdkmath.NewInt(200)).WithStaticBalanceChange(sdkmath.NewInt(2e10)),
	})
}

func TestAutoForceSettle(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)
	t.Logf("depKeepers: %+v", depKeepers)
	params := keeper.GetParams(ctx)
	var startTime int64 = 100
	ctx = ctx.WithBlockTime(time.Unix(startTime, 0))
	user := sample.RandAccAddress()
	rate := sdkmath.NewInt(100)
	sp := sample.RandAccAddress()
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
		{ToAddress: sp.String(), Rate: rate},
	}
	userFlows := types.UserFlows{Flows: flowChanges, From: user}
	err = keeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	require.NoError(t, err)
	userStreamRecord, found = keeper.GetStreamRecord(ctx, user)
	t.Logf("user stream record: %+v", userStreamRecord)
	require.True(t, found)
	require.Equal(t, 1, len(userStreamRecord.OutFlows))
	require.Equal(t, userStreamRecord.OutFlows[0].ToAddress, sp.String())
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
	require.Equal(t, autoSettleQueue[0].Addr, user.String())
	require.Equal(t, autoSettleQueue[0].Timestamp, startTime+int64(params.ReserveTime)-int64(params.ForcedSettleTime))
	// 1 day pass
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(86400) * time.Second))
	// update and deposit to user for extra 100s
	depKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).Return(false).AnyTimes()
	userAddBalance := rate.MulRaw(100)
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithStaticBalanceChange(userAddBalance)
	ret, err := keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	userStreamRecord = ret
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
	usrBeforeForceSettle, _ := keeper.GetStreamRecord(ctx, user)
	t.Logf("usrBeforeForceSettle: %s", usrBeforeForceSettle)
	usr, _ := keeper.GetStreamRecord(ctx, user)
	err = keeper.UpdateStreamRecord(ctx, usr, change, true)
	require.NoError(t, err)
	keeper.SetStreamRecord(ctx, usr)
	usrAfterForceSettle, found := keeper.GetStreamRecord(ctx, user)
	require.True(t, found)
	t.Logf("usrAfterForceSettle: %s", usrAfterForceSettle)
	// user has been force settled
	require.Equal(t, usrAfterForceSettle.StaticBalance, sdkmath.ZeroInt())
	require.Equal(t, usrAfterForceSettle.BufferBalance, sdkmath.ZeroInt())
	require.Equal(t, usrAfterForceSettle.NetflowRate, sdkmath.ZeroInt())
	require.Equal(t, usrAfterForceSettle.Status, types.STREAM_ACCOUNT_STATUS_FROZEN)
	change = types.NewDefaultStreamRecordChangeWithAddr(sp)
	_, err = keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)
	spStreamRecord, _ = keeper.GetStreamRecord(ctx, sp)
	t.Logf("sp stream record: %+v", spStreamRecord)
	autoSettleQueue3 := keeper.GetAllAutoSettleRecord(ctx)
	t.Logf("auto settle queue: %+v", autoSettleQueue3)
	require.Equal(t, len(autoSettleQueue3), 0)
	govStreamRecord, found := keeper.GetStreamRecord(ctx, types.GovernanceAddress)
	require.True(t, found)
	t.Logf("gov stream record: %+v", govStreamRecord)
	require.Equal(t, govStreamRecord.StaticBalance.Add(spStreamRecord.StaticBalance), userInitBalance.Add(userAddBalance))
}
