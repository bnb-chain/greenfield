package types

import (
	"math/big"
	"time"

	"github.com/cosmos/cosmos-sdk/bsc"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	MinTransferOutExpireTimeGap = 60 * time.Second

	TransferOutChannel = "transferOut"

	TransferOutChannelID sdk.ChannelID = 1
)

var CrossTransferOutRelayFee = sdk.NewInt(1) // TODO: to be determined

type TransferOutSynPackage struct {
	TokenSymbol     [32]byte
	ContractAddress bsc.SmartChainAddress
	Amount          *big.Int
	Recipient       bsc.SmartChainAddress
	RefundAddress   sdk.AccAddress
	ExpireTime      uint64
}
