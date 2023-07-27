package keeper_test

import (
	"errors"
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
	sr := &types.StreamRecord{Account: user.String(),
		OutFlowCount:      1,
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		NetflowRate:       sdkmath.ZeroInt(),
		FrozenNetflowRate: sdkmath.ZeroInt(),
	}
	keeper.SetStreamRecord(ctx, sr)
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

func TestApplyUserFlows_ActiveStreamRecord(t *testing.T) {
	keeper, ctx, deepKeepers := makePaymentKeeper(t)
	ctx = ctx.WithIsCheckTx(true)

	from := sample.RandAccAddress()
	userFlows := types.UserFlows{
		From: from,
	}

	toAddr1 := sample.RandAccAddress()
	outFlow1 := types.OutFlow{
		ToAddress: toAddr1.String(),
		Rate:      sdkmath.NewInt(100),
	}
	userFlows.Flows = append(userFlows.Flows, outFlow1)

	toAddr2 := sample.RandAccAddress()
	outFlow2 := types.OutFlow{
		ToAddress: toAddr2.String(),
		Rate:      sdkmath.NewInt(200),
	}
	userFlows.Flows = append(userFlows.Flows, outFlow2)

	// no bank account
	deepKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(false).Times(1)
	err := keeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	require.ErrorContains(t, err, "balance not enough")

	// has bank account, but balance is not enough
	deepKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()
	deepKeepers.BankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("transfer error")).Times(1)
	err = keeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	require.ErrorContains(t, err, "balance not enough")

	// has bank account, and balance is enough
	deepKeepers.BankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	err = keeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	require.NoError(t, err)

	fromRecord, _ := keeper.GetStreamRecord(ctx, from)
	require.True(t, fromRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.True(t, fromRecord.NetflowRate.Int64() == -300)
	require.True(t, fromRecord.StaticBalance.Int64() == 0)
	require.True(t, fromRecord.FrozenNetflowRate.Int64() == 0)
	require.True(t, fromRecord.LockBalance.Int64() == 0)
	require.True(t, fromRecord.BufferBalance.Int64() > 0)

	to1Record, _ := keeper.GetStreamRecord(ctx, toAddr1)
	require.True(t, to1Record.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.True(t, to1Record.NetflowRate.Int64() == 100)
	require.True(t, to1Record.StaticBalance.Int64() == 0)
	require.True(t, to1Record.FrozenNetflowRate.Int64() == 0)
	require.True(t, to1Record.LockBalance.Int64() == 0)
	require.True(t, to1Record.BufferBalance.Int64() == 0)

	to2Record, _ := keeper.GetStreamRecord(ctx, toAddr2)
	require.True(t, to2Record.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.True(t, to2Record.NetflowRate.Int64() == 200)
	require.True(t, to2Record.StaticBalance.Int64() == 0)
	require.True(t, to2Record.FrozenNetflowRate.Int64() == 0)
	require.True(t, to2Record.LockBalance.Int64() == 0)
	require.True(t, to2Record.BufferBalance.Int64() == 0)
}

func TestApplyUserFlows_Frozen(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)

	from := sample.RandAccAddress()
	toAddr1 := sample.RandAccAddress()
	toAddr2 := sample.RandAccAddress()

	// the account is frozen, and during auto settle or auto resume
	fromStreamRecord := types.NewStreamRecord(from, ctx.BlockTime().Unix())
	fromStreamRecord.Status = types.STREAM_ACCOUNT_STATUS_FROZEN
	fromStreamRecord.NetflowRate = sdkmath.NewInt(-100)
	fromStreamRecord.FrozenNetflowRate = sdkmath.NewInt(-200)
	fromStreamRecord.StaticBalance = sdkmath.ZeroInt()
	fromStreamRecord.OutFlowCount = 4
	keeper.SetStreamRecord(ctx, fromStreamRecord)

	keeper.SetOutFlow(ctx, from, &types.OutFlow{
		ToAddress: toAddr1.String(),
		Rate:      sdkmath.NewInt(40),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
	})
	keeper.SetOutFlow(ctx, from, &types.OutFlow{
		ToAddress: sample.RandAccAddress().String(),
		Rate:      sdkmath.NewInt(60),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
	})
	keeper.SetOutFlow(ctx, from, &types.OutFlow{
		ToAddress: toAddr2.String(),
		Rate:      sdkmath.NewInt(120),
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	})
	keeper.SetOutFlow(ctx, from, &types.OutFlow{
		ToAddress: sample.RandAccAddress().String(),
		Rate:      sdkmath.NewInt(80),
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	})

	to1StreamRecord := types.NewStreamRecord(toAddr1, ctx.BlockTime().Unix())
	to1StreamRecord.NetflowRate = sdkmath.NewInt(300)
	to1StreamRecord.StaticBalance = sdkmath.NewInt(300)
	keeper.SetStreamRecord(ctx, to1StreamRecord)

	to2StreamRecord := types.NewStreamRecord(toAddr2, ctx.BlockTime().Unix())
	to2StreamRecord.NetflowRate = sdkmath.NewInt(400)
	to2StreamRecord.StaticBalance = sdkmath.NewInt(400)
	keeper.SetStreamRecord(ctx, to2StreamRecord)

	userFlows := types.UserFlows{
		From: from,
	}

	outFlow1 := types.OutFlow{
		ToAddress: toAddr1.String(),
		Rate:      sdkmath.NewInt(-40),
	}
	userFlows.Flows = append(userFlows.Flows, outFlow1)

	outFlow2 := types.OutFlow{
		ToAddress: toAddr2.String(),
		Rate:      sdkmath.NewInt(-60),
	}
	userFlows.Flows = append(userFlows.Flows, outFlow2)

	// update frozen stream record needs force flag
	err := keeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	require.ErrorContains(t, err, "frozen")

	ctx = ctx.WithValue(types.ForceUpdateStreamRecordKey, true)
	err = keeper.ApplyUserFlowsList(ctx, []types.UserFlows{userFlows})
	require.NoError(t, err)

	fromRecord, _ := keeper.GetStreamRecord(ctx, from)
	require.True(t, fromRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.True(t, fromRecord.StaticBalance.Int64() == 0)
	require.True(t, fromRecord.NetflowRate.Int64() == -60)
	require.True(t, fromRecord.FrozenNetflowRate.Int64() == -140)
	require.True(t, fromRecord.LockBalance.Int64() == 0)
	require.True(t, fromRecord.BufferBalance.Int64() == 0)

	outFlows := keeper.GetOutFlows(ctx, from)
	require.True(t, len(outFlows) == 3)
	// the out flow to toAddr1 should be deleted
	// the out flow to toAddr2 should be still there
	to1Found := false
	for _, outFlow := range outFlows {
		if outFlow.ToAddress == toAddr1.String() {
			to1Found = true
		}
		if outFlow.ToAddress == toAddr2.String() {
			require.True(t, outFlow.Rate.Int64() == 60)
			require.True(t, outFlow.Status == types.OUT_FLOW_STATUS_FROZEN)
		}
	}
	require.True(t, !to1Found)

	to1Record, _ := keeper.GetStreamRecord(ctx, toAddr1)
	require.True(t, to1Record.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.True(t, to1Record.NetflowRate.Int64() == 260)
	require.True(t, to1Record.FrozenNetflowRate.Int64() == 0)
	require.True(t, to1Record.LockBalance.Int64() == 0)
	require.True(t, to1Record.BufferBalance.Int64() == 0)

	to2Record, _ := keeper.GetStreamRecord(ctx, toAddr2)
	require.True(t, to2Record.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.True(t, to2Record.NetflowRate.Int64() == 400) // the outflow is frozen, which means the flow had been deduced
	require.True(t, to2Record.FrozenNetflowRate.Int64() == 0)
	require.True(t, to2Record.LockBalance.Int64() == 0)
	require.True(t, to2Record.BufferBalance.Int64() == 0)
}
