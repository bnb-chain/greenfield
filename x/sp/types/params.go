package types

import (
	"errors"
	fmt "fmt"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

// SP params default values
const (
	// Default maximum number of SP
	DefaultMaxStorageProviders uint32 = 100
	// Dafault
	DefaultDepositDenom = "deposit"
)

// DefaultMinDeposit defines the minimum deposit amount for all storage provider
var DefaulMinDeposit math.Int = math.NewInt(10000)

var (
	KeyMaxStorageProviders = []byte("MaxStorageProviders")
	KeyDepostDenom          = []byte("DepositDenom")
	KeyMinDeposit      = []byte("MinDeposit")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(maxStorageProviders uint32, depositDenom string, minDeposit math.Int) Params {
	return Params{
		MaxStorageProviders: maxStorageProviders,
		DepositDenom:          depositDenom,
		MinDeposit:      minDeposit,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultMaxStorageProviders, DefaultDepositDenom, DefaulMinDeposit)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMaxStorageProviders, &p.MaxStorageProviders, validateMaxStorageProviders),
		paramtypes.NewParamSetPair(KeyDepostDenom, &p.DepositDenom, validateDepositDenom),
		paramtypes.NewParamSetPair(KeyMinDeposit, &p.MinDeposit, validateMinDeposit),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateMaxStorageProviders(p.MaxStorageProviders); err != nil {
		return err
	}

	if err := validateDepositDenom(p.DepositDenom); err != nil {
		return err
	}

	if err := validateMinDeposit(p.MinDeposit); err != nil {
		return err
	}
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

func validateMaxStorageProviders(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max storage providers must be positive: %d", v)
	}

	return nil
}

func validateDepositDenom(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if strings.TrimSpace(v) == "" {
		return errors.New("deposit denom cannot be blank")
	}

	if err := sdk.ValidateDenom(v); err != nil {
		return err
	}

	return nil
}

func validateMinDeposit(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v.LT(sdk.NewInt(0)) {
		return fmt.Errorf("minimum deposit amount cannot be lower than 0")
	}

	return nil
}
