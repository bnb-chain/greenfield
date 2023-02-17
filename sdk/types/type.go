package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

type TxOption struct {
	Mode      *tx.BroadcastMode
	GasLimit  uint64
	Memo      string
	FeeAmount sdk.Coins
	FeePayer  sdk.AccAddress
}
