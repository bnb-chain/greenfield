package types

import (
	"fmt"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"gopkg.in/yaml.v2"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyReserveTime              = []byte("ReserveTime")
	KeyForcedSettleTime         = []byte("ForcedSettleTime")
	KeyPaymentAccountCountLimit = []byte("PaymentAccountCountLimit")
	KeyMaxAutoForceSettleNum    = []byte("MaxAutoForceSettleNum")
	KeyFeeDenom                 = []byte("FeeDenom")

	DefaultReserveTime              uint64 = 180 * 24 * 60 * 60 // 180 days
	DefaultForcedSettleTime         uint64 = 24 * 60 * 60       // 1 day
	DefaultPaymentAccountCountLimit uint64 = 200
	DefaultMaxAutoForceSettleNum    uint64 = 100
	DefaultFeeDenom                 string = "gweibnb"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	reserveTime uint64,
	forcedSettleTime uint64,
	paymentAccountCountLimit uint64,
	maxAutoForceSettleNum uint64,
	feeDenom string,
) Params {
	return Params{
		ReserveTime:              reserveTime,
		ForcedSettleTime:         forcedSettleTime,
		PaymentAccountCountLimit: paymentAccountCountLimit,
		MaxAutoForceSettleNum:    maxAutoForceSettleNum,
		FeeDenom:                 feeDenom,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultReserveTime,
		DefaultForcedSettleTime,
		DefaultPaymentAccountCountLimit,
		DefaultMaxAutoForceSettleNum,
		DefaultFeeDenom,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyReserveTime, &p.ReserveTime, validateReserveTime),
		paramtypes.NewParamSetPair(KeyForcedSettleTime, &p.ForcedSettleTime, validateForcedSettleTime),
		paramtypes.NewParamSetPair(KeyPaymentAccountCountLimit, &p.PaymentAccountCountLimit, validatePaymentAccountCountLimit),
		paramtypes.NewParamSetPair(KeyMaxAutoForceSettleNum, &p.MaxAutoForceSettleNum, validateMaxAutoForceSettleNum),
		paramtypes.NewParamSetPair(KeyFeeDenom, &p.FeeDenom, validateFeeDenom),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateReserveTime(p.ReserveTime); err != nil {
		return err
	}

	if err := validateForcedSettleTime(p.ForcedSettleTime); err != nil {
		return err
	}

	if err := validatePaymentAccountCountLimit(p.PaymentAccountCountLimit); err != nil {
		return err
	}

	if err := validatePaymentAccountCountLimit(p.MaxAutoForceSettleNum); err != nil {
		return err
	}

	if err := validateFeeDenom(p.FeeDenom); err != nil {
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

// validateForcedSettleTime validates the ForcedSettleTime param
func validateForcedSettleTime(v interface{}) error {
	ForcedSettleTime, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = ForcedSettleTime

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

// validateMaxAutoForceSettleNum validates the MaxAutoForceSettleNum param
func validateMaxAutoForceSettleNum(v interface{}) error {
	maxAutoForceSettleNum, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = maxAutoForceSettleNum

	return nil
}

// validateFeeDenom validates the FeeDenom param
func validateFeeDenom(v interface{}) error {
	feeDenom, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	// TODO implement validation
	_ = feeDenom

	return nil
}
