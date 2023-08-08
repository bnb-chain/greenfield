package keeper_test

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/x/bridge/keeper"
	"github.com/bnb-chain/greenfield/x/bridge/types"
)

func (s *TestSuite) TestTransferOutAck() {
	refundPackage := types.TransferOutRefundPackage{
		RefundAmount:  big.NewInt(1),
		RefundAddress: sdk.AccAddress("refundAddress"),
		RefundReason:  1,
	}

	packageBytes, err := refundPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*s.bridgeKeeper)

	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()

	// empty payload
	result := transferOutApp.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, nil)
	s.Require().Nil(result.Err, "result should be nil")
	s.Require().Nil(result.Payload, "result should be nil")

	// wrong payload
	result = transferOutApp.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, []byte{1})
	s.Require().Contains(result.Err.Error(), "deserialize transfer out refund package failed")

	// send coins failed
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test send coins error")).Times(1)
	result = transferOutApp.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Contains(result.Err.Error(), "test send coins error")

	// success case
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	result = transferOutApp.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(err, result.Err, "error should be nil")
}

func (s *TestSuite) TestTransferOutSynAndFailAck() {
	synPackage := types.TransferOutSynPackage{
		Amount:        big.NewInt(1),
		Recipient:     sdk.AccAddress{},
		RefundAddress: sdk.AccAddress("refundAddress"),
	}

	packageBytes, err := synPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferOutApp := keeper.NewTransferOutApp(*s.bridgeKeeper)

	s.crossChainKeeper.EXPECT().CreateRawIBCPackageWithFee(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), nil).AnyTimes()
	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()

	// syn package
	result := transferOutApp.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, nil)
	s.Require().Nil(result.Payload, "result should be nil")

	// fail ack package
	// wrong payload
	result = transferOutApp.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, []byte{1})
	s.Require().Contains(result.Err.Error(), "deserialize transfer out syn package failed")

	// send coins failed
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test send coins error")).Times(1)
	result = transferOutApp.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Contains(result.Err.Error(), "send coins error")

	// success case
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	result = transferOutApp.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(err, result.Err, "error should be nil")
}

func (s *TestSuite) TestTransferIn() {
	transferInSynPackage := types.TransferInSynPackage{
		Amount:          big.NewInt(1),
		ReceiverAddress: sdk.AccAddress("receiverAddress"),
		RefundAddress:   sdk.AccAddress("refundAddress"),
	}

	packageBytes, err := transferInSynPackage.Serialize()
	s.Require().Nil(err, "encode refund package error")

	transferInApp := keeper.NewTransferInApp(*s.bridgeKeeper)

	s.crossChainKeeper.EXPECT().CreateRawIBCPackageWithFee(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(uint64(0), nil).AnyTimes()
	s.stakingKeeper.EXPECT().BondDenom(gomock.Any()).Return("BNB").AnyTimes()

	// syn package
	// wrong payload
	s.Require().Panics(func() { transferInApp.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, []byte{1}) })

	// send coins failed and refund
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(fmt.Errorf("test send coins error")).Times(1)
	result := transferInApp.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Contains(result.Err.Error(), "test send coins error")
	unpacked, err := types.TransferInRefundPackageArgs.Unpack(result.Payload)
	s.Require().NoError(err)

	unpackedStruct := abi.ConvertType(unpacked[0], types.TransferInRefundPackageStruct{})
	pkgStruct, ok := unpackedStruct.(types.TransferInRefundPackageStruct)
	s.Require().True(ok)

	s.Require().Equal(transferInSynPackage.Amount, pkgStruct.RefundAmount)
	s.Require().Equal(transferInSynPackage.RefundAddress.String(), pkgStruct.RefundAddress.String())
	s.Require().Equal(uint32(types.REFUND_REASON_INSUFFICIENT_BALANCE), pkgStruct.RefundReason)

	// success case
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).Times(1)
	result = transferInApp.ExecuteSynPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, packageBytes)
	s.Require().Nil(result.Err, "error should be nil")

	// unexpected package type
	result = transferInApp.ExecuteAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, nil)
	s.Require().Nil(result.Payload, "result should be nil")

	result = transferInApp.ExecuteFailAckPackage(s.ctx, &sdk.CrossChainAppContext{Sequence: 1}, nil)
	s.Require().Nil(result.Payload, "result should be nil")
}
