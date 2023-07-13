package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgCancelSwapOut_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCancelSwapOut
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCancelSwapOut{
				StorageProvider:            "invalid_address",
				GlobalVirtualGroupFamilyId: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCancelSwapOut{
				StorageProvider:            sample.AccAddress(),
				GlobalVirtualGroupFamilyId: 1,
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
