package types

import (
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// storage params default values
const (
	DefaultMaxSegmentSize          uint64 = 16 * 1024 * 1024 // 16M
	DefaultRedundantDataChunkNum   uint32 = 4
	DefaultRedundantParityChunkNum uint32 = 2
	DefaultMaxPayloadSize          uint64 = 2 * 1024 * 1024 * 1024
)

var (
	KeyMaxSegmentSize          = []byte("MaxSegmentSize")
	KeyRedundantDataChunkNum   = []byte("RedundantDataChunkNum")
	KeyRedundantParityChunkNum = []byte("RedundantParityChunkNum")
	KeyMaxPayloadSize          = []byte("MaxPayloadSize")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxSegmentSize uint64, redundantDataChunkNum uint32,
	redundantParityChunkNum uint32, maxPayloadSize uint64) Params {
	return Params{
		MaxSegmentSize:          maxSegmentSize,
		RedundantDataChunkNum:   redundantDataChunkNum,
		RedundantParityChunkNum: redundantParityChunkNum,
		MaxPayloadSize:          maxPayloadSize,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxSegmentSize, DefaultRedundantDataChunkNum,
		DefaultRedundantParityChunkNum, DefaultMaxPayloadSize)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxSegmentSize, &p.MaxSegmentSize, nil),
		paramtypes.NewParamSetPair(KeyRedundantDataChunkNum, &p.RedundantDataChunkNum, nil),
		paramtypes.NewParamSetPair(KeyRedundantParityChunkNum, &p.RedundantParityChunkNum, nil),
		paramtypes.NewParamSetPair(KeyMaxPayloadSize, &p.MaxPayloadSize, nil),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	// TODO: should validate the params
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}
