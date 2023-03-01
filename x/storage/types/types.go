package types

import (
	sdkmath "cosmossdk.io/math"
)

const (
	SecondarySPNum = 6
	// MinChargeSize is the minimum size to charge for a storage object
	MinChargeSize = 1024
)

// Type aliases to the SDK's math sub-module
//
// Deprecated: Functionality of this package has been moved to it's own module:
// cosmossdk.io/math
//
// Please use the above module instead of this package.
type (
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
)

func EncodeSequence(u Uint) []byte {
	return u.Bytes()
}

func DecodeSequence(bz []byte) Uint {
	u := sdkmath.NewUint(0)
	return u.SetBytes(bz)
}
