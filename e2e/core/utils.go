package core

import (
	"fmt"
	"math/rand"

	"github.com/bnb-chain/greenfield/sdk/keys"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
)

func GenRandomAddr() sdk.AccAddress {
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%d", rand.Int()))))
}

func GenRandomHexString(len int) string {
	b := make([]byte, len)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func GenRandomKeyManager() keys.KeyManager {
	//privKey := crypto.AddressHash([]byte(fmt.Sprintf("%d", rand.Int()))).String()
	//fmt.Printf("privKey: %s", privKey)
	keyManager, err := keys.NewPrivateKeyManager(GenRandomHexString(32))
	if err != nil {
		panic(err)
	}
	return keyManager
}
