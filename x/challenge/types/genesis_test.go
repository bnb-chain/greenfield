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

				OngoingChallengeList: []types.OngoingChallenge{
					{
						ChallengeId: "0",
					},
					{
						ChallengeId: "1",
					},
				},
				RecentSlashList: []types.RecentSlash{
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
				OngoingChallengeList: []types.OngoingChallenge{
					{
						ChallengeId: "0",
					},
					{
						ChallengeId: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated recentSlash",
			genState: &types.GenesisState{
				RecentSlashList: []types.RecentSlash{
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
				RecentSlashList: []types.RecentSlash{
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
