package types

import (
	"cosmossdk.io/errors"
)

// x/permission module sentinel errors
var (
	ErrInvalidPrincipal  = errors.Register(ModuleName, 1100, "Invalid principal")
	ErrInvalidStatement  = errors.Register(ModuleName, 1101, "Invalid statement")
	ErrLimitExceeded     = errors.Register(ModuleName, 1102, "Num limit exceeded")
	ErrPermissionExpired = errors.Register(ModuleName, 1103, "Permission expired")
)
