package types

import (
	"testing"

	"github.com/bnb-chain/bfs/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgMockSealObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMockSealObject
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgMockSealObject{
				Operator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgMockSealObject{
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
