package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (s *TestSuite) TestDeposit_ToBankAccount() {
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	s.accountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	// deposit to self
	owner := sample.RandAccAddress()
	msg := types.NewMsgDeposit(owner.String(), owner.String(), sdkmath.NewInt(1000))
	_, err := s.msgServer.Deposit(s.ctx, msg)
	s.Require().NoError(err)
	record, _ := s.paymentKeeper.GetStreamRecord(s.ctx, owner)
	s.Require().True(record.StaticBalance.Int64() == msg.Amount.Int64())

	// deposit to other account
	to := sample.RandAccAddress()
	msg = types.NewMsgDeposit(owner.String(), to.String(), sdkmath.NewInt(1000))
	_, err = s.msgServer.Deposit(s.ctx, msg)
	s.Require().NoError(err)
	record, _ = s.paymentKeeper.GetStreamRecord(s.ctx, to)
	s.Require().True(record.StaticBalance.Int64() == msg.Amount.Int64())
}

func (s *TestSuite) TestDeposit_ToPaymentAccountAccount() {
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	s.accountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	// deposit to self
	owner := sample.RandAccAddress()
	paymentAddress := s.paymentKeeper.DerivePaymentAccountAddress(owner, 0)
	paymentAccount1 := &types.PaymentAccount{
		Owner:      owner.String(),
		Addr:       paymentAddress.String(),
		Refundable: true,
	}

	s.ctx = s.ctx.WithBlockHeight(10)

	// set
	s.paymentKeeper.SetPaymentAccount(s.ctx, paymentAccount1)

	msg := types.NewMsgDeposit(owner.String(), paymentAddress.String(), sdkmath.NewInt(1000))
	_, err := s.msgServer.Deposit(s.ctx, msg)
	s.Require().NoError(err)
	record, _ := s.paymentKeeper.GetStreamRecord(s.ctx, owner)
	s.Require().True(record.StaticBalance.Int64() == msg.Amount.Int64())

	// deposit to other account
	to := sample.RandAccAddress()
	msg = types.NewMsgDeposit(owner.String(), to.String(), sdkmath.NewInt(1000))
	_, err = s.msgServer.Deposit(s.ctx, msg)
	s.Require().NoError(err)
	record, _ = s.paymentKeeper.GetStreamRecord(s.ctx, to)
	s.Require().True(record.StaticBalance.Int64() == msg.Amount.Int64())
}

func (s *TestSuite) TestDeposit_ToActiveStreamRecord() {
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	s.accountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	owner := sample.RandAccAddress()
	paymentAddr := sample.RandAccAddress()
	record := types.NewStreamRecord(paymentAddr, s.ctx.BlockTime().Unix())
	s.paymentKeeper.SetStreamRecord(s.ctx, record)

	// deposit to active stream record
	msg := types.NewMsgDeposit(owner.String(), paymentAddr.String(), sdkmath.NewInt(1000))
	_, err := s.msgServer.Deposit(s.ctx, msg)
	s.Require().NoError(err)
	recordAfter, _ := s.paymentKeeper.GetStreamRecord(s.ctx, paymentAddr)
	s.Require().True(recordAfter.StaticBalance.Int64() == msg.Amount.Int64())
}

func (s *TestSuite) TestDeposit_ToFrozenStreamRecord() {
	s.bankKeeper.EXPECT().SendCoinsFromAccountToModule(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	s.accountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	owner := sample.RandAccAddress()
	paymentAddr := sample.RandAccAddress()
	record := types.NewStreamRecord(paymentAddr, s.ctx.BlockTime().Unix())
	record.Status = types.STREAM_ACCOUNT_STATUS_FROZEN
	record.FrozenNetflowRate = sdkmath.NewInt(-10)
	s.paymentKeeper.SetStreamRecord(s.ctx, record)

	// deposit to frozen stream record
	msg := types.NewMsgDeposit(owner.String(), paymentAddr.String(), sdkmath.NewInt(1000))
	_, err := s.msgServer.Deposit(s.ctx, msg)
	s.Require().NoError(err)
	recordAfter, _ := s.paymentKeeper.GetStreamRecord(s.ctx, paymentAddr)
	s.Require().True(recordAfter.StaticBalance.Int64() == msg.Amount.Int64())
}
