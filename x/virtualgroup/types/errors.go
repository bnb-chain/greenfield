package types

import (
	"cosmossdk.io/errors"
)

// x/virtualgroup module sentinel errors
var (
	ErrGVGFamilyNotExist      = errors.Register(ModuleName, 1100, "global virtual group family not exist.")
	ErrGVGNotExistInFamily    = errors.Register(ModuleName, 1101, "global virtual group not exist in family.")
	ErrGVGNotExist            = errors.Register(ModuleName, 1102, "global virtual group not exist.")
	ErrGVGNotEmpty            = errors.Register(ModuleName, 1103, "the store size of gvg is not zero")
	ErrGenSequenceIDError     = errors.Register(ModuleName, 1104, "generate sequence id error.")
	ErrWithdrawAmountTooLarge = errors.Register(ModuleName, 1105, "withdrawal amount is too large.")

	ErrLVGNotExist = errors.Register(ModuleName, 1106, "local virtual group not exist.")

	ErrInvalidDenom = errors.Register(ModuleName, 2000, "Invalid denom.")
)
