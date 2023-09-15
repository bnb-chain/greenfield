package types

import (
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
)

func TestDepositDenom(t *testing.T) {
	tests := []struct {
		name  string
		denom interface{}
		err   string
	}{

		{
			name:  "valid",
			denom: "denom",
		},
		{
			name:  "invalid type",
			denom: 1,
			err:   "invalid parameter type",
		},
		{
			name:  "empty",
			denom: " ",
			err:   "deposit denom cannot be blank",
		},
		{
			name:  "invalid denom",
			denom: "%",
			err:   "invalid denom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDepositDenom(tt.denom)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestGVGStakingPerBytes(t *testing.T) {
	var nilInt math.Int
	tests := []struct {
		name  string
		ratio interface{}
		err   string
	}{

		{
			name:  "valid",
			ratio: sdk.NewInt(1),
		},
		{
			name:  "invalid type",
			ratio: 1,
			err:   "invalid parameter type",
		},
		{
			name:  "invalid value",
			ratio: nilInt,
			err:   "invalid value for GVG staking per bytes",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGVGStakingPerBytes(tt.ratio)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMaxGlobalVirtualGroupNumPerFamily(t *testing.T) {
	tests := []struct {
		name   string
		number interface{}
		err    string
	}{

		{
			name:   "valid",
			number: uint32(1),
		},
		{
			name:   "invalid type",
			number: 1,
			err:    "invalid parameter type",
		},
		{
			name:   "invalid size",
			number: uint32(0),
			err:    "max GVG per family must be positive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMaxGlobalVirtualGroupNumPerFamily(tt.number)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMaxStoreSizePerFamily(t *testing.T) {
	tests := []struct {
		name string
		size interface{}
		err  string
	}{

		{
			name: "valid",
			size: uint64(1),
		},
		{
			name: "invalid type",
			size: 1,
			err:  "invalid parameter type",
		},
		{
			name: "invalid size",
			size: uint64(0),
			err:  "max store size per GVG family must be positive",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMaxStoreSizePerFamily(tt.size)
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestValidateParams(t *testing.T) {
	err := DefaultParams().Validate()
	require.NoError(t, err)
}
