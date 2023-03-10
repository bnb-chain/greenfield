package types

import (
	"cosmossdk.io/errors"
)

// x/gensptx module sentinel errors
var (
	ErrSample = errors.Register(ModuleName, 1100, "sample error")
)
