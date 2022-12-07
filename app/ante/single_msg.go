package ante

import (
	"cosmossdk.io/errors"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
)

// SingleMessageDecorator prevents tx with multi msgs from being executed
type SingleMessageDecorator struct{}

// AnteHandle rejects txs that includes more than one msgs
func (smd SingleMessageDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if msgs := tx.GetMsgs(); len(msgs) > 1 {
		return ctx, errors.Wrapf(
			sdkerrors.ErrInvalidType, "Only one msg is allowed",
		)
	}
	return next(ctx, tx, simulate)
}

// SignModeDecorator prevents tx with multi msgs from being executed
type SignModeDecorator struct {
	signModeHandler authsigning.SignModeHandler
}

func NewSignModeDecorator(signModeHandler authsigning.SignModeHandler) SignModeDecorator {
	return SignModeDecorator{
		signModeHandler: signModeHandler,
	}
}

// AnteHandle rejects txs that includes more than one msgs
func (smd SignModeDecorator) AnteHandle(ctx sdk.Context, tx sdk.Tx, simulate bool, next sdk.AnteHandler) (newCtx sdk.Context, err error) {
	if smd.signModeHandler.DefaultMode() != signing.SignMode_SIGN_MODE_EIP_712 {
		return ctx, errors.Wrapf(
			sdkerrors.ErrInvalidType, "Only EIP712 sign mode is allowed",
		)
	}
	return next(ctx, tx, simulate)
}
