package keeper_test

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

// nolint
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.SpKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

func TestKeeper(t *testing.T) {
	k, ctx := keepertest.SpKeeper(t)
	sp := types.StorageProvider{}
	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)

	sp.OperatorAddress = spAcc.String()

	k.SetStorageProvider(ctx, sp)
	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(t, found, true)
}
