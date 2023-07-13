package keeper_test

import (
	"errors"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func TestTryResumeStreamRecord_InResumingOrSettling(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// further deposit to a resuming account is not allowed
	streamRecord := &types.StreamRecord{
		Status:      types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate: sdkmath.NewInt(-100),
	}
	deposit := sdkmath.NewInt(100)
	err := keeper.TryResumeStreamRecord(ctx, streamRecord, deposit)
	require.ErrorContains(t, err, "resuming")
}

func TestTryResumeStreamRecord_ResumeInOneBlock(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// resume account in one call
	params := keeper.GetParams(ctx)
	rate := sdkmath.NewInt(100)
	user := sample.RandAccAddress()
	streamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: rate.Neg(),
		OutFlowCount:      1,
	}
	keeper.SetStreamRecord(ctx, streamRecord)

	gvg := sample.RandAccAddress()
	outFlow := &types.OutFlow{
		ToAddress: gvg.String(),
		Rate:      rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow)

	err := keeper.TryResumeStreamRecord(ctx, streamRecord, rate.MulRaw(int64(params.VersionedParams.ReserveTime)))
	require.NoError(t, err)

	userStreamRecord, _ := keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, userStreamRecord.NetflowRate, rate.Neg())
	require.Equal(t, userStreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvgStreamRecord, _ := keeper.GetStreamRecord(ctx, gvg)
	require.True(t, gvgStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvgStreamRecord.NetflowRate, rate)
	require.Equal(t, gvgStreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())
}

func TestTryResumeStreamRecord_ResumeInMultipleBlocks(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// resume account in multiple blocks
	params := keeper.GetParams(ctx)
	params.MaxAutoResumeFlowCount = 1
	_ = keeper.SetParams(ctx, params)

	rate := sdkmath.NewInt(300)
	user := sample.RandAccAddress()
	streamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: rate.Neg(),
		OutFlowCount:      3,
	}
	keeper.SetStreamRecord(ctx, streamRecord)

	gvgAddress := []sdk.AccAddress{sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress()}

	gvg1 := gvgAddress[0]
	gvg1Rate := sdk.NewInt(50)
	outFlow1 := &types.OutFlow{
		ToAddress: gvg1.String(),
		Rate:      gvg1Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow1)

	gvg2 := gvgAddress[1]
	gvg2Rate := sdk.NewInt(100)
	outFlow2 := &types.OutFlow{
		ToAddress: gvg2.String(),
		Rate:      gvg2Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow2)

	gvg3 := gvgAddress[2]
	gvg3Rate := sdk.NewInt(150)
	outFlow3 := &types.OutFlow{
		ToAddress: gvg3.String(),
		Rate:      gvg3Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow3)

	// try to resume stream record
	err := keeper.TryResumeStreamRecord(ctx, streamRecord, rate.SubRaw(10).MulRaw(int64(params.VersionedParams.ReserveTime)))
	require.NoError(t, err) //only added static balance
	found := keeper.ExistsAutoResumeRecord(ctx, ctx.BlockTime().Unix(), user)
	require.True(t, !found)
	streamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, streamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	err = keeper.TryResumeStreamRecord(ctx, streamRecord, rate.MulRaw(int64(params.VersionedParams.ReserveTime)))
	require.NoError(t, err)

	// still frozen
	userStreamRecord, _ := keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	_, found = keeper.GetAutoResumeRecord(ctx, ctx.BlockTime().Unix(), user)
	require.True(t, found)

	// resume in end block
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	// resume in end block
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	// resume in end block
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, userStreamRecord.NetflowRate, rate.Neg())
	require.Equal(t, userStreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg1StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg1)
	require.True(t, gvg1StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg1StreamRecord.NetflowRate, gvg1Rate)
	require.Equal(t, gvg1StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg2StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg2)
	require.True(t, gvg2StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg2StreamRecord.NetflowRate, gvg2Rate)
	require.Equal(t, gvg2StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg3StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg3)
	require.True(t, gvg3StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg3StreamRecord.NetflowRate, gvg3Rate)
	require.Equal(t, gvg3StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())
}

func TestTryResumeStreamRecord_ResumeInMultipleBlocks_BalanceNotEnoughFinally(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// resume account in multiple blocks
	params := keeper.GetParams(ctx)
	params.MaxAutoResumeFlowCount = 1
	_ = keeper.SetParams(ctx, params)

	rate := sdkmath.NewInt(300)
	user := sample.RandAccAddress()
	streamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: rate.Neg(),
		OutFlowCount:      3,
	}
	keeper.SetStreamRecord(ctx, streamRecord)

	gvgAddress := []sdk.AccAddress{sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress()}

	gvg1 := gvgAddress[0]
	gvg1Rate := sdk.NewInt(50)
	outFlow1 := &types.OutFlow{
		ToAddress: gvg1.String(),
		Rate:      gvg1Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow1)

	gvg2 := gvgAddress[1]
	gvg2Rate := sdk.NewInt(100)
	outFlow2 := &types.OutFlow{
		ToAddress: gvg2.String(),
		Rate:      gvg2Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow2)

	gvg3 := gvgAddress[2]
	gvg3Rate := sdk.NewInt(150)
	outFlow3 := &types.OutFlow{
		ToAddress: gvg3.String(),
		Rate:      gvg3Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow3)

	err := keeper.TryResumeStreamRecord(ctx, streamRecord, rate.MulRaw(int64(params.VersionedParams.ReserveTime)))
	require.NoError(t, err)

	// still frozen
	userStreamRecord, _ := keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	_, found := keeper.GetAutoResumeRecord(ctx, ctx.BlockTime().Unix(), user)
	require.True(t, found)

	// resume in end block
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	// resume in end block
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	// time flies
	timestamp := ctx.BlockTime().Unix() + int64(params.VersionedParams.ReserveTime)*2
	ctx = ctx.WithBlockTime(time.Unix(timestamp, 0))

	depKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()
	depKeepers.BankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("fail to transfer")).AnyTimes()

	// resume in end block
	ctx = ctx.WithValue(types.ForceUpdateStreamRecordKey, true)

	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, userStreamRecord.NetflowRate, rate.Neg())
	require.Equal(t, userStreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())
	require.True(t, userStreamRecord.StaticBalance.IsNegative()) // the static balance becomes negative

	gvg1StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg1)
	require.True(t, gvg1StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg1StreamRecord.NetflowRate, gvg1Rate)
	require.Equal(t, gvg1StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg2StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg2)
	require.True(t, gvg2StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg2StreamRecord.NetflowRate, gvg2Rate)
	require.Equal(t, gvg2StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg3StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg3)
	require.True(t, gvg3StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg3StreamRecord.NetflowRate, gvg3Rate)
	require.Equal(t, gvg3StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	// there will be an auto settle record
	autoSettles := keeper.GetAllAutoSettleRecord(ctx)
	found = false
	for _, settle := range autoSettles {
		if settle.GetAddr() == user.String() {
			found = true
		}
	}
	require.True(t, found, "")

	keeper.AutoSettle(ctx)
	keeper.AutoSettle(ctx)
	keeper.AutoSettle(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.Equal(t, userStreamRecord.NetflowRate, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.FrozenNetflowRate, rate.Neg())
}

func TestAutoSettle_AccountIsInResuming(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// resume account in multiple blocks
	params := keeper.GetParams(ctx)
	params.MaxAutoResumeFlowCount = 1
	_ = keeper.SetParams(ctx, params)

	rate := sdkmath.NewInt(300)
	user := sample.RandAccAddress()
	streamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: rate.Neg(),
		OutFlowCount:      3,
	}
	keeper.SetStreamRecord(ctx, streamRecord)

	gvgAddress := []sdk.AccAddress{sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress()}

	gvg1 := gvgAddress[0]
	gvg1Rate := sdk.NewInt(50)
	outFlow1 := &types.OutFlow{
		ToAddress: gvg1.String(),
		Rate:      gvg1Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow1)

	gvg2 := gvgAddress[1]
	gvg2Rate := sdk.NewInt(100)
	outFlow2 := &types.OutFlow{
		ToAddress: gvg2.String(),
		Rate:      gvg2Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow2)

	gvg3 := gvgAddress[2]
	gvg3Rate := sdk.NewInt(150)
	outFlow3 := &types.OutFlow{
		ToAddress: gvg3.String(),
		Rate:      gvg3Rate,
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow3)

	err := keeper.TryResumeStreamRecord(ctx, streamRecord, rate.MulRaw(int64(params.VersionedParams.ReserveTime)))
	require.NoError(t, err)

	// still frozen
	userStreamRecord, _ := keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	_, found := keeper.GetAutoResumeRecord(ctx, ctx.BlockTime().Unix(), user)
	require.True(t, found)

	// add auto settle record
	keeper.SetAutoSettleRecord(ctx, &types.AutoSettleRecord{
		Timestamp: ctx.BlockTime().Unix(),
		Addr:      user.String(),
	})

	keeper.AutoSettle(ctx)
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	keeper.AutoSettle(ctx)
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	keeper.AutoSettle(ctx)
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, userStreamRecord.NetflowRate, rate.Neg())
	require.Equal(t, userStreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg1StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg1)
	require.True(t, gvg1StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg1StreamRecord.NetflowRate, gvg1Rate)
	require.Equal(t, gvg1StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg2StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg2)
	require.True(t, gvg2StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg2StreamRecord.NetflowRate, gvg2Rate)
	require.Equal(t, gvg2StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg3StreamRecord, _ := keeper.GetStreamRecord(ctx, gvg3)
	require.True(t, gvg3StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg3StreamRecord.NetflowRate, gvg3Rate)
	require.Equal(t, gvg3StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())
}

func TestAutoSettle_SettleInOneBlock(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	depKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()
	depKeepers.BankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("fail to transfer")).AnyTimes()

	// freeze account in one block
	rate := sdkmath.NewInt(100)
	user := sample.RandAccAddress()
	userStreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       rate.Neg(),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      1,
	}
	keeper.SetStreamRecord(ctx, userStreamRecord)

	gvg := sample.RandAccAddress()
	gvgStreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       rate,
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvgStreamRecord)

	outFlow := &types.OutFlow{
		ToAddress: gvg.String(),
		Rate:      rate,
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
	}
	keeper.SetOutFlow(ctx, user, outFlow)

	keeper.SetAutoSettleRecord(ctx, &types.AutoSettleRecord{
		Timestamp: ctx.BlockTime().Unix(),
		Addr:      user.String(),
	})

	keeper.AutoSettle(ctx)

	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.Equal(t, userStreamRecord.NetflowRate, sdkmath.ZeroInt())
	require.Equal(t, userStreamRecord.FrozenNetflowRate, rate.Neg())

	gvgOutFlow := keeper.GetOutFlow(ctx, user, types.OUT_FLOW_STATUS_FROZEN, gvg)
	require.Equal(t, gvgOutFlow.Status, types.OUT_FLOW_STATUS_FROZEN)
	require.Equal(t, gvgOutFlow.Rate, rate)
}

func TestAutoSettle_SettleInMultipleBlocks(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// freeze account in multiple blocks
	params := keeper.GetParams(ctx)
	params.MaxAutoSettleFlowCount = 1
	_ = keeper.SetParams(ctx, params)

	depKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()
	depKeepers.BankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("fail to transfer")).AnyTimes()

	rate := sdkmath.NewInt(300)
	user := sample.RandAccAddress()
	userStreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       rate.Neg(),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      3,
	}
	keeper.SetStreamRecord(ctx, userStreamRecord)

	gvgAddress := []sdk.AccAddress{sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress()}

	gvg1 := gvgAddress[0]
	gvg1StreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg1.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       sdkmath.NewInt(50),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvg1StreamRecord)

	outFlow1 := &types.OutFlow{
		ToAddress: gvg1.String(),
		Rate:      sdkmath.NewInt(50),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
	}
	keeper.SetOutFlow(ctx, user, outFlow1)

	gvg2 := gvgAddress[1]
	gvg2StreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg2.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       sdkmath.NewInt(100),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvg2StreamRecord)

	outFlow2 := &types.OutFlow{
		ToAddress: gvg2.String(),
		Rate:      sdkmath.NewInt(100),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
	}
	keeper.SetOutFlow(ctx, user, outFlow2)

	gvg3 := gvgAddress[2]
	gvg3StreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg3.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       sdkmath.NewInt(150),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvg3StreamRecord)

	outFlow3 := &types.OutFlow{
		ToAddress: gvg3.String(),
		Rate:      sdkmath.NewInt(150),
		Status:    types.OUT_FLOW_STATUS_ACTIVE,
	}
	keeper.SetOutFlow(ctx, user, outFlow3)

	keeper.SetAutoSettleRecord(ctx, &types.AutoSettleRecord{
		Timestamp: ctx.BlockTime().Unix(),
		Addr:      user.String(),
	})

	keeper.AutoSettle(ctx) // this is for settle stream, it is counted
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	keeper.AutoSettle(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.True(t, !userStreamRecord.NetflowRate.IsZero())
	require.True(t, !userStreamRecord.FrozenNetflowRate.IsZero())

	keeper.AutoSettle(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.True(t, !userStreamRecord.NetflowRate.IsZero())
	require.True(t, !userStreamRecord.FrozenNetflowRate.IsZero())

	keeper.AutoSettle(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.True(t, userStreamRecord.NetflowRate.IsZero())
	require.True(t, userStreamRecord.FrozenNetflowRate.Equal(rate.Neg()))

	gvg1OutFlow := keeper.GetOutFlow(ctx, user, types.OUT_FLOW_STATUS_FROZEN, gvg1)
	require.Equal(t, gvg1OutFlow.Status, types.OUT_FLOW_STATUS_FROZEN)
	require.Equal(t, gvg1OutFlow.Rate, sdkmath.NewInt(50))

	gvg2OutFlow := keeper.GetOutFlow(ctx, user, types.OUT_FLOW_STATUS_FROZEN, gvg2)
	require.Equal(t, gvg2OutFlow.Status, types.OUT_FLOW_STATUS_FROZEN)
	require.Equal(t, gvg2OutFlow.Rate, sdkmath.NewInt(100))

	gvg3OutFlow := keeper.GetOutFlow(ctx, user, types.OUT_FLOW_STATUS_FROZEN, gvg3)
	require.Equal(t, gvg3OutFlow.Status, types.OUT_FLOW_STATUS_FROZEN)
	require.Equal(t, gvg3OutFlow.Rate, sdkmath.NewInt(150))
}

func TestAutoSettle_SettleInMultipleBlocks_AutoResumeExists(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	// freeze account in multiple blocks
	params := keeper.GetParams(ctx)
	params.MaxAutoSettleFlowCount = 1
	params.MaxAutoResumeFlowCount = 1
	_ = keeper.SetParams(ctx, params)

	depKeepers.AccountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()
	depKeepers.BankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("fail to transfer")).AnyTimes()

	rate := sdkmath.NewInt(300)
	user := sample.RandAccAddress()
	userStreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate:       sdk.ZeroInt(),
		FrozenNetflowRate: rate.Neg(),
		OutFlowCount:      3,
	}
	keeper.SetStreamRecord(ctx, userStreamRecord)

	gvgAddress := []sdk.AccAddress{sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress()}

	gvg1 := gvgAddress[0]
	gvg1StreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg1.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvg1StreamRecord)

	outFlow1 := &types.OutFlow{
		ToAddress: gvg1.String(),
		Rate:      sdkmath.NewInt(50),
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow1)

	gvg2 := gvgAddress[1]
	gvg2StreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg2.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvg2StreamRecord)

	outFlow2 := &types.OutFlow{
		ToAddress: gvg2.String(),
		Rate:      sdkmath.NewInt(100),
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow2)

	gvg3 := gvgAddress[2]
	gvg3StreamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		Account:           gvg3.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_ACTIVE,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: sdkmath.NewInt(0),
		OutFlowCount:      0,
	}
	keeper.SetStreamRecord(ctx, gvg3StreamRecord)

	outFlow3 := &types.OutFlow{
		ToAddress: gvg3.String(),
		Rate:      sdkmath.NewInt(150),
		Status:    types.OUT_FLOW_STATUS_FROZEN,
	}
	keeper.SetOutFlow(ctx, user, outFlow3)

	// resume the stream record
	err := keeper.TryResumeStreamRecord(ctx, userStreamRecord, rate.MulRaw(int64(params.VersionedParams.ReserveTime)))
	require.NoError(t, err) //only added static balance
	found := keeper.ExistsAutoResumeRecord(ctx, ctx.BlockTime().Unix(), user)
	require.True(t, found)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	// add auto settle record
	settleTime := ctx.BlockTime().Unix()
	keeper.SetAutoSettleRecord(ctx, &types.AutoSettleRecord{
		Timestamp: settleTime,
		Addr:      user.String(),
	})

	keeper.AutoSettle(ctx) // this is for settle stream, it is counted
	keeper.AutoSettle(ctx)
	keeper.AutoSettle(ctx)
	keeper.AutoSettle(ctx)
	keeper.AutoSettle(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)

	keeper.AutoResume(ctx)
	keeper.AutoResume(ctx)
	keeper.AutoResume(ctx)
	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)

	timestamp := ctx.BlockTime().Unix()
	ctx = ctx.WithBlockTime(time.Unix(timestamp+10, 0))
	keeper.AutoSettle(ctx) // it will pick up the auto settle record
	autoSettles := keeper.GetAllAutoSettleRecord(ctx)
	var record types.AutoSettleRecord
	for _, settle := range autoSettles {
		if settle.GetAddr() == user.String() {
			record = settle
		}
	}
	// old settle record removed, new settle record added
	require.True(t, record.Timestamp != settleTime, "")

	userStreamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.True(t, userStreamRecord.NetflowRate.Equal(rate.Neg()))
	require.True(t, userStreamRecord.FrozenNetflowRate.IsZero())

	gvg1StreamRecord, _ = keeper.GetStreamRecord(ctx, gvg1)
	require.True(t, gvg1StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg1StreamRecord.NetflowRate, sdk.NewInt(50))
	require.Equal(t, gvg1StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg2StreamRecord, _ = keeper.GetStreamRecord(ctx, gvg2)
	require.True(t, gvg2StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg2StreamRecord.NetflowRate, sdk.NewInt(100))
	require.Equal(t, gvg2StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())

	gvg3StreamRecord, _ = keeper.GetStreamRecord(ctx, gvg3)
	require.True(t, gvg3StreamRecord.Status == types.STREAM_ACCOUNT_STATUS_ACTIVE)
	require.Equal(t, gvg3StreamRecord.NetflowRate, sdk.NewInt(150))
	require.Equal(t, gvg3StreamRecord.FrozenNetflowRate, sdkmath.ZeroInt())
}
