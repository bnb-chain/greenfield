package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgUpdateParams_ValidateBasic(t *testing.T) {

	wrongParams := DefaultParams()
	wrongParams.HeartbeatInterval = 0

	tests := []struct {
		name string
		msg  MsgUpdateParams
		err  error
	}{
		{
			name: "invalid authority",
			msg: MsgUpdateParams{
				Authority: "invalid_address",
				Params:    DefaultParams(),
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "invalid params",
			msg: MsgUpdateParams{
				Authority: sample.AccAddress(),
				Params:    wrongParams,
			},
			err: ErrInvalidParams,
		}, {
			name: "invalid authority and params",
			msg: MsgUpdateParams{
				Authority: sample.AccAddress(),
				Params:    DefaultParams(),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.err != nil {
				require.ErrorContains(t, err, tt.err.Error())
				return
			}
			require.NoError(t, err)
		})
	}
}
