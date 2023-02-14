package ante

import (
	"fmt"

	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// CheckTxFeeWithGashubMinGasPrices check the tx fee by the gashub params.
func CheckTxFeeWithGashubMinGasPrices(ctx sdk.Context, ghk ante.GashubKeeper, tx sdk.Tx) (sdk.Coins, int64, error) {
	feeTx, ok := tx.(sdk.FeeTx)
	if !ok {
		return nil, 0, sdkerrors.Wrap(sdkerrors.ErrTxDecode, "Tx must be a FeeTx")
	}

	feeCoins := feeTx.GetFee()
	gas := feeTx.GetGas()

	// Ensure that the provided fees meet a minimum threshold for the validator,
	// if this is a CheckTx. This is only for local mempool purposes, and thus
	// is only ran on check tx.
	if ctx.IsCheckTx() {
		params := ghk.GetParams(ctx)
		minGasPriceStr := params.GetMinGasPrice()
		if minGasPriceStr != "" {
			gp, err := sdk.ParseCoinNormalized(minGasPriceStr)
			if err != nil {
				return nil, 0, fmt.Errorf("invalid min gas price: %s", minGasPriceStr)
			}
			requiredFee := sdk.NewCoin(gp.Denom, gp.Amount.Mul(sdk.NewInt(int64(gas))))

			if !feeCoins.IsAnyGTE([]sdk.Coin{requiredFee}) {
				return nil, 0, errors.Wrapf(sdkerrors.ErrInsufficientFee, "insufficient fees; got: %s required: %s", feeCoins, requiredFee)
			}
		}
	}

	priority := ante.GetTxPriority(feeCoins, int64(gas))
	return feeCoins, priority, nil
}
