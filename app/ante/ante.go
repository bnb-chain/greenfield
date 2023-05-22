package ante

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authsigning "github.com/cosmos/cosmos-sdk/x/auth/signing"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// HandlerOptions are the options required for constructing a greenfield AnteHandler.
type HandlerOptions struct {
	AccountKeeper          ante.AccountKeeper
	BankKeeper             authtypes.BankKeeper
	ExtensionOptionChecker ante.ExtensionOptionChecker
	FeegrantKeeper         ante.FeegrantKeeper
	SignModeHandler        authsigning.SignModeHandler
	TxFeeChecker           ante.TxFeeChecker
	GashubKeeper           ante.GashubKeeper
}

// NewAnteHandler returns an AnteHandler that checks and increments sequence
// numbers, checks signatures & account numbers, and deducts fees from the first
// signer.
func NewAnteHandler(options HandlerOptions) (sdk.AnteHandler, error) {
	if options.AccountKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "account keeper is required for ante builder")
	}

	if options.BankKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "bank keeper is required for ante builder")
	}

	if options.SignModeHandler == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "sign mode handler is required for ante builder")
	}

	if options.GashubKeeper == nil {
		return nil, errors.Wrap(sdkerrors.ErrLogic, "gashub keeper is required for ante builder")
	}

	anteDecorators := []sdk.AnteDecorator{
		ante.NewSetUpContextDecorator(),               // outermost AnteDecorator. SetUpContext must be called first
		NewSignModeDecorator(options.SignModeHandler), // Only SignMode_SIGN_MODE_EIP_712 is allowed
		ante.NewExtensionOptionsDecorator(options.ExtensionOptionChecker),
		ante.NewValidateBasicDecorator(),
		ante.NewTxTimeoutHeightDecorator(),
		ante.NewValidateMemoDecorator(options.AccountKeeper),
		ante.NewValidateTxSizeDecorator(options.AccountKeeper, options.GashubKeeper),
		ante.NewConsumeMsgGasDecorator(options.AccountKeeper, options.GashubKeeper),
		ante.NewDeductFeeDecorator(options.AccountKeeper, options.BankKeeper, options.FeegrantKeeper, options.TxFeeChecker),
		ante.NewSetPubKeyDecorator(options.AccountKeeper), // SetPubKeyDecorator must be called before all signature verification decorators
		ante.NewValidateSigCountDecorator(options.AccountKeeper),
		ante.NewSigVerificationDecorator(options.AccountKeeper, options.SignModeHandler),
		ante.NewIncrementSequenceDecorator(options.AccountKeeper),
	}

	return sdk.ChainAnteDecorators(anteDecorators...), nil
}

// SignModeDecorator only allow EIP712 tx to be executed
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
	if len(smd.signModeHandler.Modes()) > 1 || smd.signModeHandler.Modes()[0] != signing.SignMode_SIGN_MODE_EIP_712 {
		return ctx, errors.Wrapf(
			sdkerrors.ErrInvalidType, "Only EIP712 sign mode is allowed",
		)
	}
	return next(ctx, tx, simulate)
}
