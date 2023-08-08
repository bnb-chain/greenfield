package keeper_test

import (
	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (s *TestSuite) TestCreatePaymentAccount() {
	creator := sample.RandAccAddress()

	// create first one
	msg := types.NewMsgCreatePaymentAccount(creator.String())
	_, err := s.msgServer.CreatePaymentAccount(s.ctx, msg)
	s.Require().NoError(err)

	record, _ := s.paymentKeeper.GetPaymentAccountCount(s.ctx, creator)
	s.Require().True(record.Count == 1)

	// create another one
	msg = types.NewMsgCreatePaymentAccount(creator.String())
	_, err = s.msgServer.CreatePaymentAccount(s.ctx, msg)
	s.Require().NoError(err)

	record, _ = s.paymentKeeper.GetPaymentAccountCount(s.ctx, creator)
	s.Require().True(record.Count == 2)

	// limit the number of payment account
	params := s.paymentKeeper.GetParams(s.ctx)
	params.PaymentAccountCountLimit = 2
	_ = s.paymentKeeper.SetParams(s.ctx, params)

	msg = types.NewMsgCreatePaymentAccount(creator.String())
	_, err = s.msgServer.CreatePaymentAccount(s.ctx, msg)
	s.Require().Error(err)
}
