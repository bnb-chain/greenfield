package types

import (
	"errors"
	"fmt"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyChallengeCountPerBlock            = []byte("ChallengeCountPerBlock")
	DefaultChallengeCountPerBlock uint64 = 1
)

var (
	KeyChallengeKeepAlivePeriod            = []byte("ChallengeKeepAlivePeriod")
	DefaultChallengeKeepAlivePeriod uint64 = 300
)

var (
	KeySlashCoolingOffPeriod            = []byte("SlashCoolingOffPeriod")
	DefaultSlashCoolingOffPeriod uint64 = 300
)

var (
	KeySlashAmountSizeRate     = []byte("SlashAmountSizeRate")
	DefaultSlashAmountSizeRate = sdk.NewDecWithPrec(85, 4)
)

var (
	KeySlashAmountMin     = []byte("SlashAmountMin")
	DefaultSlashAmountMin = math.NewIntFromBigInt(big.NewInt(1e16))
)

var (
	KeySlashAmountMax     = []byte("SlashAmountMax")
	DefaultSlashAmountMax = math.NewIntFromBigInt(big.NewInt(1e18))
)

var (
	KeyRewardValidatorRatio     = []byte("RewardValidatorRatio")
	DefaultRewardValidatorRatio = sdk.NewDecWithPrec(9, 1)
)

var (
	KeyRewardSubmitterRatio     = []byte("RewardSubmitterRatio")
	DefaultRewardSubmitterRatio = sdk.NewDecWithPrec(1, 3)
)

var (
	KeyRewardSubmitterThreshold     = []byte("RewardSubmitterThreshold")
	DefaultRewardSubmitterThreshold = math.NewIntFromBigInt(big.NewInt(1e15))
)

var (
	KeyHeartbeatInterval            = []byte("HeartbeatInterval")
	DefaultHeartbeatInterval uint64 = 1000
)

var (
	KeyAttestationInturnInterval            = []byte("AttestationInturnInterval")
	DefaultAttestationInturnInterval uint64 = 120 // in seconds
)

var (
	KeyAttestationKeptCount            = []byte("AttestationKeptCount")
	DefaultAttestationKeptCount uint64 = 10
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	challengeCountPerBlock uint64,
	challengeKeepAlivePeriod uint64,
	slashCoolingOffPeriod uint64,
	slashAmountSizeRate sdk.Dec,
	slashAmountMin math.Int,
	slashAmountMax math.Int,
	rewardValidatorRatio sdk.Dec,
	rewardSubmitterRatio sdk.Dec,
	rewardSubmitterThreshold math.Int,
	heartbeatInterval uint64,
	attestationInturnInterval uint64,
	attestationKeptCount uint64,
) Params {
	return Params{
		ChallengeCountPerBlock:    challengeCountPerBlock,
		ChallengeKeepAlivePeriod:  challengeKeepAlivePeriod,
		SlashCoolingOffPeriod:     slashCoolingOffPeriod,
		SlashAmountSizeRate:       slashAmountSizeRate,
		SlashAmountMin:            slashAmountMin,
		SlashAmountMax:            slashAmountMax,
		RewardValidatorRatio:      rewardValidatorRatio,
		RewardSubmitterRatio:      rewardSubmitterRatio,
		RewardSubmitterThreshold:  rewardSubmitterThreshold,
		HeartbeatInterval:         heartbeatInterval,
		AttestationInturnInterval: attestationInturnInterval,
		AttestationKeptCount:      attestationKeptCount,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultChallengeCountPerBlock,
		DefaultChallengeKeepAlivePeriod,
		DefaultSlashCoolingOffPeriod,
		DefaultSlashAmountSizeRate,
		DefaultSlashAmountMin,
		DefaultSlashAmountMax,
		DefaultRewardValidatorRatio,
		DefaultRewardSubmitterRatio,
		DefaultRewardSubmitterThreshold,
		DefaultHeartbeatInterval,
		DefaultAttestationInturnInterval,
		DefaultAttestationKeptCount,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyChallengeCountPerBlock, &p.ChallengeCountPerBlock, validateChallengeCountPerBlock),
		paramtypes.NewParamSetPair(KeyChallengeKeepAlivePeriod, &p.ChallengeKeepAlivePeriod, validateChallengeKeepAlivePeriod),
		paramtypes.NewParamSetPair(KeySlashCoolingOffPeriod, &p.SlashCoolingOffPeriod, validateSlashCoolingOffPeriod),
		paramtypes.NewParamSetPair(KeySlashAmountSizeRate, &p.SlashAmountSizeRate, validateSlashAmountSizeRate),
		paramtypes.NewParamSetPair(KeySlashAmountMin, &p.SlashAmountMin, validateSlashAmountMin),
		paramtypes.NewParamSetPair(KeySlashAmountMax, &p.SlashAmountMax, validateSlashAmountMax),
		paramtypes.NewParamSetPair(KeyRewardValidatorRatio, &p.RewardValidatorRatio, validateRewardValidatorRatio),
		paramtypes.NewParamSetPair(KeyRewardSubmitterRatio, &p.RewardSubmitterRatio, validateRewardSubmitterRatio),
		paramtypes.NewParamSetPair(KeyRewardSubmitterThreshold, &p.RewardSubmitterThreshold, validateRewardSubmitterThreshold),
		paramtypes.NewParamSetPair(KeyHeartbeatInterval, &p.HeartbeatInterval, validateHeartbeatInterval),
		paramtypes.NewParamSetPair(KeyAttestationInturnInterval, &p.AttestationInturnInterval, validateAttestationInturnInterval),
		paramtypes.NewParamSetPair(KeyAttestationKeptCount, &p.AttestationKeptCount, validateAttestationKeptCount),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateChallengeCountPerBlock(p.ChallengeCountPerBlock); err != nil {
		return err
	}

	if err := validateChallengeKeepAlivePeriod(p.ChallengeKeepAlivePeriod); err != nil {
		return err
	}

	if err := validateSlashCoolingOffPeriod(p.SlashCoolingOffPeriod); err != nil {
		return err
	}

	if err := validateSlashAmountSizeRate(p.SlashAmountSizeRate); err != nil {
		return err
	}

	if err := validateSlashAmountMin(p.SlashAmountMin); err != nil {
		return err
	}

	if err := validateSlashAmountMax(p.SlashAmountMax); err != nil {
		return err
	}

	if err := validateRewardValidatorRatio(p.RewardValidatorRatio); err != nil {
		return err
	}

	if err := validateRewardSubmitterRatio(p.RewardSubmitterRatio); err != nil {
		return err
	}

	if err := validateRewardSubmitterThreshold(p.RewardSubmitterThreshold); err != nil {
		return err
	}

	if p.SlashAmountMax.LTE(p.SlashAmountMin) {
		return errors.New("max slash amount should be bigger than min slash amount")
	}

	if p.RewardValidatorRatio.Add(p.RewardSubmitterRatio).GT(sdk.NewDec(1)) {
		return errors.New("the sum of validator and challenger reward ratio should be equal to or less than one")
	}

	if err := validateHeartbeatInterval(p.HeartbeatInterval); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateChallengeCountPerBlock validates the ChallengeCountPerBlock param
func validateChallengeCountPerBlock(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return nil
}

// validateChallengeKeepAlivePeriod validates the ChallengeKeepAlivePeriod param
func validateChallengeKeepAlivePeriod(v interface{}) error {
	challengeKeepAlivePeriod, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if challengeKeepAlivePeriod == 0 {
		return errors.New("keep alive period cannot be zero")
	}

	return nil
}

// validateSlashCoolingOffPeriod validates the SlashCoolingOffPeriod param
func validateSlashCoolingOffPeriod(v interface{}) error {
	_, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	return nil
}

// validateSlashAmountSizeRate validates the SlashAmountPerSizeRate param
func validateSlashAmountSizeRate(v interface{}) error {
	slashAmountSizeRate, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if slashAmountSizeRate.LT(sdk.ZeroDec()) {
		return errors.New("slash amount size rate cannot be lower than zero")
	}

	return nil
}

// validateSlashAmountMin validates the SlashAmountMin param
func validateSlashAmountMin(v interface{}) error {
	slashAmountMin, ok := v.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if slashAmountMin.LT(sdk.ZeroInt()) {
		return errors.New("min slash amount cannot be lower than zero")
	}

	return nil
}

// validateSlashAmountMax validates the SlashAmountMax param
func validateSlashAmountMax(v interface{}) error {
	slashAmountMax, ok := v.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if slashAmountMax.LT(sdk.ZeroInt()) {
		return errors.New("max slash amount cannot be lower than zero")
	}

	return nil
}

// validateRewardValidatorRatio validates the RewardValidatorRatio param
func validateRewardValidatorRatio(v interface{}) error {
	rewardValidatorRatio, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if rewardValidatorRatio.LT(sdk.ZeroDec()) {
		return errors.New("validator reward ratio cannot be lower than zero")
	}

	return nil
}

// validateRewardSubmitterRatio validates the RewardSubmitterRatio param
func validateRewardSubmitterRatio(v interface{}) error {
	rewardSubmitterRatio, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if rewardSubmitterRatio.LT(sdk.ZeroDec()) {
		return errors.New("submitter reward ratio cannot be lower than zero")
	}

	return nil
}

// validateRewardSubmitterThreshold validates the RewardSubmitterThreshold param
func validateRewardSubmitterThreshold(v interface{}) error {
	rewardSubmitterThreshold, ok := v.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if rewardSubmitterThreshold.LT(sdk.ZeroInt()) {
		return errors.New("submitter reward threshold cannot be lower than zero")
	}

	return nil
}

// validateHeartbeatInterval validates the HeartbeatInterval param
func validateHeartbeatInterval(v interface{}) error {
	heartbeatInterval, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}
	if heartbeatInterval == 0 {
		return errors.New("heartbeat interval cannot be zero")
	}

	return nil
}

// validateAttestationInturnInterval validates the AttestationInturnInterval param
func validateAttestationInturnInterval(v interface{}) error {
	inturnInterval, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}
	if inturnInterval == 0 {
		return errors.New("attestation inturn interval cannot be zero")
	}

	return nil
}

// validateAttestationKeptCount validates the AttestationKeptCount param
func validateAttestationKeptCount(v interface{}) error {
	count, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}
	if count == 0 {
		return errors.New("attestation kept count interval cannot be zero")
	}

	return nil
}
