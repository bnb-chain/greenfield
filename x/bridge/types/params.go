package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

var (
	DefaultTransferOutRelayerFeeParam    = sdkmath.NewInt(250000000000000) // 0.00025
	DefaultTransferOutAckRelayerFeeParam = sdkmath.NewInt(0)
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		TransferOutRelayerFee:    DefaultTransferOutRelayerFeeParam,
		TransferOutAckRelayerFee: DefaultTransferOutAckRelayerFeeParam,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	err := validateRelayerFee(p.TransferOutRelayerFee)
	if err != nil {
		return err
	}

	err = validateRelayerFee(p.TransferOutAckRelayerFee)
	if err != nil {
		return err
	}
	return nil
}

func validateRelayerFee(i interface{}) error {
	fee, ok := i.(sdkmath.Int)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	if fee.IsNil() {
		return fmt.Errorf("relay fee should not be nil")
	}

	if fee.IsNegative() {
		return fmt.Errorf("relay fee should not less than 0")
	}

	return nil
}
