package sample

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/crypto/keys/ed25519"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
)

// AccAddress returns a sample account address
func AccAddress() string {
	pk := ed25519.GenPrivKey().PubKey()
	addr := pk.Address()
	return sdk.AccAddress(addr).String()
}

func RandSignBytes() (addr sdk.AccAddress, signBytes []byte, sig []byte) {
	signBytes = RandStr(256)
	privKey, _ := ethsecp256k1.GenPrivKey()

	sig, _ = privKey.Sign(sdk.Keccak256(signBytes))
	pk := privKey.PubKey()
	addr = sdk.AccAddress(pk.Address())
	return addr, signBytes, sig
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

func RandBlsPubKeyAndBlsProofBz() ([]byte, []byte) {
	blsPriv, _ := bls.RandKey()
	blsPubKeyBz := blsPriv.PublicKey().Marshal()
	blsProofBz := blsPriv.Sign(tmhash.Sum(blsPubKeyBz)).Marshal()
	return blsPubKeyBz, blsProofBz
}

func RandBlsPubKeyAndBlsProof() (string, string) {
	blsPubKey, proof := RandBlsPubKeyAndBlsProofBz()
	return hex.EncodeToString(blsPubKey), hex.EncodeToString(proof)
}
