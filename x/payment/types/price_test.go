package types

import (
	"testing"

	"github.com/bnb-chain/greenfield/testutil/sample"

	sdkmath "cosmossdk.io/math"
)

func TestStreamRecordChange(t *testing.T) {
	addr := sample.RandAccAddress()
	src := NewDefaultStreamRecordChangeWithAddr(addr)
	t.Logf("src: %+v", src)
	src2 := NewDefaultStreamRecordChangeWithAddr(addr).WithRateChange(sdkmath.ZeroInt())
	t.Logf("src2: %+v", src2)
	src3 := NewDefaultStreamRecordChangeWithAddr(addr).WithRateChange(sdkmath.ZeroInt()).WithStaticBalanceChange(sdkmath.NewIntFromUint64(111))
	t.Logf("src3: %+v", src3)
}
