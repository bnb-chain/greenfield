package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type TxOption struct {
	Async     bool // default sync mode if not provided
	GasLimit  uint64
	Memo      string
	FeeAmount sdk.Coins
	FeePayer  sdk.AccAddress
}
