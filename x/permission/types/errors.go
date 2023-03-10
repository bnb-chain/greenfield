package types

import (
	"cosmossdk.io/errors"
)

// x/permission module sentinel errors
var (
	ErrInvalidPrincipal    = errors.Register(ModuleName, 1100, "Invalid principal")
	ErrInvalidStatement    = errors.Register(ModuleName, 1101, "Invalid statement")
	ErrInvalidPolicy       = errors.Register(ModuleName, 1102, "Invalid policy")
	ErrInvalidResourceType = errors.Register(ModuleName, 1103, "Invalid resource type")
	ErrLimitExceeded       = errors.Register(ModuleName, 1105, "Num limit exceeded")
)
