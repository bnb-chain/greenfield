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
	DefaultChallengeCountPerBlock uint64 = 3
)

var (
	KeySlashCoolingOffPeriod            = []byte("SlashCoolingOffPeriod")
	DefaultSlashCoolingOffPeriod uint64 = 100
)

var (
	KeySlashAmountSizeRate     = []byte("SlashAmountSizeRate")
	DefaultSlashAmountSizeRate = sdk.NewDecWithPrec(5, 1)
)

var (
	KeySlashAmountMin     = []byte("SlashAmountMin")
	DefaultSlashAmountMin = math.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(10), big.NewInt(1e18)))
)

var (
	KeySlashAmountMax     = []byte("SlashAmountMax")
	DefaultSlashAmountMax = math.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(100), big.NewInt(1e18)))
)

var (
	KeyRewardValidatorRatio     = []byte("RewardValidatorRatio")
	DefaultRewardValidatorRatio = sdk.NewDecWithPrec(5, 1)
)

var (
	KeyRewardChallengerRatio     = []byte("RewardChallengerRatio")
	DefaultRewardChallengerRatio = sdk.NewDecWithPrec(3, 1)
)

var (
	KeyHeartbeatInterval            = []byte("HeartbeatInterval")
	DefaultHeartbeatInterval uint64 = 100
)

var (
	KeyHeartbeatRewardRate     = []byte("HeartbeatRewardRate")
	DefaultHeartbeatRewardRate = sdk.NewDecWithPrec(1, 3)
)

var (
	KeyHeartbeatRewardThreshold     = []byte("HeartbeatRewardThreshold")
	DefaultHeartbeatRewardThreshold = math.NewIntFromBigInt(big.NewInt(1e15))
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	challengeCountPerBlock uint64,
	slashCoolingOffPeriod uint64,
	slashAmountSizeRate sdk.Dec,
	slashAmountMin math.Int,
	slashAmountMax math.Int,
	rewardValidatorRatio sdk.Dec,
	rewardChallengerRatio sdk.Dec,
	heartbeatInterval uint64,
	heartbeatRewardRate sdk.Dec,
	heartbeatRewardThreshold math.Int,
) Params {
	return Params{
		ChallengeCountPerBlock:   challengeCountPerBlock,
		SlashCoolingOffPeriod:    slashCoolingOffPeriod,
		SlashAmountSizeRate:      slashAmountSizeRate,
		SlashAmountMin:           slashAmountMin,
		SlashAmountMax:           slashAmountMax,
		RewardValidatorRatio:     rewardValidatorRatio,
		RewardChallengerRatio:    rewardChallengerRatio,
		HeartbeatInterval:        heartbeatInterval,
		HeartbeatRewardRate:      heartbeatRewardRate,
		HeartbeatRewardThreshold: heartbeatRewardThreshold,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultChallengeCountPerBlock,
		DefaultSlashCoolingOffPeriod,
		DefaultSlashAmountSizeRate,
		DefaultSlashAmountMin,
		DefaultSlashAmountMax,
		DefaultRewardValidatorRatio,
		DefaultRewardChallengerRatio,
		DefaultHeartbeatInterval,
		DefaultHeartbeatRewardRate,
		DefaultHeartbeatRewardThreshold,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyChallengeCountPerBlock, &p.ChallengeCountPerBlock, validateChallengeCountPerBlock),
		paramtypes.NewParamSetPair(KeySlashCoolingOffPeriod, &p.SlashCoolingOffPeriod, validateSlashCoolingOffPeriod),
		paramtypes.NewParamSetPair(KeySlashAmountSizeRate, &p.SlashAmountSizeRate, validateSlashAmountSizeRate),
		paramtypes.NewParamSetPair(KeySlashAmountMin, &p.SlashAmountMin, validateSlashAmountMin),
		paramtypes.NewParamSetPair(KeySlashAmountMax, &p.SlashAmountMax, validateSlashAmountMax),
		paramtypes.NewParamSetPair(KeyRewardValidatorRatio, &p.RewardValidatorRatio, validateRewardValidatorRatio),
		paramtypes.NewParamSetPair(KeyRewardChallengerRatio, &p.RewardChallengerRatio, validateRewardChallengerRatio),
		paramtypes.NewParamSetPair(KeyHeartbeatInterval, &p.HeartbeatInterval, validateHeartbeatInterval),
		paramtypes.NewParamSetPair(KeyHeartbeatRewardRate, &p.HeartbeatRewardRate, validateHeartbeatRewardRate),
		paramtypes.NewParamSetPair(KeyHeartbeatRewardThreshold, &p.HeartbeatRewardThreshold, validateHeartbeatRewardThreshold),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateChallengeCountPerBlock(p.ChallengeCountPerBlock); err != nil {
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

	if err := validateRewardChallengerRatio(p.RewardChallengerRatio); err != nil {
		return err
	}

	if p.SlashAmountMax.LTE(p.SlashAmountMin) {
		return errors.New("max slash amount should be bigger than min slash amount")
	}

	if p.RewardValidatorRatio.Add(p.RewardChallengerRatio).GT(sdk.NewDec(1)) {
		return errors.New("the sum of validator and challenger reward ratio should be equal to or less than one")
	}

	if err := validateHeartbeatInterval(p.HeartbeatInterval); err != nil {
		return err
	}

	if err := validateHeartbeatRewardRate(p.HeartbeatRewardRate); err != nil {
		return err
	}

	if err := validateHeartbeatRewardThreshold(p.HeartbeatRewardThreshold); err != nil {
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

// validateRewardChallengerRatio validates the RewardChallengerRatio param
func validateRewardChallengerRatio(v interface{}) error {
	rewardChallengerRatio, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if rewardChallengerRatio.LT(sdk.ZeroDec()) {
		return errors.New("challenger reward ratio cannot be lower than zero")
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

// validateHeartbeatRewardRate validates the HeartbeatRewardRate param
func validateHeartbeatRewardRate(v interface{}) error {
	heartbeatRewardRate, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if heartbeatRewardRate.LT(sdk.ZeroDec()) {
		return errors.New("heartbeat reward rate cannot be lower than zero")
	}

	return nil
}

// validateHeartbeatRewardThreshold validates the HeartbeatRewardThreshold param
func validateHeartbeatRewardThreshold(v interface{}) error {
	heartbeatRewardThreshold, ok := v.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if heartbeatRewardThreshold.LT(sdk.ZeroInt()) {
		return errors.New("heartbeat reward threshold cannot be lower than zero")
	}

	return nil
}
