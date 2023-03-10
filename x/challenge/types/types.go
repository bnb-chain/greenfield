package types

import (
	sdkmath "cosmossdk.io/math"
)

type (
	Int  = sdkmath.Int
	Uint = sdkmath.Uint
)

// RedundancyIndexPrimary defines the redundancy index for primary storage provider (asked by storage provider api)
const RedundancyIndexPrimary = int32(-1)

// BlsSignatureLength defines the length of bls signature
const BlsSignatureLength = 96
