package types

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

var (
	DefaultTransferOutRelayerFeeParam    = sdkmath.NewInt(250000000000000) // 0.00025
	DefaultTransferOutAckRelayerFeeParam = sdkmath.NewInt(0)
)

var (
	KeyParamTransferOutRelayerFee    = []byte("TransferOutRelayerFee")
	KeyParamTransferOutAckRelayerFee = []byte("TransferOutAckRelayerFee")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		TransferOutRelayerFee:    DefaultTransferOutRelayerFeeParam,
		TransferOutAckRelayerFee: DefaultTransferOutAckRelayerFeeParam,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyParamTransferOutRelayerFee, &p.TransferOutRelayerFee, validateRelayerFee),
		paramtypes.NewParamSetPair(KeyParamTransferOutAckRelayerFee, &p.TransferOutAckRelayerFee, validateRelayerFee),
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
