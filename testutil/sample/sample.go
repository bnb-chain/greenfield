package sample

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/cometbft/cometbft/crypto/tmhash"
	"github.com/cosmos/cosmos-sdk/crypto/keys/eth/ethsecp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
)

func RandAccAddress() sdk.AccAddress {
	pk, err := ethsecp256k1.GenPrivKey()
	if err != nil {
		panic(err)
	}
	return sdk.AccAddress(pk.PubKey().Address())
}

func RandAccAddressHex() string {
	pk, err := ethsecp256k1.GenPrivKey()
	if err != nil {
		panic(err)
	}
	return sdk.AccAddress(pk.PubKey().Address()).String()
}

func RandSignBytes() (addr sdk.AccAddress, signBytes, sig []byte) {
	signBytes = RandStr(256)
	privKey, _ := ethsecp256k1.GenPrivKey()

	sig, _ = privKey.Sign(sdk.Keccak256(signBytes))
	pk := privKey.PubKey()
	addr = sdk.AccAddress(pk.Address())
	return addr, signBytes, sig
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
