package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
	"github.com/bnb-chain/bfs/testutil/sample"
)

func TestMsgMockCreateBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMockCreateBucket
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgMockCreateBucket{
				Operator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgMockCreateBucket{
				Operator: sample.AccAddress(),
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
