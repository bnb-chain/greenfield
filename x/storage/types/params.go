package types

import (
	"fmt"
	"math"
	"math/big"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// storage params default values
const (
	DefaultMaxSegmentSize            uint64 = 16 * 1024 * 1024 // 16M
	DefaultRedundantDataChunkNum     uint32 = 4
	DefaultRedundantParityChunkNum   uint32 = 2
	DefaultMaxPayloadSize            uint64 = 64 * 1024 * 1024 * 1024
	DefaultMaxBucketsPerAccount      uint32 = 100
	DefaultMinChargeSize             uint64 = 1 * 1024 * 1024 // 1M
	DefaultDiscontinueCountingWindow uint64 = 10000
	DefaultDiscontinueObjectMax      uint64 = math.MaxUint64
	DefaultDiscontinueBucketMax      uint64 = math.MaxUint64
	DefaultDiscontinueConfirmPeriod  int64  = 604800 // 7 days (in second)
	DefaultDiscontinueDeletionMax    uint64 = 100
	DefaultStalePolicyCleanupMax     uint64 = 200
	DefaultMinUpdateQuotaInterval    uint64 = 2592000 // 30 days (in second)

	DefaultMaxLocalVirtualGroupNumPerBucket uint32 = 10
	DefaultMirrorBucketRelayerFee                  = "1300000000000000" // 0.0013
	DefaultMirrorBucketAckRelayerFee               = "250000000000000"  // 0.00025
	DefaultMirrorObjectRelayerFee                  = "1300000000000000" // 0.0013
	DefaultMirrorObjectAckRelayerFee               = "250000000000000"  // 0.00025
	DefaultMirrorGroupRelayerFee                   = "1300000000000000" // 0.0013
	DefaultMirrorGroupAckRelayerFee                = "250000000000000"  // 0.00025
)

var (
	KeyMaxSegmentSize                   = []byte("MaxSegmentSize")
	KeyRedundantDataChunkNum            = []byte("RedundantDataChunkNum")
	KeyRedundantParityChunkNum          = []byte("RedundantParityChunkNum")
	KeyMaxPayloadSize                   = []byte("MaxPayloadSize")
	KeyMinChargeSize                    = []byte("MinChargeSize")
	KeyMaxBucketsPerAccount             = []byte("MaxBucketsPerAccount")
	KeyDiscontinueCountingWindow        = []byte("DiscontinueCountingWindow")
	KeyDiscontinueObjectMax             = []byte("DiscontinueObjectMax")
	KeyDiscontinueBucketMax             = []byte("DiscontinueBucketMax")
	KeyDiscontinueConfirmPeriod         = []byte("DiscontinueConfirmPeriod")
	KeyDiscontinueDeletionMax           = []byte("DiscontinueDeletionMax")
	KeyStalePolicyCleanupMax            = []byte("StalePolicyCleanupMax")
	KeyMinUpdateQuotaInterval           = []byte("MinUpdateQuotaInterval")
	KeyBscMirrorBucketRelayerFee        = []byte("BscMirrorBucketRelayerFee")
	KeyBscMirrorBucketAckRelayerFee     = []byte("BscMirrorBucketAckRelayerFee")
	KeyBscMirrorObjectRelayerFee        = []byte("BscMirrorObjectRelayerFee")
	KeyBscMirrorObjectAckRelayerFee     = []byte("BscMirrorObjectAckRelayerFee")
	KeyBscMirrorGroupRelayerFee         = []byte("BscMirrorGroupRelayerFee")
	KeyBscMirrorGroupAckRelayerFee      = []byte("BscMirrorGroupAckRelayerFee")
	KeyMaxLocalVirtualGroupNumPerBucket = []byte("MaxLocalVirtualGroupNumPerBucket")
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
	discontinueCountingWindow, discontinueObjectMax, discontinueBucketMax uint64,
	discontinueConfirmPeriod int64,
	discontinueDeletionMax uint64,
	stalePoliesCleanupMax uint64,
	minUpdateQuotaInterval uint64,
	maxLocalVirtualGroupNumPerBucket uint32,
) Params {
	return Params{
		VersionedParams: VersionedParams{
			MaxSegmentSize:          maxSegmentSize,
			RedundantDataChunkNum:   redundantDataChunkNum,
			RedundantParityChunkNum: redundantParityChunkNum,
			MinChargeSize:           minChargeSize,
		},
		MaxPayloadSize:                   maxPayloadSize,
		MaxBucketsPerAccount:             maxBucketsPerAccount,
		BscMirrorBucketRelayerFee:        mirrorBucketRelayerFee,
		BscMirrorBucketAckRelayerFee:     mirrorBucketAckRelayerFee,
		BscMirrorObjectRelayerFee:        mirrorObjectRelayerFee,
		BscMirrorObjectAckRelayerFee:     mirrorObjectAckRelayerFee,
		BscMirrorGroupRelayerFee:         mirrorGroupRelayerFee,
		BscMirrorGroupAckRelayerFee:      mirrorGroupAckRelayerFee,
		DiscontinueCountingWindow:        discontinueCountingWindow,
		DiscontinueObjectMax:             discontinueObjectMax,
		DiscontinueBucketMax:             discontinueBucketMax,
		DiscontinueConfirmPeriod:         discontinueConfirmPeriod,
		DiscontinueDeletionMax:           discontinueDeletionMax,
		StalePolicyCleanupMax:            stalePoliesCleanupMax,
		MinQuotaUpdateInterval:           minUpdateQuotaInterval,
		MaxLocalVirtualGroupNumPerBucket: maxLocalVirtualGroupNumPerBucket,
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
		DefaultDiscontinueCountingWindow, DefaultDiscontinueObjectMax, DefaultDiscontinueBucketMax,
		DefaultDiscontinueConfirmPeriod, DefaultDiscontinueDeletionMax, DefaultStalePolicyCleanupMax,
		DefaultMinUpdateQuotaInterval, DefaultMaxLocalVirtualGroupNumPerBucket,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxSegmentSize, &p.VersionedParams.MaxSegmentSize, validateMaxSegmentSize),
		paramtypes.NewParamSetPair(KeyRedundantDataChunkNum, &p.VersionedParams.RedundantDataChunkNum, validateRedundantDataChunkNum),
		paramtypes.NewParamSetPair(KeyRedundantParityChunkNum, &p.VersionedParams.RedundantParityChunkNum, validateRedundantParityChunkNum),
		paramtypes.NewParamSetPair(KeyMinChargeSize, &p.VersionedParams.MinChargeSize, validateMinChargeSize),

		paramtypes.NewParamSetPair(KeyMaxPayloadSize, &p.MaxPayloadSize, validateMaxPayloadSize),
		paramtypes.NewParamSetPair(KeyMaxBucketsPerAccount, &p.MaxBucketsPerAccount, validateMaxBucketsPerAccount),
		paramtypes.NewParamSetPair(KeyBscMirrorBucketRelayerFee, &p.BscMirrorBucketRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyBscMirrorBucketAckRelayerFee, &p.BscMirrorBucketAckRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyBscMirrorObjectRelayerFee, &p.BscMirrorObjectRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyBscMirrorObjectAckRelayerFee, &p.BscMirrorObjectAckRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyBscMirrorGroupRelayerFee, &p.BscMirrorGroupRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyBscMirrorGroupAckRelayerFee, &p.BscMirrorGroupAckRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyDiscontinueCountingWindow, &p.DiscontinueCountingWindow, validateDiscontinueCountingWindow),
		paramtypes.NewParamSetPair(KeyDiscontinueObjectMax, &p.DiscontinueObjectMax, validateDiscontinueObjectMax),
		paramtypes.NewParamSetPair(KeyDiscontinueBucketMax, &p.DiscontinueBucketMax, validateDiscontinueBucketMax),
		paramtypes.NewParamSetPair(KeyDiscontinueConfirmPeriod, &p.DiscontinueConfirmPeriod, validateDiscontinueConfirmPeriod),
		paramtypes.NewParamSetPair(KeyDiscontinueDeletionMax, &p.DiscontinueDeletionMax, validateDiscontinueDeletionMax),
		paramtypes.NewParamSetPair(KeyStalePolicyCleanupMax, &p.StalePolicyCleanupMax, validateStalePolicyCleanupMax),
		paramtypes.NewParamSetPair(KeyMinUpdateQuotaInterval, &p.MinQuotaUpdateInterval, validateMinUpdateQuotaInterval),
		paramtypes.NewParamSetPair(KeyMaxLocalVirtualGroupNumPerBucket, &p.MaxLocalVirtualGroupNumPerBucket, validateMaxLocalVirtualGroupNumPerBucket),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMaxSegmentSize(p.VersionedParams.MaxSegmentSize); err != nil {
		return err
	}
	if err := validateRedundantDataChunkNum(p.VersionedParams.RedundantDataChunkNum); err != nil {
		return err
	}
	if err := validateRedundantParityChunkNum(p.VersionedParams.RedundantParityChunkNum); err != nil {
		return err
	}
	if err := validateMinChargeSize(p.VersionedParams.MinChargeSize); err != nil {
		return err
	}
	if err := validateMaxPayloadSize(p.MaxPayloadSize); err != nil {
		return err
	}
	if err := validateMaxBucketsPerAccount(p.MaxBucketsPerAccount); err != nil {
		return err
	}
	if err := validateRelayerFee(p.BscMirrorBucketRelayerFee); err != nil {
		return err
	}
	if err := validateRelayerFee(p.BscMirrorBucketAckRelayerFee); err != nil {
		return err
	}
	if err := validateRelayerFee(p.BscMirrorObjectRelayerFee); err != nil {
		return err
	}
	if err := validateRelayerFee(p.BscMirrorObjectAckRelayerFee); err != nil {
		return err
	}
	if err := validateRelayerFee(p.BscMirrorGroupRelayerFee); err != nil {
		return err
	}
	if err := validateRelayerFee(p.BscMirrorGroupAckRelayerFee); err != nil {
		return err
	}
	if err := validateDiscontinueCountingWindow(p.DiscontinueCountingWindow); err != nil {
		return err
	}
	if err := validateDiscontinueObjectMax(p.DiscontinueObjectMax); err != nil {
		return err
	}
	if err := validateDiscontinueBucketMax(p.DiscontinueBucketMax); err != nil {
		return err
	}
	if err := validateDiscontinueConfirmPeriod(p.DiscontinueConfirmPeriod); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// String implements the Stringer interface.
func (p VersionedParams) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func (p *Params) GetMaxSegmentSize() uint64 {
	if p != nil {
		return p.VersionedParams.MaxSegmentSize
	}
	return 0
}

func (p *Params) GetRedundantDataChunkNum() uint32 {
	if p != nil {
		return p.VersionedParams.RedundantDataChunkNum
	}
	return 0
}

func (p *Params) GetRedundantParityChunkNum() uint32 {
	if p != nil {
		return p.VersionedParams.RedundantParityChunkNum
	}
	return 0
}

func (p *Params) GetMinChargeSize() uint64 {
	if p != nil {
		return p.VersionedParams.MinChargeSize
	}
	return 0
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

func validateDiscontinueCountingWindow(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("discontinue counting window must be positive: %d", v)
	}

	return nil
}

func validateDiscontinueObjectMax(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDiscontinueBucketMax(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateDiscontinueConfirmPeriod(i interface{}) error {
	v, ok := i.(int64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v <= 0 {
		return fmt.Errorf("discontinue confirm period must be positive: %d", v)
	}
	return nil
}

func validateDiscontinueDeletionMax(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("discontinue deletion max must be positive: %d", v)
	}
	return nil
}

func validateStalePolicyCleanupMax(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("StalePolicyCleanupMax must be positive: %d", v)
	}
	return nil
}

func validateMinUpdateQuotaInterval(i interface{}) error {
	_, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	return nil
}

func validateMaxLocalVirtualGroupNumPerBucket(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max buckets per account must be positive: %d", v)
	}

	return nil
}
