package keeper_test

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/bridge/keeper"
	"github.com/bnb-chain/greenfield/x/bridge/types"
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

func (s *TestSuite) TestTransferOutAck() {
	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "generate key failed")

	refundPackage := types.TransferOutRefundPackage{
		RefundAmount: big.NewInt(1),
		RefundAddr:   addr1,
		RefundReason: 1,
	}

	packageBytes, err := refundPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*s.bridgeKeeper)

	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	result := transferOutApp.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(err, result.Err, "error should be nil")
}

func (s *TestSuite) TestTransferOutFailAck() {
	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "generate key failed")

	synPackage := types.TransferOutSynPackage{
		Amount:        big.NewInt(1),
		Recipient:     sdk.AccAddress{},
		RefundAddress: addr1,
	}

	packageBytes, err := synPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*s.bridgeKeeper)

	s.crossChainKeeper.EXPECT().CreateRawIBCPackageWithFee(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), nil).AnyTimes()
	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	result := transferOutApp.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(err, result.Err, "error should be nil")
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
				RefundAddress:   sdk.AccAddress{1},
			},
			expectedPass: false,
			errorMsg:     "receiver address should not be empty",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(1),
				ReceiverAddress: sdk.AccAddress{1},
				RefundAddress:   sdk.AccAddress{},
			},
			expectedPass: false,
			errorMsg:     "refund address should not be empty",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(-1),
				ReceiverAddress: sdk.AccAddress{1},
				RefundAddress:   sdk.AccAddress{1},
			},
			expectedPass: false,
			errorMsg:     "amount should not be negative",
		},
		{
			transferInPackage: types.TransferInSynPackage{
				Amount:          big.NewInt(1),
				ReceiverAddress: sdk.AccAddress{1},
				RefundAddress:   sdk.AccAddress{1},
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

func (s *TestSuite) TestTransferInSyn() {
	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "generate key failed")

	transferInSynPackage := types.TransferInSynPackage{
		Amount:          big.NewInt(1),
		ReceiverAddress: addr1,
		RefundAddress:   sdk.AccAddress{1},
	}

	packageBytes, err := transferInSynPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferInApp := keeper.NewTransferInApp(*s.bridgeKeeper)

	s.crossChainKeeper.EXPECT().CreateRawIBCPackageWithFee(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), nil).AnyTimes()
	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	result := transferInApp.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(err, result.Err, "error should be nil")
}

func (s *TestSuite) TestTransferInRefund() {
	addr1, _, err := testutil.GenerateCoinKey(hd.Secp256k1, s.cdc)
	s.Require().Nil(err, "generate key failed")

	transferInRefundPackage := types.TransferInRefundPackage{
		RefundAmount:  big.NewInt(1),
		RefundAddress: addr1,
		RefundReason:  123,
	}

	packageBytes, err := transferInRefundPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferInApp := keeper.NewTransferInApp(*s.bridgeKeeper)

	s.crossChainKeeper.EXPECT().CreateRawIBCPackageWithFee(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), nil).AnyTimes()
	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	result := transferInApp.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(err, result.Err, "error should be nil")
}
