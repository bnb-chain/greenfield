package types

// DONTCOVER

import (
	"cosmossdk.io/errors"
)

// x/bridge module sentinel errors
var (
	ErrInvalidExpireTime = errors.Register(ModuleName, 1, "expire time is invalid")
	ErrUnsupportedDenom  = errors.Register(ModuleName, 2, "denom is not unsupported")
	ErrInvalidToAddress  = errors.Register(ModuleName, 3, "to address is invalid")
	ErrInvalidPackage    = errors.Register(ModuleName, 4, "package is invalid")
)
