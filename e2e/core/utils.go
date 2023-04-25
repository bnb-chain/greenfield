package core

import (
	cryptoRand "crypto/rand"
	"fmt"
	"math/rand"

	"github.com/cometbft/cometbft/crypto"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"sigs.k8s.io/yaml"

	"github.com/bnb-chain/greenfield/sdk/keys"
)

func GenRandomAddr() sdk.AccAddress {
	// #nosec
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%d", rand.Int()))))
}

func GenRandomHexString(len int) string {
	b := make([]byte, len)
	// #nosec
	_, err := cryptoRand.Read(b)
	if err != nil {
		panic(err)
	}
	return fmt.Sprintf("%x", b)
}

func GenRandomKeyManager() keys.KeyManager {
	keyManager, err := keys.NewPrivateKeyManager(GenRandomHexString(32))
	if err != nil {
		panic(err)
	}
	return keyManager
}

func YamlString(data interface{}) string {
	bz, err := yaml.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(bz)
}

// RandInt64 generate random int64 between min and max
func RandInt64(min, max int64) int64 {
	return min + rand.Int63n(max-min)
}
