package keyring

import (
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	ethHd "github.com/evmos/ethermint/crypto/hd"
)

var (
	// SupportedAlgorithms defines the list of signing algorithms used on Greenfield:
	//  - eth_secp256k1 (Ethereum)
	//  - eth_bls (Ethereum)
	SupportedAlgorithms = keyring.SigningAlgoList{ethHd.EthSecp256k1, hd.EthBLS}
	// SupportedAlgorithmsLedger defines the list of signing algorithms used on Greenfield for the Ledger device:
	//  - eth_secp256k1 (Ethereum)
	//  - eth_bls (Ethereum)
	SupportedAlgorithmsLedger = keyring.SigningAlgoList{ethHd.EthSecp256k1, hd.EthBLS}
)

// ETHAlgoOption defines a function keys options for the ethereum Secp256k1 curve.
// It supports eth_secp256k1, eth_bls keys for accounts.
func ETHAlgoOption() keyring.Option {
	return func(options *keyring.Options) {
		options.SupportedAlgos = SupportedAlgorithms
		options.SupportedAlgosLedger = SupportedAlgorithmsLedger
	}
}
