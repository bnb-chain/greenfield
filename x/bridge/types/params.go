package types

import (
	"fmt"
	"math/big"

	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
)

const (
	DefaultTransferOutRelayerSynFeeParam string = "1"
	DefaultTransferOutRelayerAckFeeParam string = "0"
)

var (
	KeyParamTransferOutRelayerSynFee = []byte("TransferOutRelayerSynFee")
	KeyParamTransferOutRelayerAckFee = []byte("TransferOutRelayerAckFee")
)

var _ paramtypes.ParamSet = (*Params)(nil)

// ParamKeyTable the param key table for launch module
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}

// DefaultParams returns a default set of parameters
func DefaultParams() Params {
	return Params{
		TransferOutRelayerSynFee: DefaultTransferOutRelayerSynFeeParam,
		TransferOutRelayerAckFee: DefaultTransferOutRelayerAckFeeParam,
	}
}

// ParamSetPairs get the params.ParamSet
func (p *Params) ParamSetPairs() paramtypes.ParamSetPairs {
	return paramtypes.ParamSetPairs{
		paramtypes.NewParamSetPair(KeyParamTransferOutRelayerSynFee, p.TransferOutRelayerSynFee, validateTransferOutRelayerFee),
		paramtypes.NewParamSetPair(KeyParamTransferOutRelayerAckFee, p.TransferOutRelayerAckFee, validateTransferOutRelayerFee),
	}
}

// Validate validates the set of params
func (p Params) Validate() error {
	relayerSynFee, valid := big.NewInt(0).SetString(p.TransferOutRelayerSynFee, 10)
	if !valid {
		return fmt.Errorf("invalid transfer out relayer syn fee, is %s", p.TransferOutRelayerSynFee)
	}
	if relayerSynFee.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("transfer out relayer syn fee should not be negative, is %s", p.TransferOutRelayerSynFee)
	}

	relayerAckFee, valid := big.NewInt(0).SetString(p.TransferOutRelayerAckFee, 10)
	if !valid {
		return fmt.Errorf("invalid transfer out relayer ack fee, is %s", p.TransferOutRelayerAckFee)
	}
	if relayerAckFee.Cmp(big.NewInt(0)) < 0 {
		return fmt.Errorf("transfer out relayer ack fee should not be negative, is %s", p.TransferOutRelayerAckFee)
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
