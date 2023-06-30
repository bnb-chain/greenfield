package app

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseDenomFromBalanceKey(t *testing.T) {
	key, _ := hex.DecodeString("0214cf150037e47b0c53e826a2d0050de1da2c8f5caa424e42")
	require.Equal(t, "BNB", parseDenomFromBalanceKey(key))
}

func TestParseDenomFromSupplyKey(t *testing.T) {
	key, _ := hex.DecodeString("00424e42")
	require.Equal(t, "BNB", parseDenomFromSupplyKey(key))
}
