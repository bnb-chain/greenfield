package types

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyEventCountPerBlock = []byte("EventCountPerBlock")
	// TODO: Determine the default value
	DefaultEventCountPerBlock uint64 = 3
)

var (
	KeyChallengeExpirePeriod = []byte("ChallengeExpirePeriod")
	// TODO: Determine the default value
	DefaultChallengeExpirePeriod uint64 = 100
)

var (
	KeySlashCoolingOffPeriod = []byte("SlashCoolingOffPeriod")
	// TODO: Determine the default value
	DefaultSlashCoolingOffPeriod uint64 = 100
)

var (
	KeySlashDenom     = []byte("SlashDenom")
	DefaultSlashDenom = "deposit"
)

var (
	KeySlashAmountPerKb = []byte("SlashAmountPerKb")
	// TODO: Determine the default value
	DefaultSlashAmountPerKb = sdk.ZeroDec()
)

var (
	KeySlashAmountMin = []byte("SlashAmountMin")
	// TODO: Determine the default value
	DefaultSlashAmountMin = math.NewInt(10)
)

var (
	KeySlashAmountMax = []byte("SlashAmountMax")
	// TODO: Determine the default value
	DefaultSlashAmountMax = math.NewInt(100)
)

var (
	KeyRewardValidatorRatio = []byte("RewardValidatorRatio")
	// TODO: Determine the default value
	DefaultRewardValidatorRatio = sdk.NewDecWithPrec(5, 1)
)

var (
	KeyRewardChallengerRatio = []byte("RewardChallengerRatio")
	// TODO: Determine the default value
	DefaultRewardChallengerRatio = sdk.NewDecWithPrec(3, 1)
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	eventCountPerBlock uint64,
	challengeExpirePeriod uint64,
	slashCoolingOffPeriod uint64,
	slashDenom string,
	slashAmountPerKb sdk.Dec,
	slashAmountMin math.Int,
	slashAmountMax math.Int,
	rewardValidatorRatio sdk.Dec,
	rewardChallengerRatio sdk.Dec,
) Params {
	return Params{
		EventCountPerBlock:    eventCountPerBlock,
		ChallengeExpirePeriod: challengeExpirePeriod,
		SlashCoolingOffPeriod: slashCoolingOffPeriod,
		SlashDenom:            slashDenom,
		SlashAmountPerKb:      slashAmountPerKb,
		SlashAmountMin:        slashAmountMin,
		SlashAmountMax:        slashAmountMax,
		RewardValidatorRatio:  rewardValidatorRatio,
		RewardChallengerRatio: rewardChallengerRatio,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultEventCountPerBlock,
		DefaultChallengeExpirePeriod,
		DefaultSlashCoolingOffPeriod,
		DefaultSlashDenom,
		DefaultSlashAmountPerKb,
		DefaultSlashAmountMin,
		DefaultSlashAmountMax,
		DefaultRewardValidatorRatio,
		DefaultRewardChallengerRatio,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyEventCountPerBlock, &p.EventCountPerBlock, validateEventCountPerBlock),
		paramtypes.NewParamSetPair(KeyChallengeExpirePeriod, &p.ChallengeExpirePeriod, validateChallengeExpirePeriod),
		paramtypes.NewParamSetPair(KeySlashCoolingOffPeriod, &p.SlashCoolingOffPeriod, validateSlashCoolingOffPeriod),
		paramtypes.NewParamSetPair(KeySlashDenom, &p.SlashDenom, validateSlashDemon),
		paramtypes.NewParamSetPair(KeySlashAmountPerKb, &p.SlashAmountPerKb, validateSlashAmountPerKb),
		paramtypes.NewParamSetPair(KeySlashAmountMin, &p.SlashAmountMin, validateSlashAmountMin),
		paramtypes.NewParamSetPair(KeySlashAmountMax, &p.SlashAmountMax, validateSlashAmountMax),
		paramtypes.NewParamSetPair(KeyRewardValidatorRatio, &p.RewardValidatorRatio, validateRewardValidatorRatio),
		paramtypes.NewParamSetPair(KeyRewardChallengerRatio, &p.RewardChallengerRatio, validateRewardChallengerRatio),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateEventCountPerBlock(p.EventCountPerBlock); err != nil {
		return err
	}

	if err := validateChallengeExpirePeriod(p.ChallengeExpirePeriod); err != nil {
		return err
	}

	if err := validateSlashCoolingOffPeriod(p.SlashCoolingOffPeriod); err != nil {
		return err
	}

	if err := validateSlashAmountPerKb(p.SlashAmountPerKb); err != nil {
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

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateEventCountPerBlock validates the EventCountPerBlock param
func validateEventCountPerBlock(v interface{}) error {
	eventCountPerBlock, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = eventCountPerBlock

	return nil
}

// validateChallengeExpirePeriod validates the ChallengeExpirePeriod param
func validateChallengeExpirePeriod(v interface{}) error {
	challengeExpirePeriod, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = challengeExpirePeriod

	return nil
}

// validateSlashCoolingOffPeriod validates the SlashCoolingOffPeriod param
func validateSlashCoolingOffPeriod(v interface{}) error {
	slashCoolingOffPeriod, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = slashCoolingOffPeriod

	return nil
}

// validateSlashDemon validates the SlashDemon param
func validateSlashDemon(v interface{}) error {
	slashDemon, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = slashDemon

	return nil
}

// validateSlashAmountPerKb validates the SlashAmountPerKb param
func validateSlashAmountPerKb(v interface{}) error {
	slashAmountPerKb, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = slashAmountPerKb

	return nil
}

// validateSlashAmountMin validates the SlashAmountMin param
func validateSlashAmountMin(v interface{}) error {
	slashAmountMin, ok := v.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = slashAmountMin

	return nil
}

// validateSlashAmountMax validates the SlashAmountMax param
func validateSlashAmountMax(v interface{}) error {
	slashAmountMax, ok := v.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = slashAmountMax

	return nil
}

// validateRewardValidatorRatio validates the RewardValidatorRatio param
func validateRewardValidatorRatio(v interface{}) error {
	rewardValidatorRatio, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = rewardValidatorRatio

	return nil
}

// validateRewardChallengerRatio validates the RewardChallengerRatio param
func validateRewardChallengerRatio(v interface{}) error {
	rewardChallengerRatio, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = rewardChallengerRatio

	return nil
}
