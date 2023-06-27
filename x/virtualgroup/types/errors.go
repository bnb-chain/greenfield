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
	ErrInvalidBlsPubKey            = errors.Register(ModuleName, 1122, "invalid bls public key")
	ErrLimitationExceed            = errors.Register(ModuleName, 1123, "limitation exceed.")
	ErrRebindingGVGsToBucketFailed = errors.Register(ModuleName, 1124, "rebinding gvgs to bucket failed.")
	ErrInsufficientStaking         = errors.Register(ModuleName, 1125, "insufficient staking for gvg")

	ErrInvalidDenom = errors.Register(ModuleName, 2000, "Invalid denom.")
)
