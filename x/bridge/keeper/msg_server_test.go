package keeper_test

import (
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/x/bridge/keeper"
	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
)

func TestCrossTransferOut(t *testing.T) {
	suite, k, ctx := keepertest.BridgeKeeper(t)

	msgServer := keeper.NewMsgServerImpl(*k)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	addr2, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)

	msgTransferOut := types.NewMsgTransferOut(addr1.String(), addr2.String(), &sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(1),
	})

	suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types2.FeeCollectorName, addr1, sdk.Coins{sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(1000),
	}})

	_, err = msgServer.TransferOut(ctx, msgTransferOut)
	if err != nil {
		println(err.Error())
	}
}
