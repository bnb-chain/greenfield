package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgHeartbeat_ValidateBasic(t *testing.T) {
	var sig [96]byte
	tests := []struct {
		name string
		msg  MsgHeartbeat
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgHeartbeat{
				Submitter: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid vote aggregated signature",
			msg: MsgHeartbeat{
				Submitter:        sample.AccAddress(),
				VoteValidatorSet: []uint64{1},
				VoteAggSignature: []byte{1, 2, 3},
			},
			err: ErrInvalidVoteAggSignature,
		}, {
			name: "valid message",
			msg: MsgHeartbeat{
				Submitter:        sample.AccAddress(),
				VoteValidatorSet: []uint64{1},
				VoteAggSignature: sig[:],
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
