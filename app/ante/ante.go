package ante

import (
	"fmt"
	"runtime/debug"

	sdkerrors "cosmossdk.io/errors"
	"github.com/tendermint/tendermint/libs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	errortypes "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authante "github.com/cosmos/cosmos-sdk/x/auth/ante"
)

// NewAnteHandler returns an ante handler responsible for attempting to route a
// transaction to an internal ante handler for performing transaction-level
// processing (e.g. fee payment, signature verification) before being
// passed onto its respective handler.
func NewAnteHandler(options ante.HandlerOptions) sdk.AnteHandler {
	return func(
		ctx sdk.Context, tx sdk.Tx, sim bool,
	) (newCtx sdk.Context, err error) {
		var anteHandler sdk.AnteHandler

		defer Recover(ctx.Logger(), &err)

		txWithExtensions, ok := tx.(authante.HasExtensionOptionsTx)
		if ok {
			opts := txWithExtensions.GetExtensionOptions()
			if len(opts) > 0 {
				switch typeURL := opts[0].GetTypeUrl(); typeURL {
				case "/ethermint.types.v1.ExtensionOptionsWeb3Tx":
					// handle as normal Cosmos SDK tx and the signature is checked for EIP712 representation
					anteHandler = newCosmosAnteHandlerEip712(options)
					return anteHandler(ctx, tx, sim)
				default:
					return ctx, sdkerrors.Wrapf(
						errortypes.ErrUnknownExtensionOptions,
						"rejecting tx with unsupported extension option: %s", typeURL,
					)
				}
			}
		}

		return ctx, sdkerrors.Wrapf(errortypes.ErrUnknownRequest, "invalid transaction type: %T", tx)
	}
}

// newCosmosAnteHandlerEip712 creates the ante handler for transactions signed with EIP712
func newCosmosAnteHandlerEip712(options ante.HandlerOptions) sdk.AnteHandler {
	return sdk.ChainAnteDecorators(
		RejectMessagesDecorator{}, // reject MsgEthereumTxs
		ante.NewSetUpContextDecorator(),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewConsumeGasForTxSizeDecorator(options.AccountKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		// SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewSetPubKeyDecorator(options.AccountKeeper),
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigGasConsumeDecorator(options.AccountKeeper, options.SigGasConsumer),
		// Note: signature verification uses EIP instead of the cosmos signature validator
		NewEip712SigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	)
}

func Recover(logger log.Logger, err *error) {
	if r := recover(); r != nil {
		*err = sdkerrors.Wrapf(sdkerrors.ErrPanic, "%v", r)

		if e, ok := r.(error); ok {
			logger.Error(
				"ante handler panicked",
				"error", e,
				"stack trace", string(debug.Stack()),
			)
		} else {
			logger.Error(
				"ante handler panicked",
				"recover", fmt.Sprintf("%v", r),
			)
		}
	}
}
