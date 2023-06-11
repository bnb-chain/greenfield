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
	KeyMaxAutoForceSettleNum    = []byte("MaxAutoForceSettleNum")
	KeyFeeDenom                 = []byte("FeeDenom")
	KeyValidatorTaxRate         = []byte("ValidatorTaxRate")

	DefaultReserveTime      uint64  = 180 * 24 * 60 * 60       // 180 days
	DefaultValidatorTaxRate sdk.Dec = sdk.NewDecWithPrec(1, 2) // 1%

	DefaultForcedSettleTime         uint64 = 24 * 60 * 60 // 1 day
	DefaultPaymentAccountCountLimit uint64 = 200
	DefaultMaxAutoForceSettleNum    uint64 = 100
	DefaultFeeDenom                 string = "BNB"
)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// NewParams creates a new Params instance
func NewParams(
	reserveTime uint64,
	validatorTaxRate sdk.Dec,
	forcedSettleTime uint64,
	paymentAccountCountLimit uint64,
	maxAutoForceSettleNum uint64,
	feeDenom string,
) Params {
	return Params{
		VersionedParams:          VersionedParams{ReserveTime: reserveTime, ValidatorTaxRate: validatorTaxRate},
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
		DefaultValidatorTaxRate,
		DefaultForcedSettleTime,
		DefaultPaymentAccountCountLimit,
		DefaultMaxAutoForceSettleNum,
		DefaultFeeDenom,
	)
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyReserveTime, &p.VersionedParams.ReserveTime, validateReserveTime),
		paramtypes.NewParamSetPair(KeyValidatorTaxRate, &p.VersionedParams.ValidatorTaxRate, validateValidatorTaxRate),
		paramtypes.NewParamSetPair(KeyForcedSettleTime, &p.ForcedSettleTime, validateForcedSettleTime),
		paramtypes.NewParamSetPair(KeyPaymentAccountCountLimit, &p.PaymentAccountCountLimit, validatePaymentAccountCountLimit),
		paramtypes.NewParamSetPair(KeyMaxAutoForceSettleNum, &p.MaxAutoForceSettleNum, validateMaxAutoForceSettleNum),
		paramtypes.NewParamSetPair(KeyFeeDenom, &p.FeeDenom, validateFeeDenom),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	if err := validateReserveTime(p.VersionedParams.ReserveTime); err != nil {
		return err
	}

	if err := validateValidatorTaxRate(p.VersionedParams.ValidatorTaxRate); err != nil {
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

// validateMaxAutoForceSettleNum validates the MaxAutoForceSettleNum param
func validateMaxAutoForceSettleNum(v interface{}) error {
	maxAutoForceSettleNum, ok := v.(uint64)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", v)
	}

	if maxAutoForceSettleNum <= 0 {
		return fmt.Errorf("max auto force settle num must be positive")
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
