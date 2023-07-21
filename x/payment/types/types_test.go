package types

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerifyModuleAddress(t *testing.T) {
	require.Equal(t, "0x25b2f7C1aA3cCCeF718e8e3A7Ec1A7C521eef2a9", GovernanceAddress.String())
	require.Equal(t, "0xdF5F0588f6B09f0B9E58D3426252db25Dc74E7a1", ValidatorTaxPoolAddress.String())
}
