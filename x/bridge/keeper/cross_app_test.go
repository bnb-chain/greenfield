package keeper_test

import (
	"bytes"
	"math/big"
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/x/bridge/keeper"
	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	"github.com/stretchr/testify/require"
)

func TestTransferOutCheck(t *testing.T) {
	tests := []struct {
		refundPackage types.TransferOutRefundPackage
		expectedPass  bool
		errorMsg      string
	}{
		{
			refundPackage: types.TransferOutRefundPackage{
				RefundAmount: big.NewInt(1),
				RefundAddr:   []byte{},
				RefundReason: 0,
			},
			expectedPass: false,
			errorMsg:     "refund address is empty",
		},
		{
			refundPackage: types.TransferOutRefundPackage{
				RefundAmount: big.NewInt(-1),
				RefundAddr:   bytes.Repeat([]byte{1}, 20),
				RefundReason: 0,
			},
			expectedPass: false,
			errorMsg:     "amount to refund should not be negative",
		},
		{
			refundPackage: types.TransferOutRefundPackage{
				RefundAmount: big.NewInt(1),
				RefundAddr:   bytes.Repeat([]byte{1}, 20),
				RefundReason: 0,
			},
			expectedPass: true,
		},
	}

	crossApp := keeper.NewTransferOutApp(keeper.Keeper{})
	for _, test := range tests {
		err := crossApp.CheckPackage(&test.refundPackage)
		if test.expectedPass {
			require.Nil(t, err, "error should be nil")
		} else {
			require.NotNil(t, err, " error should not be nil")
			require.Contains(t, err.Error(), test.errorMsg)
		}
	}
}

func TestTransferOutAck(t *testing.T) {
	suite, _, ctx := keepertest.BridgeKeeper(t)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "generate key failed")

	refundPackage := types.TransferOutRefundPackage{
		RefundAmount: big.NewInt(1),
		RefundAddr:   addr1,
		RefundReason: 1,
	}

	packageBytes, err := rlp.EncodeToBytes(&refundPackage)
	require.Nil(t, err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*suite.BridgeKeeper)

	crossChainAccount := suite.AccountKeeper.GetModuleAccount(ctx, crosschaintypes.ModuleName)

	moduleBalanceBefore := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")

	result := transferOutApp.ExecuteAckPackage(ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	require.Nil(t, result.Err, "error should be nil")
	moduleBalanceAfter := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")
	accountBalanceAfter := suite.BankKeeper.GetBalance(ctx, refundPackage.RefundAddr, "stake")

	require.Equal(t, big.NewInt(0).Add(moduleBalanceAfter.Amount.BigInt(), refundPackage.RefundAmount).String(), moduleBalanceBefore.Amount.BigInt().String())
	require.Equal(t, refundPackage.RefundAmount.String(), accountBalanceAfter.Amount.BigInt().String())
}

func TestTransferOutFailAck(t *testing.T) {
	suite, _, ctx := keepertest.BridgeKeeper(t)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "generate key failed")

	synPackage := types.TransferOutSynPackage{
		Amount:        big.NewInt(1),
		Recipient:     sdk.EthAddress{},
		RefundAddress: addr1,
	}

	packageBytes, err := rlp.EncodeToBytes(&synPackage)
	require.Nil(t, err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*suite.BridgeKeeper)

	crossChainAccount := suite.AccountKeeper.GetModuleAccount(ctx, crosschaintypes.ModuleName)

	moduleBalanceBefore := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")

	result := transferOutApp.ExecuteFailAckPackage(ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	require.Nil(t, result.Err, "error should be nil")
	moduleBalanceAfter := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")
	accountBalanceAfter := suite.BankKeeper.GetBalance(ctx, synPackage.RefundAddress, "stake")

	require.Equal(t, big.NewInt(0).Add(moduleBalanceAfter.Amount.BigInt(), synPackage.Amount).String(), moduleBalanceBefore.Amount.BigInt().String())
	require.Equal(t, synPackage.Amount.String(), accountBalanceAfter.Amount.BigInt().String())
}

func TestTransferInCheck(t *testing.T) {
	tests := []struct {
		transferInPackage types.TransferInSynPackage
		expectedPass      bool
		errorMsg          string
	}{
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(1),
				ReceiverAddress: sdk.AccAddress{},
				RefundAddress:   sdk.EthAddress{1},
			},
			expectedPass: false,
			errorMsg:     "receiver address should not be empty",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(1),
				ReceiverAddress: sdk.AccAddress{1},
				RefundAddress:   sdk.EthAddress{},
			},
			expectedPass: false,
			errorMsg:     "refund address should not be empty",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(-1),
				ReceiverAddress: sdk.AccAddress{1},
				RefundAddress:   sdk.EthAddress{1},
			},
			expectedPass: false,
			errorMsg:     "amount should not be negative",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(1),
				ReceiverAddress: sdk.AccAddress{1},
				RefundAddress:   sdk.EthAddress{1},
			},
			expectedPass: true,
		},
	}

	crossApp := keeper.NewTransferInApp(keeper.Keeper{})
	for _, test := range tests {
		err := crossApp.CheckTransferInSynPackage(&test.transferInPackage)
		if test.expectedPass {
			require.Nil(t, err, "error should be nil")
		} else {
			require.NotNil(t, err, " error should not be nil")
			require.Contains(t, err.Error(), test.errorMsg)
		}
	}
}

func TestTransferInSyn(t *testing.T) {
	suite, _, ctx := keepertest.BridgeKeeper(t)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "generate key failed")

	transferInSynPackage := types.TransferInSynPackage{
		Amount:          big.NewInt(1),
		ReceiverAddress: addr1,
		RefundAddress:   sdk.EthAddress{1},
	}

	packageBytes, err := rlp.EncodeToBytes(&transferInSynPackage)
	require.Nil(t, err, "encode refund package error")

	transferInApp := keeper.NewTransferInApp(*suite.BridgeKeeper)

	crossChainAccount := suite.AccountKeeper.GetModuleAccount(ctx, crosschaintypes.ModuleName)

	moduleBalanceBefore := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")

	result := transferInApp.ExecuteSynPackage(ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	require.Nil(t, result.Err, "error should be nil")

	moduleBalanceAfter := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")
	accountBalanceAfter := suite.BankKeeper.GetBalance(ctx, addr1, "stake")

	require.Equal(t, transferInSynPackage.Amount.String(), accountBalanceAfter.Amount.BigInt().String())

	require.Equal(t, big.NewInt(0).Add(moduleBalanceAfter.Amount.BigInt(), transferInSynPackage.Amount).String(), moduleBalanceBefore.Amount.BigInt().String())
}
