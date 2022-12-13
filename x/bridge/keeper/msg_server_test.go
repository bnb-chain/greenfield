package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/x/bridge/keeper"
	"github.com/bnb-chain/bfs/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.BridgeKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
