package types

import (
	"fmt"
	"math/big"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// storage params default values
const (
	DefaultMaxSegmentSize           uint64 = 16 * 1024 * 1024 // 16M
	DefaultRedundantDataChunkNum    uint32 = 4
	DefaultRedundantParityChunkNum  uint32 = 2
	DefaultMaxPayloadSize           uint64 = 2 * 1024 * 1024 * 1024
	DefaultMaxBucketsPerAccount     uint32 = 100
	DefaultMinChargeSize            uint64 = 1 * 1024 * 1024 // 1M
	DefaultDiscontinueRequestWindow uint64 = 10000
	DefaultDiscontinueRequestMax    uint64 = 128
	DefaultDiscontinueConfirmPeriod uint64 = 100

	DefaultMirrorBucketRelayerFee    = "1000000000000000" // 0.01
	DefaultMirrorBucketAckRelayerFee = "0"
	DefaultMirrorObjectRelayerFee    = "1000000000000000" // 0.01
	DefaultMirrorObjectAckRelayerFee = "0"
	DefaultMirrorGroupRelayerFee     = "1000000000000000" // 0.01
	DefaultMirrorGroupAckRelayerFee  = "0"
)

var (
	KeyMaxSegmentSize           = []byte("MaxSegmentSize")
	KeyRedundantDataChunkNum    = []byte("RedundantDataChunkNum")
	KeyRedundantParityChunkNum  = []byte("RedundantParityChunkNum")
	KeyMaxPayloadSize           = []byte("MaxPayloadSize")
	KeyMinChargeSize            = []byte("MinChargeSize")
	KeyMaxBucketsPerAccount     = []byte("MaxBucketsPerAccount")
	KeyDiscontinueRequestWindow = []byte("DiscontinueRequestWindow")
	KeyDiscontinueRequestMax    = []byte("DiscontinueRequestMax")
	KeyDiscontinueConfirmPeriod = []byte("DiscontinueConfirmPeriod")

	KeyMirrorBucketRelayerFee    = []byte("MirrorBucketRelayerFee")
	KeyMirrorBucketAckRelayerFee = []byte("MirrorBucketAckRelayerFee")
	KeyMirrorObjectRelayerFee    = []byte("MirrorObjectRelayerFee")
	KeyMirrorObjectAckRelayerFee = []byte("MirrorObjectAckRelayerFee")
	KeyMirrorGroupRelayerFee     = []byte("MirrorGroupRelayerFee")
	KeyMirrorGroupAckRelayerFee  = []byte("MirrorGroupAckRelayerFee")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	maxSegmentSize uint64, redundantDataChunkNum uint32,
	redundantParityChunkNum uint32, maxPayloadSize uint64, maxBucketsPerAccount uint32,
	minChargeSize uint64, mirrorBucketRelayerFee, mirrorBucketAckRelayerFee string,
	mirrorObjectRelayerFee, mirrorObjectAckRelayerFee string,
	mirrorGroupRelayerFee, mirrorGroupAckRelayerFee string,
	discontinueRequestWindow, discontinueRequestMax, discontinueConfirmPeriod uint64,
) Params {
	return Params{
		MaxSegmentSize:            maxSegmentSize,
		RedundantDataChunkNum:     redundantDataChunkNum,
		RedundantParityChunkNum:   redundantParityChunkNum,
		MaxPayloadSize:            maxPayloadSize,
		MinChargeSize:             minChargeSize,
		MaxBucketsPerAccount:      maxBucketsPerAccount,
		MirrorBucketRelayerFee:    mirrorBucketRelayerFee,
		MirrorBucketAckRelayerFee: mirrorBucketAckRelayerFee,
		MirrorObjectRelayerFee:    mirrorObjectRelayerFee,
		MirrorObjectAckRelayerFee: mirrorObjectAckRelayerFee,
		MirrorGroupRelayerFee:     mirrorGroupRelayerFee,
		MirrorGroupAckRelayerFee:  mirrorGroupAckRelayerFee,
		DiscontinueRequestWindow:  discontinueRequestWindow,
		DiscontinueRequestMax:     discontinueRequestMax,
		DiscontinueConfirmPeriod:  discontinueConfirmPeriod,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultMaxSegmentSize, DefaultRedundantDataChunkNum,
		DefaultRedundantParityChunkNum, DefaultMaxPayloadSize, DefaultMaxBucketsPerAccount,
		DefaultMinChargeSize, DefaultMirrorBucketRelayerFee, DefaultMirrorBucketAckRelayerFee,
		DefaultMirrorObjectRelayerFee, DefaultMirrorObjectAckRelayerFee,
		DefaultMirrorGroupRelayerFee, DefaultMirrorGroupAckRelayerFee,
		DefaultDiscontinueRequestWindow, DefaultDiscontinueRequestMax, DefaultDiscontinueConfirmPeriod,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxSegmentSize, &p.MaxSegmentSize, validateMaxSegmentSize),
		paramtypes.NewParamSetPair(KeyRedundantDataChunkNum, &p.RedundantDataChunkNum, validateRedundantDataChunkNum),
		paramtypes.NewParamSetPair(KeyRedundantParityChunkNum, &p.RedundantParityChunkNum, validateRedundantParityChunkNum),
		paramtypes.NewParamSetPair(KeyMaxPayloadSize, &p.MaxPayloadSize, validateMaxPayloadSize),
		paramtypes.NewParamSetPair(KeyMaxBucketsPerAccount, &p.MaxBucketsPerAccount, validateMaxBucketsPerAccount),
		paramtypes.NewParamSetPair(KeyMinChargeSize, &p.MinChargeSize, validateMinChargeSize),
		paramtypes.NewParamSetPair(KeyMirrorBucketRelayerFee, &p.MirrorBucketRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyMirrorBucketAckRelayerFee, &p.MirrorBucketAckRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyMirrorObjectRelayerFee, &p.MirrorObjectRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyMirrorObjectAckRelayerFee, &p.MirrorObjectAckRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyMirrorGroupRelayerFee, &p.MirrorGroupRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyMirrorGroupAckRelayerFee, &p.MirrorGroupAckRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyDiscontinueRequestWindow, &p.DiscontinueRequestWindow, validateDiscontinueRequestWindow),
		paramtypes.NewParamSetPair(KeyDiscontinueRequestMax, &p.DiscontinueRequestMax, validateDiscontinueRequestMax),
		paramtypes.NewParamSetPair(KeyDiscontinueConfirmPeriod, &p.DiscontinueConfirmPeriod, validateDiscontinueConfirmPeriod),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMaxSegmentSize(p.MaxSegmentSize); err != nil {
		return err
	}
	if err := validateRedundantDataChunkNum(p.RedundantDataChunkNum); err != nil {
		return err
	}
	if err := validateRedundantParityChunkNum(p.RedundantParityChunkNum); err != nil {
		return err
	}
	if err := validateMaxPayloadSize(p.MaxPayloadSize); err != nil {
		return err
	}
	if err := validateMaxBucketsPerAccount(p.MaxBucketsPerAccount); err != nil {
		return err
	}
	if err := validateMinChargeSize(p.MinChargeSize); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMaxSegmentSize(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max segment size must be positive: %d", v)
	}

	return nil
}

func validateMaxPayloadSize(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max payload size must be positive: %d", v)
	}

	return nil
}

func validateMaxBucketsPerAccount(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max buckets per account must be positive: %d", v)
	}

	return nil
}

func validateMinChargeSize(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("min charge size must be positive: %d", v)
	}

	return nil
}

func validateRedundantDataChunkNum(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("redundant data chunk num must be positive: %d", v)
	}

	return nil
}
func validateRedundantParityChunkNum(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("redundant parity size chunk num must be positive: %d", v)
	}

	return nil
}

func validateRelayerFee(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	relayerFee := big.NewInt(0)
	relayerFee, valid := relayerFee.SetString(v, 10)

	if !valid {
		return fmt.Errorf("invalid transfer out relayer fee, %s", v)
	}

	if relayerFee.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("invalid transfer out relayer fee, %s", v)
	}

	return nil
}

func validateDiscontinueRequestWindow(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("discontine request window must be positive: %d", v)
	}

	return nil
}

func validateDiscontinueRequestMax(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDiscontinueConfirmPeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("discontine confirm period must be positive: %d", v)
	}
	return nil
}
