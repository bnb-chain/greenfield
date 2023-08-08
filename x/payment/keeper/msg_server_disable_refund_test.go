package keeper_test

import (
	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (s *TestSuite) TestDisableRefund() {
	// payment account does not exist
	creator1 := sample.RandAccAddress()
	msg := types.NewMsgDisableRefund(creator1.String(), sample.RandAccAddress().String())
	_, err := s.msgServer.DisableRefund(s.ctx, msg)
	s.Require().Error(err)

	// the message is not from the owner
	creator2 := sample.RandAccAddress()
	createAccountMsg := types.NewMsgCreatePaymentAccount(creator2.String())
	_, err = s.msgServer.CreatePaymentAccount(s.ctx, createAccountMsg)
	s.Require().NoError(err)
	paymentAccountAddr := s.paymentKeeper.DerivePaymentAccountAddress(creator2, 0)
	record, _ := s.paymentKeeper.GetPaymentAccount(s.ctx, paymentAccountAddr)
	s.Require().True(record.Owner == creator2.String())

	msg = types.NewMsgDisableRefund(creator1.String(), paymentAccountAddr.String())
	_, err = s.msgServer.DisableRefund(s.ctx, msg)
	s.Require().Error(err)

	// disable refund success
	msg = types.NewMsgDisableRefund(creator2.String(), paymentAccountAddr.String())
	_, err = s.msgServer.DisableRefund(s.ctx, msg)
	s.Require().NoError(err)
	record, _ = s.paymentKeeper.GetPaymentAccount(s.ctx, paymentAccountAddr)
	s.Require().True(record.Refundable == false)

	// cannot disable it again
	msg = types.NewMsgDisableRefund(creator2.String(), paymentAccountAddr.String())
	_, err = s.msgServer.DisableRefund(s.ctx, msg)
	s.Require().Error(err)
}
