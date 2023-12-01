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
	// DefaultGVGStakingPerBytes defines the default gvg staking price
	DefaultGVGStakingPerBytes                = sdk.NewInt(16000) // 20%~30% of store price
	DefaultMaxGlobalVirtualGroupNumPerFamily = uint32(10)
	DefaultMaxStoreSizePerFamily             = uint64(64) * 1024 * 1024 * 1024 * 1024 //64T
	DefaultSwapInValidityPeriod              = uint64(60) * 60 * 24 * 7               // 7 days
	DefaultSPConcurrentExitNum               = uint32(1)

	KeyDepositDenom                      = []byte("DepositDenom")
	KeyGVGStakingPerBytes                = []byte("GVGStakingPerBytes")
	KeyMaxGlobalVirtualGroupNumPerFamily = []byte("MaxGlobalVirtualGroupNumPerFamily")
	KeyMaxStoreSizePerFamily             = []byte("MaxStoreSizePerFamily")
	KeySwapInValidityPeriod              = []byte("SwapInValidityPeriod")
	KeySPConcurrentExitNum               = []byte("SPConcurrentExitNum")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(depositDenom string, gvgStakingPerBytes math.Int, maxGlobalVirtualGroupPerFamily uint32,
	maxStoreSizePerFamily, swapInValidityPeriod uint64, spConcurrentExitNum uint32) Params {
	return Params{
		DepositDenom:                      depositDenom,
		GvgStakingPerBytes:                gvgStakingPerBytes,
		MaxGlobalVirtualGroupNumPerFamily: maxGlobalVirtualGroupPerFamily,
		MaxStoreSizePerFamily:             maxStoreSizePerFamily,
		SwapInValidityPeriod:              swapInValidityPeriod,
		SpConcurrentExitNum:               spConcurrentExitNum,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(DefaultDepositDenom,
		DefaultGVGStakingPerBytes,
		DefaultMaxGlobalVirtualGroupNumPerFamily,
		DefaultMaxStoreSizePerFamily,
		DefaultSwapInValidityPeriod,
		DefaultSPConcurrentExitNum)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyDepositDenom, &p.DepositDenom, validateDepositDenom),
		paramtypes.NewParamSetPair(KeyGVGStakingPerBytes, &p.GvgStakingPerBytes, validateGVGStakingPerBytes),
		paramtypes.NewParamSetPair(KeyMaxGlobalVirtualGroupNumPerFamily, &p.MaxGlobalVirtualGroupNumPerFamily, validateMaxGlobalVirtualGroupNumPerFamily),
		paramtypes.NewParamSetPair(KeyMaxStoreSizePerFamily, &p.MaxStoreSizePerFamily, validateMaxStoreSizePerFamily),
		paramtypes.NewParamSetPair(KeySwapInValidityPeriod, &p.SwapInValidityPeriod, validateSwapInValidityPeriod),
		paramtypes.NewParamSetPair(KeySPConcurrentExitNum, &p.SpConcurrentExitNum, validateSPConcurrentExitNum),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateDepositDenom(p.DepositDenom); err != nil {
		return err
	}
	if err := validateGVGStakingPerBytes(p.GvgStakingPerBytes); err != nil {
		return err
	}
	if err := validateMaxGlobalVirtualGroupNumPerFamily(p.MaxGlobalVirtualGroupNumPerFamily); err != nil {
		return err
	}
	if err := validateMaxStoreSizePerFamily(p.MaxStoreSizePerFamily); err != nil {
		return err
	}
	if err := validateSwapInValidityPeriod(p.SwapInValidityPeriod); err != nil {
		return err
	}
	if err := validateSwapInValidityPeriod(p.SwapInValidityPeriod); err != nil {
		return err
	}

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

func validateGVGStakingPerBytes(i interface{}) error {
	v, ok := i.(math.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v.IsNil() || !v.IsPositive() {
		return fmt.Errorf("invalid value for GVG staking per bytes")
	}
	return nil
}

func validateMaxGlobalVirtualGroupNumPerFamily(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max GVG per family must be positive: %d", v)
	}

	return nil
}

func validateMaxStoreSizePerFamily(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("max store size per GVG family must be positive: %d", v)
	}

	return nil
}

func validateSwapInValidityPeriod(i interface{}) error {
	v, ok := i.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}
	if v == 0 {
		return fmt.Errorf("swapIn info validity period must be positive: %d", v)
	}
	return nil
}

func validateSPConcurrentExitNum(i interface{}) error {
	v, ok := i.(uint32)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if v == 0 {
		return fmt.Errorf("number of sp concurrent exit must be positive: %d", v)
	}

	return nil
}
