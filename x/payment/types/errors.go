package types

// DONTCOVER

import (
	errorsmod "cosmossdk.io/errors"
)

// x/payment module sentinel errors
var (
	ErrReachPaymentAccountLimit    = errorsmod.Register(ModuleName, 1200, "reach payment account count limit")
	ErrInvalidState                = errorsmod.Register(ModuleName, 1201, "invalid state")
	ErrPaymentAccountNotFound      = errorsmod.Register(ModuleName, 1202, "payment account not found")
	ErrStreamRecordNotFound        = errorsmod.Register(ModuleName, 1203, "stream record not found")
	ErrNotPaymentAccountOwner      = errorsmod.Register(ModuleName, 1204, "not payment account owner")
	ErrPaymentAccountNotRefundable = errorsmod.Register(ModuleName, 1205, "payment account not refundable")
	ErrInsufficientBalance         = errorsmod.Register(ModuleName, 1206, "insufficient balance")
)
