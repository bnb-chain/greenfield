package types

import (
	sdkmath "cosmossdk.io/math"
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

func MustMarshalUint(n sdkmath.Uint) []byte {
	nb, err := n.Marshal()
	if err != nil {
		panic(err)
	}
	return nb
}

func MustUnmarshalUint(data []byte) sdkmath.Uint {
	n := sdkmath.ZeroUint()
	err := n.Unmarshal(data)
	if err != nil {
		panic(err)
	}
	return n
}
