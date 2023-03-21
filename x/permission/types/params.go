package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

const (
	DefaultMaxStatementsNum  uint64 = 10
	DefaultMaxPolicyGroupNum uint64 = 10
)

var (
	KeyMaxStatementsNum   = []byte("MaxStatementsNum")
	KeyMaxPolicyGroupSIze = []byte("MaxPolicyGroupSize")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maximumStatementsNum, maximumGroupNum uint64) Params {
	return Params{
		MaximumStatementsNum: maximumStatementsNum,
		MaximumGroupNum:      maximumGroupNum,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxStatementsNum, DefaultMaxPolicyGroupNum)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxStatementsNum, &p.MaximumStatementsNum, validateMaximumStatementsNum),
		paramtypes.NewParamSetPair(KeyMaxPolicyGroupSIze, &p.MaximumGroupNum, validateMaximumGroupNum),
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
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
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
