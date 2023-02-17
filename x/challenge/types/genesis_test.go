package types_test

import (
	"testing"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				Params: types.DefaultParams(),
				OngoingChallenges: []types.Challenge{
					{
						Id: uint64(0),
					},
					{
						Id: uint64(1),
					},
				},
				RecentSlashes: []types.Slash{
					{
						Id: 0,
					},
					{
						Id: 1,
					},
				},
				RecentSlashCount: 2,
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated ongoingChallenge",
			genState: &types.GenesisState{
				OngoingChallenges: []types.Challenge{
					{
						Id: uint64(0),
					},
					{
						Id: uint64(0),
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated recentSlash",
			genState: &types.GenesisState{
				RecentSlashes: []types.Slash{
					{
						Id: 0,
					},
					{
						Id: 0,
					},
				},
			},
			valid: false,
		},
		{
			desc: "invalid recentSlash count",
			genState: &types.GenesisState{
				RecentSlashes: []types.Slash{
					{
						Id: 1,
					},
				},
				RecentSlashCount: 0,
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
