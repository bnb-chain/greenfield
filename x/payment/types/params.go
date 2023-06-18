package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var _ paramtypes.ParamSet = (*Params)(nil)

var (
	KeyReserveTime              = []byte("ReserveTime")
	KeyForcedSettleTime         = []byte("ForcedSettleTime")
	KeyPaymentAccountCountLimit = []byte("PaymentAccountCountLimit")
	KeyMaxAutoSettleFlowCount   = []byte("MaxAutoSettleFlowCount")
	KeyMaxAutoResumeFlowCount   = []byte("MaxAutoResumeFlowCount")
	KeyFeeDenom                 = []byte("FeeDenom")
	KeyValidatorTaxRate         = []byte("ValidatorTaxRate")

	DefaultReserveTime              uint64  = 180 * 24 * 60 * 60 // 180 days
	DefaultForcedSettleTime         uint64  = 24 * 60 * 60       // 1 day
	DefaultPaymentAccountCountLimit uint64  = 200
	DefaultMaxAutoSettleFlowCount   uint64  = 100
	DefaultMaxAutoResumeFlowCount   uint64  = 100
	DefaultFeeDenom                 string  = "BNB"
	DefaultValidatorTaxRate         sdk.Dec = sdk.NewDecWithPrec(1, 2) // 1%
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
	MaxAutoSettleFlowCount uint64,
	maxAutoResumeFlowCount uint64,
	feeDenom string,
	validatorTaxRate sdk.Dec,
) Params {
	return Params{
		ReserveTime:              reserveTime,
		ForcedSettleTime:         forcedSettleTime,
		PaymentAccountCountLimit: paymentAccountCountLimit,
		MaxAutoSettleFlowCount:   MaxAutoSettleFlowCount,
		MaxAutoResumeFlowCount:   maxAutoResumeFlowCount,
		FeeDenom:                 feeDenom,
		ValidatorTaxRate:         validatorTaxRate,
	}
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return NewParams(
		DefaultReserveTime,
		DefaultForcedSettleTime,
		DefaultPaymentAccountCountLimit,
		DefaultMaxAutoSettleFlowCount,
		DefaultMaxAutoResumeFlowCount,
		DefaultFeeDenom,
		DefaultValidatorTaxRate,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyReserveTime, &p.ReserveTime, validateReserveTime),
		paramtypes.NewParamSetPair(KeyForcedSettleTime, &p.ForcedSettleTime, validateForcedSettleTime),
		paramtypes.NewParamSetPair(KeyPaymentAccountCountLimit, &p.PaymentAccountCountLimit, validatePaymentAccountCountLimit),
		paramtypes.NewParamSetPair(KeyMaxAutoSettleFlowCount, &p.MaxAutoSettleFlowCount, validateMaxAutoSettleFlowCount),
		paramtypes.NewParamSetPair(KeyMaxAutoResumeFlowCount, &p.MaxAutoResumeFlowCount, validateMaxAutoResumeFlowCount),
		paramtypes.NewParamSetPair(KeyFeeDenom, &p.FeeDenom, validateFeeDenom),
		paramtypes.NewParamSetPair(KeyValidatorTaxRate, &p.ValidatorTaxRate, validateValidatorTaxRate),
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

	if err := validateMaxAutoSettleFlowCount(p.MaxAutoSettleFlowCount); err != nil {
		return err
	}

	if err := validateMaxAutoResumeFlowCount(p.MaxAutoResumeFlowCount); err != nil {
		return err
	}

	if err := validateFeeDenom(p.FeeDenom); err != nil {
		return err
	}

	if err := validateValidatorTaxRate(p.ValidatorTaxRate); err != nil {
		return err
	}
	return nil
}

// validateReserveTime validates the ReserveTime param
func validateReserveTime(v interface{}) error {
	reserveTime, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if reserveTime <= 0 {
		return fmt.Errorf("reserve time must be positive")
	}

	return nil
}

// validateForcedSettleTime validates the ForcedSettleTime param
func validateForcedSettleTime(v interface{}) error {
	ForcedSettleTime, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if ForcedSettleTime <= 0 {
		return fmt.Errorf("forced settle time must be positive")
	}
	return nil
}

// validatePaymentAccountCountLimit validates the PaymentAccountCountLimit param
func validatePaymentAccountCountLimit(v interface{}) error {
	paymentAccountCountLimit, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if paymentAccountCountLimit <= 0 {
		return fmt.Errorf("payment account count limit must be positive")
	}

	return nil
}

// validateMaxAutoSettleFlowCount validates the MaxAutoSettleFlowCount param
func validateMaxAutoSettleFlowCount(v interface{}) error {
	maxAutoSettleFlowCount, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if maxAutoSettleFlowCount <= 0 {
		return fmt.Errorf("max force settle flow count must be positive")
	}

	return nil
}

// validateMaxAutoResumeFlowCount validates the MaxAutoResumeFlowCount param
func validateMaxAutoResumeFlowCount(v interface{}) error {
	maxAutoResumeFlowCount, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if maxAutoResumeFlowCount <= 0 {
		return fmt.Errorf("max auto resume flow count must be positive")
	}

	return nil
}

// validateFeeDenom validates the FeeDenom param
func validateFeeDenom(v interface{}) error {
	feeDenom, ok := v.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	_ = feeDenom

	return nil
}

// validateValidatorTaxRate validates the ValidatorTaxRate param
func validateValidatorTaxRate(v interface{}) error {
	validatorTaxRate, ok := v.(sdk.Dec)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if validatorTaxRate.IsNil() || validatorTaxRate.IsNegative() || validatorTaxRate.GT(sdk.OneDec()) {
		return fmt.Errorf("validator tax ratio should be between 0 and 1, is %s", validatorTaxRate)
	}

	return nil
}
