package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgAttest_ValidateBasic(t *testing.T) {
	var sig [96]byte
	tests := []struct {
		name string
		msg  MsgAttest
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgAttest{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid vote result",
			msg: MsgAttest{
				Creator:           sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
				VoteResult:        100,
			},
			err: ErrInvalidVoteResult,
		}, {
			name: "invalid vote result",
			msg: MsgAttest{
				Creator:           sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
				VoteResult:        ChallengeResultSucceed,
				VoteValidatorSet:  make([]uint64, 0),
			},
			err: ErrInvalidVoteValidatorSet,
		}, {
			name: "invalid vote aggregated signature",
			msg: MsgAttest{
				Creator:           sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
				VoteResult:        ChallengeResultSucceed,
				VoteValidatorSet:  []uint64{1},
				VoteAggSignature:  []byte{1, 2, 3},
			},
			err: ErrInvalidVoteAggSignature,
		}, {
			name: "valid message",
			msg: MsgAttest{
				Creator:           sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
				VoteResult:        ChallengeResultSucceed,
				VoteValidatorSet:  []uint64{1},
				VoteAggSignature:  sig[:],
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
