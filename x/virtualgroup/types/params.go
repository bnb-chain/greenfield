package types

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"

	"gopkg.in/yaml.v2"
)

const (
	// DefaultDepositDenom Dafault deposit denom
	DefaultDepositDenom = "BNB"
)

var (
	// DefaultMinDeposit defines the minimum deposit amount for all storage provider
	DefaultMinDeposit = math.NewIntFromBigInt(new(big.Int).Mul(big.NewInt(10000), big.NewInt(1e18)))

	// DefaultGVGStakingPrice defines the default gvg staking price
	// TODO: Set a reasonable value.
	DefaultGVGStakingPrice = sdk.NewDecFromIntWithPrec(sdk.NewInt(2), 18)

	KeyDepositDenom    = []byte("DepositDenom")
	KeyMinDeposit      = []byte("MinDeposit")
	KeyGVGStakingPrice = []byte("GVGStakingPrice")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(depositDenom string, minDeposit math.Int, baseGVGStorageStakingPrice sdk.Dec) Params {
	return Params{
		DepositDenom:    depositDenom,
		GvgStakingPrice: baseGVGStorageStakingPrice,
		MinDeposit:      minDeposit,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDepositDenom, DefaultMinDeposit, DefaultGVGStakingPrice)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyMinDeposit, &p.MinDeposit, validateMinDeposit),
		paramtypes.NewParamSetPair(KeyDepositDenom, &p.DepositDenom, validateDepositDenom),
		paramtypes.NewParamSetPair(KeyGVGStakingPrice, &p.GvgStakingPrice, validateGVGStakingPrice),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
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

	if v.IsNil() {
		return fmt.Errorf("minimum deposit amount cannot be nil")
	}

	if v.IsNegative() {
		return fmt.Errorf("minimum deposit amount cannot be lower than 0")
	}

	return nil
}

func validateGVGStakingPrice(i interface{}) error {
	v, ok := i.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() || !v.IsPositive() || v.GT(sdk.OneDec()) {
		return fmt.Errorf("invalid secondary sp store price ratio")
	}
	return nil
}
