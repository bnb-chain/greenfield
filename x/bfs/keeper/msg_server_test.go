package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/x/bfs/keeper"
	"github.com/bnb-chain/bfs/x/bfs/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.BfsKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
