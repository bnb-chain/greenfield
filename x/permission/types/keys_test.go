package types

import (
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/math"
)

func TestParsePolicyIdFromQueueKey(t *testing.T) {
	for i := 0; i < 100; i++ {
		expiration := time.Now().Add(time.Duration(rand.Int63()))
		policyId := math.NewUint(rand.Uint64())
		key := PolicyPrefixQueue(&expiration, policyId.Bytes())
		recoverId := ParsePolicyIdFromQueueKey(key)
		if !recoverId.Equal(policyId) {
			t.Errorf("ParseIdFromQueueKey failed to recover policy id: %s", policyId.String())
		}
	}
}
