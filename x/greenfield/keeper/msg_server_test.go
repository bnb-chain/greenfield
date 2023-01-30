package keeper_test

import (
	"context"
	"testing"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/greenfield/keeper"
	"github.com/bnb-chain/greenfield/x/greenfield/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// nolint: unused
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.GreenfieldKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
