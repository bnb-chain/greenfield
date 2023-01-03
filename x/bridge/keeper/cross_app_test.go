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
				TokenSymbol:  [32]byte{},
				RefundAmount: big.NewInt(1),
				RefundAddr:   []byte{},
				RefundReason: 0,
			},
			expectedPass: false,
			errorMsg:     "refund address is empty",
		},
		{
			refundPackage: types.TransferOutRefundPackage{
				TokenSymbol:  [32]byte{},
				RefundAmount: big.NewInt(-1),
				RefundAddr:   bytes.Repeat([]byte{1}, 20),
				RefundReason: 0,
			},
			expectedPass: false,
			errorMsg:     "amount to refund should not be negative",
		},
		{
			refundPackage: types.TransferOutRefundPackage{
				TokenSymbol:  [32]byte{},
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
		TokenSymbol:  types.SymbolToBytes("stake"),
		RefundAmount: big.NewInt(1),
		RefundAddr:   addr1,
		RefundReason: 1,
	}

	packageBytes, err := rlp.EncodeToBytes(&refundPackage)
	require.Nil(t, err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*suite.BridgeKeeper)

	crossChainAccount := suite.AccountKeeper.GetModuleAccount(ctx, crosschaintypes.ModuleName)

	moduleBalanceBefore := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")

	result := transferOutApp.ExecuteAckPackage(ctx, packageBytes)
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
		TokenSymbol:     types.SymbolToBytes("stake"),
		ContractAddress: sdk.EthAddress{},
		Amount:          big.NewInt(1),
		Recipient:       sdk.EthAddress{},
		RefundAddress:   addr1,
	}

	packageBytes, err := rlp.EncodeToBytes(&synPackage)
	require.Nil(t, err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*suite.BridgeKeeper)

	crossChainAccount := suite.AccountKeeper.GetModuleAccount(ctx, crosschaintypes.ModuleName)

	moduleBalanceBefore := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")

	result := transferOutApp.ExecuteFailAckPackage(ctx, packageBytes)
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
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{},
				ReceiverAddresses: []sdk.AccAddress{},
				RefundAddresses:   []sdk.EthAddress{},
			},
			expectedPass: false,
			errorMsg:     "length of Amounts should not be 0",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{big.NewInt(1)},
				ReceiverAddresses: []sdk.AccAddress{},
				RefundAddresses:   []sdk.EthAddress{},
			},
			expectedPass: false,
			errorMsg:     "ength of RefundAddresses, ReceiverAddresses, Amounts should be the same",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{big.NewInt(1)},
				ReceiverAddresses: []sdk.AccAddress{sdk.AccAddress{}},
				RefundAddresses:   []sdk.EthAddress{},
			},
			expectedPass: false,
			errorMsg:     "length of RefundAddresses, ReceiverAddresses, Amounts should be the same",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{big.NewInt(1)},
				ReceiverAddresses: []sdk.AccAddress{sdk.AccAddress{}},
				RefundAddresses:   []sdk.EthAddress{sdk.EthAddress{}},
			},
			expectedPass: false,
			errorMsg:     "refund address should not be empty",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{big.NewInt(1)},
				ReceiverAddresses: []sdk.AccAddress{sdk.AccAddress{}},
				RefundAddresses:   []sdk.EthAddress{sdk.EthAddress{1}},
			},
			expectedPass: false,
			errorMsg:     "receiver address should not be empty",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{big.NewInt(-1)},
				ReceiverAddresses: []sdk.AccAddress{sdk.AccAddress{1}},
				RefundAddresses:   []sdk.EthAddress{sdk.EthAddress{1}},
			},
			expectedPass: false,
			errorMsg:     "amount should not be negative",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				TokenSymbol:       [32]byte{},
				ContractAddress:   sdk.EthAddress{},
				Amounts:           []*big.Int{big.NewInt(1)},
				ReceiverAddresses: []sdk.AccAddress{sdk.AccAddress{1}},
				RefundAddresses:   []sdk.EthAddress{sdk.EthAddress{1}},
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
		TokenSymbol:       types.SymbolToBytes("stake"),
		ContractAddress:   sdk.EthAddress{},
		Amounts:           []*big.Int{big.NewInt(1)},
		ReceiverAddresses: []sdk.AccAddress{addr1},
		RefundAddresses:   []sdk.EthAddress{sdk.EthAddress{1}},
	}

	packageBytes, err := rlp.EncodeToBytes(&transferInSynPackage)
	require.Nil(t, err, "encode refund package error")

	transferInApp := keeper.NewTransferInApp(*suite.BridgeKeeper)

	crossChainAccount := suite.AccountKeeper.GetModuleAccount(ctx, crosschaintypes.ModuleName)

	moduleBalanceBefore := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")

	result := transferInApp.ExecuteSynPackage(ctx, packageBytes, big.NewInt(0))
	require.Nil(t, result.Err, "error should be nil")

	moduleBalanceAfter := suite.BankKeeper.GetBalance(ctx, crossChainAccount.GetAddress(), "stake")
	accountBalanceAfter := suite.BankKeeper.GetBalance(ctx, addr1, "stake")

	require.Equal(t, transferInSynPackage.Amounts[0].String(), accountBalanceAfter.Amount.BigInt().String())

	require.Equal(t, big.NewInt(0).Add(moduleBalanceAfter.Amount.BigInt(), transferInSynPackage.Amounts[0]).String(), moduleBalanceBefore.Amount.BigInt().String())
}

func TestTransferInSynWrong(t *testing.T) {
	suite, _, ctx := keepertest.BridgeKeeper(t)

	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, suite.Cdc)
	require.Nil(t, err, "generate key failed")

	transferInSynPackage := types.TransferInSynPackage{
		TokenSymbol:       types.SymbolToBytes("wrongdenom"),
		ContractAddress:   sdk.EthAddress{},
		Amounts:           []*big.Int{big.NewInt(1)},
		ReceiverAddresses: []sdk.AccAddress{addr1},
		RefundAddresses:   []sdk.EthAddress{sdk.EthAddress{1}},
	}

	packageBytes, err := rlp.EncodeToBytes(&transferInSynPackage)
	require.Nil(t, err, "encode refund package error")

	transferInApp := keeper.NewTransferInApp(*suite.BridgeKeeper)

	result := transferInApp.ExecuteSynPackage(ctx, packageBytes, big.NewInt(0))
	require.NotNil(t, result.Err, "error should not be nil")
	require.Contains(t, result.Err.Error(), "denom is not unsupported")
}
