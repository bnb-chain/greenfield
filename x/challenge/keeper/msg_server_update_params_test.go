package keeper_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func (s *TestSuite) TestUpdateParams() {
	params := types.DefaultParams()
	params.HeartbeatInterval = 10

	tests := []struct {
		name string
		msg  types.MsgUpdateParams
		err  bool
	}{
		{
			name: "invalid authority",
			msg: types.MsgUpdateParams{
				Authority: sample.AccAddress(),
			},
			err: true,
		}, {
			name: "success",
			msg: types.MsgUpdateParams{
				Authority: s.challengeKeeper.GetAuthority(),
				Params:    params,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			_, err := s.msgServer.UpdateParams(s.ctx, &tt.msg)
			if tt.err {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}

	// verify storage
	s.Require().Equal(params, s.challengeKeeper.GetParams(s.ctx))
}
