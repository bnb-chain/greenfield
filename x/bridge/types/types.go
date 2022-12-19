package types

import (
	"bytes"
	"math/big"
	"time"

	"cosmossdk.io/errors"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MinTransferOutExpireTimeGap = 60 * time.Second

	TransferOutChannel = "transferOut"
	TransferInChannel  = "transferIn"

	TransferOutChannelID sdk.ChannelID = 1
	TransferInChannelID  sdk.ChannelID = 3
)

var CrossTransferOutRelayFee = sdk.NewInt(1) // TODO: to be determined

type RefundReason uint32

const (
	UnsupportedSymbol   RefundReason = 1
	Timeout             RefundReason = 2
	InsufficientBalance RefundReason = 3
	Unknown             RefundReason = 4
)

type TransferOutSynPackage struct {
	TokenSymbol     [32]byte
	ContractAddress sdk.EthAddress
	Amount          *big.Int
	Recipient       sdk.EthAddress
	RefundAddress   sdk.AccAddress
	ExpireTime      uint64
}

func DeserializeTransferOutSynPackage(serializedPackage []byte) (*TransferOutSynPackage, error) {
	var tp TransferOutSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidPackage, "deserialize transfer out package failed")
	}
	return &tp, nil
}

type TransferOutRefundPackage struct {
	TokenSymbol  [32]byte
	RefundAmount *big.Int
	RefundAddr   sdk.AccAddress
	RefundReason RefundReason
}

func DeserializeTransferOutRefundPackage(serializedPackage []byte) (*TransferOutRefundPackage, error) {
	var tp TransferOutRefundPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidPackage, "deserialize transfer out refund package failed")
	}
	return &tp, nil
}

func SymbolToBytes(symbol string) [32]byte {
	// length of bound token symbol length should not be larger than 32
	serializedBytes := [32]byte{}
	copy(serializedBytes[:], symbol)
	return serializedBytes
}

func BytesToSymbol(symbolBytes [32]byte) string {
	tokenSymbolBytes := make([]byte, 32)
	copy(tokenSymbolBytes[:], symbolBytes[:])
	return string(bytes.Trim(tokenSymbolBytes, "\x00"))
}

type TransferInSynPackage struct {
	TokenSymbol       [32]byte
	ContractAddress   sdk.EthAddress
	Amounts           []*big.Int
	ReceiverAddresses []sdk.AccAddress
	RefundAddresses   []sdk.EthAddress
	ExpireTime        uint64
}

func DeserializeTransferInSynPackage(serializedPackage []byte) (*TransferInSynPackage, error) {
	var tp TransferInSynPackage
	err := rlp.DecodeBytes(serializedPackage, &tp)
	if err != nil {
		return nil, errors.Wrapf(ErrInvalidPackage, "deserialize transfer in package failed")

	}
	return &tp, nil
}

type TransferInRefundPackage struct {
	ContractAddr    sdk.EthAddress
	RefundAmounts   []*big.Int
	RefundAddresses []sdk.EthAddress
	RefundReason    RefundReason
}
