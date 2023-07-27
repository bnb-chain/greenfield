package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

func TestMsgCompleteSwapOut_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCompleteSwapOut
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCompleteSwapOut{
				StorageProvider:            "invalid_address",
				GlobalVirtualGroupFamilyId: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "valid address",
			msg: MsgCompleteSwapOut{
				StorageProvider:            sample.AccAddress(),
				GlobalVirtualGroupFamilyId: 1,
			},
		},
		{
			name: "invalid gvg groups",
			msg: MsgCompleteSwapOut{
				StorageProvider:            sample.AccAddress(),
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{1, 2, 3},
			},
			err: gnfderrors.ErrInvalidMessage,
		},
		{
			name: "invalid gvg groups",
			msg: MsgCompleteSwapOut{
				StorageProvider:            sample.AccAddress(),
				GlobalVirtualGroupFamilyId: 0,
				GlobalVirtualGroupIds:      []uint32{},
			},
			err: gnfderrors.ErrInvalidMessage,
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
