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
					MaxPayloadSize:                   2000,
					MaxBucketsPerAccount:             100,
					BscMirrorBucketRelayerFee:        "1",
					BscMirrorBucketAckRelayerFee:     "2",
					BscMirrorGroupRelayerFee:         "3",
					BscMirrorGroupAckRelayerFee:      "4",
					BscMirrorObjectRelayerFee:        "5",
					BscMirrorObjectAckRelayerFee:     "6",
					OpMirrorBucketRelayerFee:         "7",
					OpMirrorBucketAckRelayerFee:      "8",
					OpMirrorGroupRelayerFee:          "9",
					OpMirrorGroupAckRelayerFee:       "10",
					OpMirrorObjectRelayerFee:         "11",
					OpMirrorObjectAckRelayerFee:      "12",
					DiscontinueCountingWindow:        1000,
					DiscontinueObjectMax:             10000,
					DiscontinueBucketMax:             10000,
					DiscontinueConfirmPeriod:         100,
					DiscontinueDeletionMax:           10,
					StalePolicyCleanupMax:            10,
					MinQuotaUpdateInterval:           10000,
					MaxLocalVirtualGroupNumPerBucket: 100,
				},
			},
			valid: true,
		},
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
