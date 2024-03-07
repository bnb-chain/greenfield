package types

import (
	"cosmossdk.io/errors"
)

// x/virtualgroup module sentinel errors
var (
	ErrGVGFamilyNotExist           = errors.Register(ModuleName, 1100, "global virtual group family not exist.")
	ErrGVGNotExistInFamily         = errors.Register(ModuleName, 1101, "global virtual group not exist in family.")
	ErrGVGNotExist                 = errors.Register(ModuleName, 1102, "global virtual group not exist.")
	ErrGVGNotEmpty                 = errors.Register(ModuleName, 1103, "the store size of gvg is not zero")
	ErrGenSequenceIDError          = errors.Register(ModuleName, 1104, "generate sequence id error.")
	ErrWithdrawAmountTooLarge      = errors.Register(ModuleName, 1105, "withdrawal amount is too large.")
	ErrSwapOutFailed               = errors.Register(ModuleName, 1106, "swap out failed.")
	ErrLVGNotExist                 = errors.Register(ModuleName, 1107, "local virtual group not exist.")
	ErrSPCanNotExit                = errors.Register(ModuleName, 1108, "the sp can not exit now.")
	ErrSettleFailed                = errors.Register(ModuleName, 1109, "fail to settle.")
	ErrInvalidGVGCount             = errors.Register(ModuleName, 1120, "the count of global virtual group ids is invalid.")
	ErrWithdrawFailed              = errors.Register(ModuleName, 1121, "with draw failed.")
	ErrInvalidSecondarySPCount     = errors.Register(ModuleName, 1122, "the number of secondary sp within the global virtual group is invalid.")
	ErrLimitationExceed            = errors.Register(ModuleName, 1123, "limitation exceed.")
	ErrDuplicateSecondarySP        = errors.Register(ModuleName, 1124, "the global virtual group has duplicate secondary sp.")
	ErrInsufficientStaking         = errors.Register(ModuleName, 1125, "insufficient staking for gvg")
	ErrDuplicateGVG                = errors.Register(ModuleName, 1126, "global virtual group is duplicate")
	ErrSwapInFailed                = errors.Register(ModuleName, 1127, "swap in failed.")
	ErrSwapInInfoNotExist          = errors.Register(ModuleName, 1128, "swap in info not exist.")
	ErrGVGStatisticsNotExist       = errors.Register(ModuleName, 1129, "global virtual group statistics not exist.")
	ErrGVGFamilyStatisticsNotExist = errors.Register(ModuleName, 1130, "global virtual group family statistics not exist.")

	ErrInvalidDenom = errors.Register(ModuleName, 2000, "Invalid denom.")
)
