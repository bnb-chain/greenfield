package storage

import (
	"math/rand"
	"sync"
	"time"
	"unsafe"
)

const (
	letterBytes   = "abcdefghijklmnopqrstuvwxyz01234569"
	letterIdxBits = 6                    // 6 bits to represent a letter index
	letterIdxMask = 1<<letterIdxBits - 1 // All 1-bits, as many as letterIdxBits
	letterIdxMax  = 63 / letterIdxBits   // # of letter indices fitting in 63 bits
)

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
	return randString(rand.Intn(10) + 1)
}

// GenRandomBucketName generate random bucket name.
func GenRandomBucketName() string {
	return randString(rand.Intn(10) + 3)
}

// GenRandomGroupName generate random group name.
func GenRandomGroupName() string {
	return randString(rand.Intn(10) + 3)
}
