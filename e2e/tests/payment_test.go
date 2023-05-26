package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
)

type PaymentTestSuite struct {
	core.BaseSuite
}

func (s *PaymentTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *PaymentTestSuite) SetupTest() {}

func (s *PaymentTestSuite) TestPaymentAccount() {
	user := s.GenAndChargeAccounts(1, 100)[0]
	ctx := context.Background()
	// create a new payment account
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: user.GetAddr().String(),
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	// query user's payment accounts
	queryGetPaymentAccountsByOwnerRequest := paymenttypes.QueryGetPaymentAccountsByOwnerRequest{
		Owner: user.GetAddr().String(),
	}
	paymentAccounts, err := s.Client.GetPaymentAccountsByOwner(ctx, &queryGetPaymentAccountsByOwnerRequest)
	s.Require().NoError(err)
	s.T().Log(paymentAccounts)
	s.Require().Equal(1, len(paymentAccounts.PaymentAccounts))
	paymentAccountAddr := paymentAccounts.PaymentAccounts[0]
	// query this payment account
	queryGetPaymentAccountRequest := paymenttypes.QueryGetPaymentAccountRequest{
		Addr: paymentAccountAddr,
	}
	paymentAccount, err := s.Client.PaymentAccount(ctx, &queryGetPaymentAccountRequest)
	s.Require().NoError(err)
	s.T().Logf("payment account: %s", core.YamlString(paymentAccount.PaymentAccount))
	s.Require().Equal(user.GetAddr().String(), paymentAccount.PaymentAccount.Owner)
	s.Require().Equal(true, paymentAccount.PaymentAccount.Refundable)
	// set this payment account to non-refundable
	msgDisableRefund := &paymenttypes.MsgDisableRefund{
		Owner: user.GetAddr().String(),
		Addr:  paymentAccountAddr,
	}
	_ = s.SendTxBlock(user, msgDisableRefund)
	// query this payment account
	paymentAccount, err = s.Client.PaymentAccount(ctx, &queryGetPaymentAccountRequest)
	s.Require().NoError(err)
	s.T().Logf("payment account: %s", core.YamlString(paymentAccount.PaymentAccount))
	s.Require().Equal(false, paymentAccount.PaymentAccount.Refundable)
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}
