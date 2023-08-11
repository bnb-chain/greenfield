package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
)

var (
	DefaultBscTransferOutRelayerFeeParam    = sdkmath.NewInt(780000000000000) // 0.00078
	DefaultBscTransferOutAckRelayerFeeParam = sdkmath.NewInt(0)
)

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		BscTransferOutRelayerFee:    DefaultBscTransferOutRelayerFeeParam,
		BscTransferOutAckRelayerFee: DefaultBscTransferOutAckRelayerFeeParam,
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	err := validateRelayerFee(p.BscTransferOutRelayerFee)
	if err != nil {
		return err
	}

	err = validateRelayerFee(p.BscTransferOutAckRelayerFee)
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
