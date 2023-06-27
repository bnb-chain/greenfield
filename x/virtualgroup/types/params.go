package types

import (
	"errors"
	"fmt"
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
	// DefaultGVGStakingPrice defines the default gvg staking price
	DefaultGVGStakingPrice                   = sdk.NewInt(16000)
	DefaultMaxGlobalVirtualGroupNumPerFamily = uint32(10)
	DefaultMaxStoreSizePerFamily             = uint64(64) * 1024 * 1024 * 1024 //64T

	KeyDepositDenom                      = []byte("DepositDenom")
	KeyGVGStakingPrice                   = []byte("GVGStakingPrice")
	KeyMaxGlobalVirtualGroupNumPerFamily = []byte("MaxGlobalVirtualGroupNumPerFamily")
	KeyMaxStoreSizePerFamily             = []byte("MaxStoreSizePerFamily")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(depositDenom string, gvgStakingPrice math.Int, maxGlobalVirtualGroupPerFamily uint32,
	maxStoreSizePerFamily uint64) Params {
	return Params{
		DepositDenom:                      depositDenom,
		GvgStakingPrice:                   gvgStakingPrice,
		MaxGlobalVirtualGroupNumPerFamily: maxGlobalVirtualGroupPerFamily,
		MaxStoreSizePerFamily:             maxStoreSizePerFamily,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDepositDenom, DefaultGVGStakingPrice, DefaultMaxGlobalVirtualGroupNumPerFamily, DefaultMaxStoreSizePerFamily)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDepositDenom, &p.DepositDenom, validateDepositDenom),
		paramtypes.NewParamSetPair(KeyGVGStakingPrice, &p.GvgStakingPrice, validateGVGStakingPrice),
		paramtypes.NewParamSetPair(KeyMaxGlobalVirtualGroupNumPerFamily, &p.MaxGlobalVirtualGroupNumPerFamily, validateMaxGlobalVirtualGroupNumPerFamily),
		paramtypes.NewParamSetPair(KeyMaxStoreSizePerFamily, &p.MaxStoreSizePerFamily, validateMaxStoreSizePerFamily),
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

func validateMaxGlobalVirtualGroupNumPerFamily(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max buckets per account must be positive: %d", v)
	}

	return nil
}

func validateMaxStoreSizePerFamily(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max buckets per account must be positive: %d", v)
	}

	return nil
}
