package types

import (
	errorsmod "cosmossdk.io/errors"
)

// x/payment module sentinel errors
var (
	ErrReachPaymentAccountLimit           = errorsmod.Register(ModuleName, 1200, "reach payment account count limit")
	ErrInvalidState                       = errorsmod.Register(ModuleName, 1201, "invalid state")
	ErrPaymentAccountNotFound             = errorsmod.Register(ModuleName, 1202, "payment account not found")
	ErrStreamRecordNotFound               = errorsmod.Register(ModuleName, 1203, "stream record not found")
	ErrNotPaymentAccountOwner             = errorsmod.Register(ModuleName, 1204, "not payment account owner")
	ErrPaymentAccountAlreadyNonRefundable = errorsmod.Register(ModuleName, 1205, "payment account has already be set as non-refundable")
	ErrInsufficientBalance                = errorsmod.Register(ModuleName, 1206, "insufficient balance")
	ErrReceiveAccountNotExist             = errorsmod.Register(ModuleName, 1207, "receive account not exist")
	ErrInvalidStreamAccountStatus         = errorsmod.Register(ModuleName, 1208, "invalid stream account status")
	ErrInvalidParams                      = errorsmod.Register(ModuleName, 1209, "invalid params")
)
