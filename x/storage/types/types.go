package types

import (
	sdkmath "cosmossdk.io/math"
)

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
