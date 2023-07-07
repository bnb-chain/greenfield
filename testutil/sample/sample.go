package sample

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
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

func RandBlsPubKey() []byte {
	blsPrivKey, _ := bls.RandKey()
	return blsPrivKey.PublicKey().Marshal()
}

func RandBlsPubKeyHex() string {
	blsPrivKey, _ := bls.RandKey()
	return hex.EncodeToString(blsPrivKey.PublicKey().Marshal())
}
