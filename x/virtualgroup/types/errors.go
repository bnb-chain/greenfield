package types

import (
	"cosmossdk.io/errors"
)

// x/virtualgroup module sentinel errors
var (
	ErrGVGFamilyNotExist      = errors.Register(ModuleName, 1100, "global virtual group family not exist.")
	ErrGVGNotExist            = errors.Register(ModuleName, 1101, "global virtual group not exist.")
	ErrGVGNotEmpty            = errors.Register(ModuleName, 1102, "the store size of gvg is not zero")
	ErrGenSequenceIDError     = errors.Register(ModuleName, 1103, "generate sequence id error.")
	ErrWithdrawAmountTooLarge = errors.Register(ModuleName, 1104, "withdrawal amount is too large.")

	ErrInvalidDenom = errors.Register(ModuleName, 2000, "Invalid denom.")
)
