package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

// nolint
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.SpKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}
