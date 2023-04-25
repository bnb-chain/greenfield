package sample

import (
	"crypto/rand"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccAddress returns a sample account address
func AccAddress() string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr).String()
}

func RandAccAddress() sdk.AccAddress {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr)
}

func Checksum() []byte {
	return sdk.Keccak256(RandStr(256))
}

func RandStr(length int) []byte {
	randBytes := make([]byte, length/2)
	// #nosec
	_, _ = rand.Read(randBytes)
	return randBytes
}
