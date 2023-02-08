package types

import (
	"cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidExpireTime = errors.Register(ModuleName, 1, "expire time is invalid")
	ErrUnsupportedDenom  = errors.Register(ModuleName, 2, "denom is not unsupported")
	ErrInvalidAddress    = errors.Register(ModuleName, 3, "address is invalid")
	ErrInvalidPackage    = errors.Register(ModuleName, 4, "package is invalid")
	ErrInvalidAmount     = errors.Register(ModuleName, 5, "amount is invalid")
	ErrInvalidLength     = errors.Register(ModuleName, 6, "length is invalid")
	ErrPackageExpired    = errors.Register(ModuleName, 7, "package is expired")
)
