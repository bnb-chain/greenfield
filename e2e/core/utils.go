package core

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
	"unsafe"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/crypto"
	"sigs.k8s.io/yaml"

	"github.com/bnb-chain/greenfield/sdk/keys"
)

const letterBytes = "abcdefghijklmnopqrstuvwxyz01234569"
const (
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

func GenRandomAddr() sdk.AccAddress {
	// #nosec
	return sdk.AccAddress(crypto.AddressHash([]byte(fmt.Sprintf("%d", rand.Int()))))
}

func GenRandomHexString(len int) string {
	b := make([]byte, len)
	// #nosec
	_, err := rand.Read(b)
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

var mtx sync.Mutex

func randString(n int) string {
	mtx.Lock()
	src := rand.NewSource(time.Now().UnixNano())
	time.Sleep(1 * time.Millisecond)
	mtx.Unlock()

	b := make([]byte, n)
	// A src.Int63() generates 63 random bits, enough for letterIdxMax characters!
	for i, cache, remain := n-1, src.Int63(), letterIdxMax; i >= 0; {
		if remain == 0 {
			cache, remain = src.Int63(), letterIdxMax
		}
		if idx := int(cache & letterIdxMask); idx < len(letterBytes) {
			b[i] = letterBytes[idx]
			i--
		}
		cache >>= letterIdxBits
		remain--
	}

	return *(*string)(unsafe.Pointer(&b))
}

// GenRandomObjectName generate random object name.
func GenRandomObjectName() string {
	return randString(20)
}

// GenRandomBucketName generate random bucket name.
func GenRandomBucketName() string {
	return randString(10)
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
