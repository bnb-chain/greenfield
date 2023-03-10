package types

import (
	"math/big"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

const (
	Denom = "BNB"

	// DecimalBNB defines number of BNB decimal places
	DecimalBNB = 18

	// DecimalGwei defines number of gweiBNB decimal places
	DecimalGwei = 9
)

type TxOption struct {
	Mode      *tx.BroadcastMode
	GasLimit  uint64
	Memo      string
	FeeAmount sdk.Coins
	FeePayer  sdk.AccAddress
	Nonce     uint64
}

func NewIntFromInt64WithDecimal(amount int64, decimal int64) sdkmath.Int {
	return sdk.NewInt(amount).Mul(sdk.NewIntFromBigInt(new(big.Int).Exp(big.NewInt(10), big.NewInt(decimal), nil)))
}
