package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (s *TestSuite) TestWithdraw_Fail() {
	creator1 := sample.RandAccAddress()
	paymentAddr1 := sample.RandAccAddress()

	// stream record not found
	msg := types.NewMsgWithdraw(creator1.String(), sample.RandAccAddress().String(), sdkmath.NewInt(100))
	_, err := s.msgServer.Withdraw(s.ctx, msg)
	s.Require().Error(err)

	// stream record is frozen
	record1 := types.NewStreamRecord(paymentAddr1, s.ctx.BlockTime().Unix())
	record1.Status = types.STREAM_ACCOUNT_STATUS_FROZEN
	s.paymentKeeper.SetStreamRecord(s.ctx, record1)

	msg = types.NewMsgWithdraw(creator1.String(), paymentAddr1.String(), sdkmath.NewInt(100))
	_, err = s.msgServer.Withdraw(s.ctx, msg)
	s.Require().Error(err)

	record1.Status = types.STREAM_ACCOUNT_STATUS_ACTIVE
	s.paymentKeeper.SetStreamRecord(s.ctx, record1)

	// payment account does not exist
	msg = types.NewMsgWithdraw(creator1.String(), paymentAddr1.String(), sdkmath.NewInt(100))
	_, err = s.msgServer.Withdraw(s.ctx, msg)
	s.Require().Error(err)

	// the message is not from the owner
	creator2 := sample.RandAccAddress()
	createAccountMsg := types.NewMsgCreatePaymentAccount(creator2.String())
	_, err = s.msgServer.CreatePaymentAccount(s.ctx, createAccountMsg)
	s.Require().NoError(err)
	paymentAddr2 := s.paymentKeeper.DerivePaymentAccountAddress(creator2, 0)
	paymentAccountRecord, _ := s.paymentKeeper.GetPaymentAccount(s.ctx, paymentAddr2)
	s.Require().True(paymentAccountRecord.Owner == creator2.String())

	record2 := types.NewStreamRecord(paymentAddr2, s.ctx.BlockTime().Unix())
	s.paymentKeeper.SetStreamRecord(s.ctx, record2)

	msg = types.NewMsgWithdraw(creator1.String(), paymentAddr2.String(), sdkmath.NewInt(100))
	_, err = s.msgServer.Withdraw(s.ctx, msg)
	s.Require().Error(err)

	// cannot withdraw after disable refund
	disableRefundMsg := types.NewMsgDisableRefund(creator2.String(), paymentAddr2.String())
	_, err = s.msgServer.DisableRefund(s.ctx, disableRefundMsg)
	s.Require().NoError(err)
	paymentAccountRecord, _ = s.paymentKeeper.GetPaymentAccount(s.ctx, paymentAddr2)
	s.Require().True(paymentAccountRecord.Refundable == false)

	msg = types.NewMsgWithdraw(creator2.String(), paymentAddr2.String(), sdkmath.NewInt(100))
	_, err = s.msgServer.Withdraw(s.ctx, msg)
	s.Require().Error(err)
}

func (s *TestSuite) TestWithdraw_Success() {
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	s.accountKeeper.EXPECT().HasAccount(gomock.Any(), gomock.Any()).
		Return(true).AnyTimes()

	creator := sample.RandAccAddress()
	createAccountMsg := types.NewMsgCreatePaymentAccount(creator.String())
	_, err := s.msgServer.CreatePaymentAccount(s.ctx, createAccountMsg)
	s.Require().NoError(err)
	paymentAddr := s.paymentKeeper.DerivePaymentAccountAddress(creator, 0)
	paymentAccountRecord, _ := s.paymentKeeper.GetPaymentAccount(s.ctx, paymentAddr)
	s.Require().True(paymentAccountRecord.Owner == creator.String())

	record := types.NewStreamRecord(paymentAddr, s.ctx.BlockTime().Unix())
	record.StaticBalance = sdkmath.NewInt(200)
	s.paymentKeeper.SetStreamRecord(s.ctx, record)

	msg := types.NewMsgWithdraw(creator.String(), paymentAddr.String(), sdkmath.NewInt(100))
	_, err = s.msgServer.Withdraw(s.ctx, msg)
	s.Require().NoError(err)
}
