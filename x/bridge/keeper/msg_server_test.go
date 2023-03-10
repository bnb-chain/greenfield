package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/bridge/keeper"
	"github.com/bnb-chain/greenfield/x/bridge/types"
)

func TestCrossTransferOut(t *testing.T) {
	suite, k, ctx := keepertest.BridgeKeeper(t)

	msgServer := keeper.NewMsgServerImpl(*k)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "error should be nil")

	addr2, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "error should be nil")

	msgTransferOut := types.NewMsgTransferOut(addr1.String(), addr2.String(), &sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(1),
	})

	err = suite.BankKeeper.SendCoinsFromModuleToAccount(ctx, types2.FeeCollectorName, addr1, sdk.Coins{sdk.Coin{
		Denom:  "stake",
		Amount: sdk.NewInt(5000000000000000),
	}})
	require.Nil(t, err, "error should be nil")

	_, err = msgServer.TransferOut(ctx, msgTransferOut)
	require.Nil(t, err, "error should be nil")
}

func TestCrossTransferOutWrong(t *testing.T) {
	suite, k, ctx := keepertest.BridgeKeeper(t)

	msgServer := keeper.NewMsgServerImpl(*k)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "error should be nil")

	addr2, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "error should be nil")

	msgTransferOut := types.NewMsgTransferOut(addr1.String(), addr2.String(), &sdk.Coin{
		Denom:  "wrongdenom",
		Amount: sdk.NewInt(1),
	})

	_, err = msgServer.TransferOut(ctx, msgTransferOut)
	require.NotNil(t, err, "error should not be nil")
	require.Contains(t, err.Error(), "denom is not supported")
}
