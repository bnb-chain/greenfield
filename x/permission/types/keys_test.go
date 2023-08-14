package types

import (
	"math/rand"
	"testing"
	"time"

	"cosmossdk.io/math"
)

func TestParsePolicyIdFromQueueKey(t *testing.T) {
	policyIds := []math.Uint{math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64()), math.NewUint(rand.Uint64())}

	expiration := time.Now()
	for _, policyId := range policyIds {
		key := PolicyPrefixQueue(&expiration, policyId.Bytes())
		recoverId := ParsePolicyIdFromQueueKey(key)
		if !recoverId.Equal(policyId) {
			t.Errorf("ParseIdFromQueueKey failed to recover policy id: %s", policyId.String())
		}
	}
}
