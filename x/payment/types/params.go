package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyReserveTime            = []byte("ReserveTime")
	DefaultReserveTime uint64 = 7 * 24 * 60 * 60 // 7 days
)

var (
	KeyLiquidateTime            = []byte("LiquidateTime")
	DefaultLiquidateTime uint64 = 24 * 60 * 60 // 1 day
)

var (
	KeyPaymentAccountCountLimit            = []byte("PaymentAccountCountLimit")
	DefaultPaymentAccountCountLimit uint64 = 200
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	reserveTime uint64,
	liquidateTime uint64,
	paymentAccountCountLimit uint64,
) Params {
	return Params{
		ReserveTime:              reserveTime,
		LiquidateTime:            liquidateTime,
		PaymentAccountCountLimit: paymentAccountCountLimit,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultReserveTime,
		DefaultLiquidateTime,
		DefaultPaymentAccountCountLimit,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyReserveTime, &p.ReserveTime, validateReserveTime),
		paramtypes.NewParamSetPair(KeyLiquidateTime, &p.LiquidateTime, validateLiquidateTime),
		paramtypes.NewParamSetPair(KeyPaymentAccountCountLimit, &p.PaymentAccountCountLimit, validatePaymentAccountCountLimit),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateReserveTime(p.ReserveTime); err != nil {
		return err
	}

	if err := validateLiquidateTime(p.LiquidateTime); err != nil {
		return err
	}

	if err := validatePaymentAccountCountLimit(p.PaymentAccountCountLimit); err != nil {
		return err
	}

	return nil
}

// String implements the Stringer interface.
func (p Params) String() string {
	out, _ := yaml.Marshal(p)
	return string(out)
}

// validateReserveTime validates the ReserveTime param
func validateReserveTime(v interface{}) error {
	reserveTime, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = reserveTime

	return nil
}

// validateLiquidateTime validates the LiquidateTime param
func validateLiquidateTime(v interface{}) error {
	liquidateTime, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = liquidateTime

	return nil
}

// validatePaymentAccountCountLimit validates the PaymentAccountCountLimit param
func validatePaymentAccountCountLimit(v interface{}) error {
	paymentAccountCountLimit, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = paymentAccountCountLimit

	return nil
}
