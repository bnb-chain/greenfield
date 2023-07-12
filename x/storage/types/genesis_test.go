package types_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/storage/types"
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
				Params: types.Params{
					VersionedParams: types.VersionedParams{
						MaxSegmentSize:          20,
						RedundantDataChunkNum:   10,
						RedundantParityChunkNum: 8,
						MinChargeSize:           100,
					},
					MaxPayloadSize:               2000,
					MaxBucketsPerAccount:         100,
					BscMirrorBucketRelayerFee:    "10",
					BscMirrorBucketAckRelayerFee: "10",
					BscMirrorGroupRelayerFee:     "10",
					BscMirrorGroupAckRelayerFee:  "10",
					BscMirrorObjectRelayerFee:    "10",
					BscMirrorObjectAckRelayerFee: "10",
					DiscontinueCountingWindow:    1000,
					DiscontinueObjectMax:         10000,
					DiscontinueBucketMax:         10000,
					DiscontinueConfirmPeriod:     100,
				},

				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
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
