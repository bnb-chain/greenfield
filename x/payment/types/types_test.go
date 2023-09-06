package types

import (
	"testing"

	sdkmath "cosmossdk.io/math"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestVerifyModuleAddress(t *testing.T) {
	require.Equal(t, "0x25b2f7C1aA3cCCeF718e8e3A7Ec1A7C521eef2a9", GovernanceAddress.String())
	require.Equal(t, "0xdF5F0588f6B09f0B9E58D3426252db25Dc74E7a1", ValidatorTaxPoolAddress.String())
}

func TestStreamRecordChange(t *testing.T) {
	addr := sample.RandAccAddress()
	src := NewDefaultStreamRecordChangeWithAddr(addr)
	t.Logf("src: %+v", src)
	src2 := NewDefaultStreamRecordChangeWithAddr(addr).WithRateChange(sdkmath.ZeroInt())
	t.Logf("src2: %+v", src2)
	src3 := NewDefaultStreamRecordChangeWithAddr(addr).WithRateChange(sdkmath.ZeroInt()).WithStaticBalanceChange(sdkmath.NewIntFromUint64(111))
	t.Logf("src3: %+v", src3)
}
