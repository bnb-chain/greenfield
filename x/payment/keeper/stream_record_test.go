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

	account := sample.RandAccAddress()
	// deposit to a resuming account is not allowed
	streamRecord := &types.StreamRecord{
		Account:     account.String(),
		Status:      types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate: sdkmath.NewInt(-100),
	}

	keeper.SetAutoResumeRecord(ctx, &types.AutoResumeRecord{
		Timestamp: ctx.BlockTime().Unix() + 10,
		Addr:      account.String(),
	})

	deposit := sdkmath.NewInt(100)
	err := keeper.TryResumeStreamRecord(ctx, streamRecord, deposit)
	require.ErrorContains(t, err, "is resuming")
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
	require.True(t, userStreamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
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

	ctx = ctx.WithValue(types.ForceUpdateStreamRecordKey, true)
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

	ctx = ctx.WithValue(types.ForceUpdateStreamRecordKey, true)
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

func TestUpdateStreamRecord_FrozenAccountLockBalance(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Now())

	user := sample.RandAccAddress()
	streamRecord := &types.StreamRecord{
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.NewInt(1000),
		Account:           user.String(),
		Status:            types.STREAM_ACCOUNT_STATUS_FROZEN,
		NetflowRate:       sdkmath.NewInt(0),
		FrozenNetflowRate: sdkmath.NewInt(100).Neg(),
		OutFlowCount:      1,
	}
	keeper.SetStreamRecord(ctx, streamRecord)

	// update fail when no force flag
	change := types.NewDefaultStreamRecordChangeWithAddr(user).
		WithLockBalanceChange(streamRecord.LockBalance.Neg())
	_, err := keeper.UpdateStreamRecordByAddr(ctx, change)
	require.ErrorContains(t, err, "is frozen")

	// update success when there is force flag
	ctx = ctx.WithValue(types.ForceUpdateStreamRecordKey, true)
	change = types.NewDefaultStreamRecordChangeWithAddr(user).
		WithLockBalanceChange(streamRecord.LockBalance.Neg())
	_, err = keeper.UpdateStreamRecordByAddr(ctx, change)
	require.NoError(t, err)

	streamRecord, _ = keeper.GetStreamRecord(ctx, user)
	require.True(t, streamRecord.Status == types.STREAM_ACCOUNT_STATUS_FROZEN)
	require.True(t, streamRecord.LockBalance.IsZero())
	require.True(t, streamRecord.StaticBalance.Int64() == 1000)
}

func TestSettleStreamRecord(t *testing.T) {
	keeper, ctx, _ := makePaymentKeeper(t)
	ctx = ctx.WithBlockTime(time.Unix(100, 0))
	user := sample.RandAccAddress()
	rate := sdkmath.NewInt(-100)
	staticBalance := sdkmath.NewInt(1e10)
	change := types.NewDefaultStreamRecordChangeWithAddr(user).WithRateChange(rate).WithStaticBalanceChange(staticBalance)
	sr := &types.StreamRecord{Account: user.String(),
		OutFlowCount:      1,
		StaticBalance:     sdkmath.ZeroInt(),
		BufferBalance:     sdkmath.ZeroInt(),
		LockBalance:       sdkmath.ZeroInt(),
		NetflowRate:       sdkmath.ZeroInt(),
		FrozenNetflowRate: sdkmath.ZeroInt(),
	}
	keeper.SetStreamRecord(ctx, sr)
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

func TestAutoForceSettle(t *testing.T) {
	keeper, ctx, depKeepers := makePaymentKeeper(t)
	t.Logf("depKeepers: %+v", depKeepers)
	params := keeper.GetParams(ctx)
	var startTime int64 = 100
	ctx = ctx.WithBlockTime(time.Unix(startTime, 0))
	user := sample.RandAccAddress()
	rate := sdkmath.NewInt(100)
	sp := sample.RandAccAddress()
	userInitBalance := sdkmath.NewInt(int64(100*params.VersionedParams.ReserveTime) + 1) // just enough for reserve
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
	outFlows := keeper.GetOutFlows(ctx, user)
	require.Equal(t, 1, len(outFlows))
	require.Equal(t, outFlows[0].ToAddress, sp.String())
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
	require.Equal(t, autoSettleQueue[0].Timestamp, startTime+int64(params.VersionedParams.ReserveTime)-int64(params.ForcedSettleTime))
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
	// reserve time - forced settle time - 1 day + 101s pass
	ctx = ctx.WithBlockTime(ctx.BlockTime().Add(time.Duration(params.VersionedParams.ReserveTime-params.ForcedSettleTime-86400+101) * time.Second))
	usrBeforeForceSettle, _ := keeper.GetStreamRecord(ctx, user)
	t.Logf("usrBeforeForceSettle: %s", usrBeforeForceSettle)

	ctx = ctx.WithValue(types.ForceUpdateStreamRecordKey, true)
	time.Sleep(1 * time.Second)
	keeper.AutoSettle(ctx)

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
