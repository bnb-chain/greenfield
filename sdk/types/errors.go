package types

import (
	"errors"
)

var (
	KeyManagerNotInitError  = errors.New("Key manager is not initialized yet ")
	ChainIdNotSetError      = errors.New("ChainID is not set yet ")
	SimulatedGasPriceError  = errors.New("Simulated gas price is 0 ")
	FeeAmountNotValidError  = errors.New("Fee Amount coin should only be BNB")
	GasInfoNotProvidedError = errors.New("Gas limit and(or) Fee Amount missing in txOpt")
)
