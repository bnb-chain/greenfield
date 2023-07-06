package types

import (
	"math/big"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)

const (
	TransferOutChannel = "transferOut"
	TransferInChannel  = "transferIn"

	TransferOutChannelID sdk.ChannelID = 1
	TransferInChannelID  sdk.ChannelID = 2
	SyncParamsChannelID                = types.SyncParamsChannelID
)

func SafeBigInt(input *big.Int) *big.Int {
	if input == nil {
		return big.NewInt(0)
	}
	return input
}

type TransferOutSynPackage struct {
	Amount        *big.Int
	Recipient     sdk.AccAddress
	RefundAddress sdk.AccAddress
}

type TransferOutSynPackageStruct struct {
	Amount        *big.Int
	Recipient     common.Address
	RefundAddress common.Address
}

var (
	transferOutSynPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Amount", Type: "uint256"},
		{Name: "Recipient", Type: "address"},
		{Name: "RefundAddress", Type: "address"},
	})

	transferOutSynPackageArgs = abi.Arguments{
		{Type: transferOutSynPackageType},
	}
)

func (pkg *TransferOutSynPackage) Serialize() ([]byte, error) {
	return transferOutSynPackageArgs.Pack(&TransferOutSynPackageStruct{
		SafeBigInt(pkg.Amount),
		common.BytesToAddress(pkg.Recipient),
		common.BytesToAddress(pkg.RefundAddress),
	})
}

func DeserializeTransferOutSynPackage(serializedPackage []byte) (*TransferOutSynPackage, error) {
	unpacked, err := transferOutSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidPackage, "deserialize transfer out sync package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], TransferOutSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(TransferOutSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidPackage, "reflect transfer out sync package failed")
	}

	tp := TransferOutSynPackage{
		pkgStruct.Amount,
		pkgStruct.Recipient.Bytes(),
		pkgStruct.RefundAddress.Bytes(),
	}
	return &tp, nil
}

type TransferOutRefundPackage struct {
	RefundAmount *big.Int
	RefundAddr   sdk.AccAddress
	RefundReason uint32
}

type TransferOutRefundPackageStruct struct {
	RefundAmount *big.Int
	RefundAddr   common.Address
	RefundReason uint32
}

var (
	transferOutRefundPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "RefundAmount", Type: "uint256"},
		{Name: "RefundAddr", Type: "address"},
		{Name: "RefundReason", Type: "uint32"},
	})

	transferOutRefundPackageArgs = abi.Arguments{
		{Type: transferOutRefundPackageType},
	}
)

func (pkg *TransferOutRefundPackage) Serialize() ([]byte, error) {
	return transferOutRefundPackageArgs.Pack(&TransferOutRefundPackageStruct{
		SafeBigInt(pkg.RefundAmount),
		common.BytesToAddress(pkg.RefundAddr),
		pkg.RefundReason,
	})
}

func DeserializeTransferOutRefundPackage(serializedPackage []byte) (*TransferOutRefundPackage, error) {
	unpacked, err := transferOutRefundPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidPackage, "deserialize transfer out refund package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], TransferOutRefundPackageStruct{})
	pkgStruct, ok := unpackedStruct.(TransferOutRefundPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidPackage, "reflect transfer out refund package failed")
	}

	tp := TransferOutRefundPackage{
		pkgStruct.RefundAmount,
		pkgStruct.RefundAddr.Bytes(),
		pkgStruct.RefundReason,
	}
	return &tp, nil
}

type TransferInSynPackage struct {
	Amount          *big.Int
	ReceiverAddress sdk.AccAddress
	RefundAddress   sdk.AccAddress
}

type TransferInSynPackageStruct struct {
	Amount          *big.Int
	ReceiverAddress common.Address
	RefundAddress   common.Address
}

var (
	transferInSynPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "Amount", Type: "uint256"},
		{Name: "ReceiverAddress", Type: "address"},
		{Name: "RefundAddress", Type: "address"},
	})

	transferInSynPackageArgs = abi.Arguments{
		{Type: transferInSynPackageType},
	}
)

func (pkg *TransferInSynPackage) Serialize() ([]byte, error) {
	return transferInSynPackageArgs.Pack(&TransferInSynPackageStruct{
		SafeBigInt(pkg.Amount),
		common.BytesToAddress(pkg.ReceiverAddress),
		common.BytesToAddress(pkg.RefundAddress),
	})
}

func DeserializeTransferInSynPackage(serializedPackage []byte) (*TransferInSynPackage, error) {
	unpacked, err := transferInSynPackageArgs.Unpack(serializedPackage)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidPackage, "deserialize transfer in sync package failed")
	}

	unpackedStruct := abi.ConvertType(unpacked[0], TransferInSynPackageStruct{})
	pkgStruct, ok := unpackedStruct.(TransferInSynPackageStruct)
	if !ok {
		return nil, errors.Wrapf(ErrInvalidPackage, "reflect transfer in sync package failed")
	}

	tp := TransferInSynPackage{
		pkgStruct.Amount,
		pkgStruct.ReceiverAddress.Bytes(),
		pkgStruct.RefundAddress.Bytes(),
	}
	return &tp, nil
}

type TransferInRefundPackage struct {
	RefundAmount  *big.Int
	RefundAddress sdk.AccAddress
	RefundReason  uint32
}

type TransferInRefundPackageStruct struct {
	RefundAmount  *big.Int
	RefundAddress common.Address
	RefundReason  uint32
}

var (
	transferInRefundPackageType, _ = abi.NewType("tuple", "", []abi.ArgumentMarshaling{
		{Name: "RefundAmount", Type: "uint256"},
		{Name: "RefundAddr", Type: "address"},
		{Name: "RefundReason", Type: "uint32"},
	})

	transferInRefundPackageArgs = abi.Arguments{
		{Type: transferInRefundPackageType},
	}
)

func (pkg *TransferInRefundPackage) Serialize() ([]byte, error) {
	return transferInRefundPackageArgs.Pack(&TransferInRefundPackageStruct{
		SafeBigInt(pkg.RefundAmount),
		common.BytesToAddress(pkg.RefundAddress),
		pkg.RefundReason,
	})
}
