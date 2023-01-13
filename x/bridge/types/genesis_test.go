package types_test

import (
	"testing"

	"github.com/bnb-chain/bfs/x/bridge/types"
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
				Params: types.Params{
					TransferOutRelayerFee:    "1",
					TransferOutAckRelayerFee: "0",
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "invalid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{
					TransferOutRelayerFee:    "1",
					TransferOutAckRelayerFee: "-1",
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: false,
		},
		{
			desc: "invalid genesis state",
			genState: &types.GenesisState{
				Params: types.Params{
					TransferOutRelayerFee:    "-1",
					TransferOutAckRelayerFee: "1",
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: false,
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
