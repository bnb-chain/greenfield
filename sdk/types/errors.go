package types

import (
	"errors"
)

var (
	KeyManagerNotInitError = errors.New("Key manager is not initialized yet ")
	ChainIdNotSetError     = errors.New("ChainID is not set yet ")
	SimulatedGasPriceError = errors.New("Simulated gas price is 0 ")
)
