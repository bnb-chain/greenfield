package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

// TODO(chris Li) add some keeper unit test
type KeeperTestSuite struct {
	suite.Suite

	keeper    *keeper.Keeper
	ctx       context.Context
	msgServer types.MsgServer
}

func (suite *KeeperTestSuite) SetupTest() {
	k, ctx := keepertest.SpKeeper(suite.T())
	suite.msgServer = keeper.NewMsgServerImpl(*k)
	suite.ctx = sdk.WrapSDKContext(ctx)
	suite.keeper = k
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}
