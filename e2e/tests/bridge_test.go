package tests

import (
	"context"
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	_ "github.com/ghodss/yaml"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	types2 "github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
)

type BridgeTestSuite struct {
	core.BaseSuite
}

func (s *BridgeTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *BridgeTestSuite) SetupTest() {}

func (s *BridgeTestSuite) TestTransferOut() {
	users := s.GenAndChargeAccounts(2, 1000000)

	from, to := users[0], users[1]
	ctx := context.Background()

	// transfer out token
	transferAmount := sdkmath.NewInt(10000)
	msgTransferOut := &bridgetypes.MsgTransferOut{
		From:   from.GetAddr().String(),
		To:     to.GetAddr().String(),
		Amount: &types.Coin{Denom: types2.Denom, Amount: transferAmount},
	}

	params, err := s.Client.BridgeQueryClient.Params(ctx, &bridgetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	totalTransferOutRelayerFee := params.Params.BscTransferOutRelayerFee.Add(params.Params.BscTransferOutAckRelayerFee)

	moduleAccount := types.MustAccAddressFromHex("0xB73C0Aac4C1E606C6E495d848196355e6CB30381")
	// query balance before
	moduleBalanceBefore, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: moduleAccount.String(),
		Denom:   s.Config.Denom,
	})

	s.Require().NoError(err)

	s.T().Logf("balance before: %s %s", from.GetAddr().String(), moduleBalanceBefore.Balance.String())

	txRes := s.SendTxBlock(from, msgTransferOut)
	s.T().Log(txRes.RawLog)

	moduleBalanceAfter, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: moduleAccount.String(),
		Denom:   s.Config.Denom,
	})
	s.T().Logf("balance after: %s %s", from.GetAddr().String(), moduleBalanceAfter.Balance.String())
	s.Require().NoError(err)

	s.Require().Equal(moduleBalanceBefore.Balance.Amount.Add(transferAmount).Add(totalTransferOutRelayerFee).String(), moduleBalanceAfter.Balance.Amount.String())
}

func TestBridgeTestSuite(t *testing.T) {
	suite.Run(t, new(BridgeTestSuite))
}
