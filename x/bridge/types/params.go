package types

import (
	"fmt"
	"math/big"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultTransferOutRelayerFeeParam    string = "1000000000000000" // 0.001
	DefaultTransferOutAckRelayerFeeParam string = "0"
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
		paramtypes.NewParamSetPair(KeyParamTransferOutRelayerFee, &p.TransferOutRelayerFee, validateTransferOutRelayerFee),
		// todo(quality): forget to use a new function after copy `validateTransferOutRelayerFee`?
		paramtypes.NewParamSetPair(KeyParamTransferOutAckRelayerFee, &p.TransferOutAckRelayerFee, validateTransferOutRelayerFee),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	// todo(quality): could reuse the function `validateTransferOutRelayerFee` and `validateTransferOutAckRelayerFee`
	relayerFee, valid := big.NewInt(0).SetString(p.TransferOutRelayerFee, 10)
	if !valid {
		return fmt.Errorf("invalid transfer out relayer fee, is %s", p.TransferOutRelayerFee)
	}
	if relayerFee.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("transfer out relayer fee should not be negative, is %s", p.TransferOutRelayerFee)
	}

	ackRelayerFee, valid := big.NewInt(0).SetString(p.TransferOutAckRelayerFee, 10)
	if !valid {
		return fmt.Errorf("invalid ack transfer out relayer fee, is %s", p.TransferOutAckRelayerFee)
	}
	if ackRelayerFee.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("transfer out ack relayer fee should not be negative, is %s", p.TransferOutAckRelayerFee)
	}

	return nil
}

func validateTransferOutRelayerFee(i interface{}) error {
	v, ok := i.(string)
	if !ok {
		return fmt.Errorf("invalid parameter type: %T", i)
	}

	relayerFee := big.NewInt(0)
	relayerFee, valid := relayerFee.SetString(v, 10)

	if !valid {
		return fmt.Errorf("invalid transfer out relayer fee, %s", v)
	}

	if relayerFee.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("invalid transfer out relayer fee, %s", v)
	}

	return nil
}
