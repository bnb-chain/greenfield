package keeper_test

import (
	"context"
	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/stretchr/testify/suite"
	"testing"
)

// TODO(chris Li) add some keeper unit test
type KeeperTestSuite struct {
	suite.Suite

	keeper      *keeper.Keeper
	ctx         context.Context
	queryClient proposal.QueryClient
	msgServer   types.MsgServer
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
