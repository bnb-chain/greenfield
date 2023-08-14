package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultMaxStatementsNum                      uint64 = 10
	DefaultMaxPolicyGroupNum                     uint64 = 10
	DefaultMaximumRemoveExpiredPoliciesIteration uint64 = 100
)

var (
	KeyMaxStatementsNum                      = []byte("MaxStatementsNum")
	KeyMaxPolicyGroupSize                    = []byte("MaxPolicyGroupSize")
	KeyMaximumRemoveExpiredPoliciesIteration = []byte("MaximumRemoveExpiredPoliciesIteration")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maximumStatementsNum, maximumGroupNum, maximumRemoveExpiredPoliciesIteration uint64) Params {
	return Params{
		MaximumStatementsNum:                  maximumStatementsNum,
		MaximumGroupNum:                       maximumGroupNum,
		MaximumRemoveExpiredPoliciesIteration: maximumRemoveExpiredPoliciesIteration,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxStatementsNum, DefaultMaxPolicyGroupNum, DefaultMaximumRemoveExpiredPoliciesIteration)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxStatementsNum, &p.MaximumStatementsNum, validateMaximumStatementsNum),
		paramtypes.NewParamSetPair(KeyMaxPolicyGroupSize, &p.MaximumGroupNum, validateMaximumGroupNum),
		paramtypes.NewParamSetPair(KeyMaximumRemoveExpiredPoliciesIteration, &p.MaximumRemoveExpiredPoliciesIteration, validateMaximumRemoveExpiredPoliciesIteration),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMaximumGroupNum(p.MaximumStatementsNum); err != nil {
		return err
	}
	if err := validateMaximumGroupNum(p.MaximumGroupNum); err != nil {
		return err
	}
	if err := validateMaximumRemoveExpiredPoliciesIteration(p.MaximumRemoveExpiredPoliciesIteration); err != nil {
		return err
	}
	return nil
}

func validateMaximumStatementsNum(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max segment size must be positive: %d", v)
	}

	return nil
}

func validateMaximumGroupNum(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max payload size must be positive: %d", v)
	}

	return nil
}

func validateMaximumRemoveExpiredPoliciesIteration(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max RemoveExpiredPolicies iteration must be positive: %d", v)
	}

	return nil
}
