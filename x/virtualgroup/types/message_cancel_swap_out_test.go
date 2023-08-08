package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

func TestMsgCancelSwapOut_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCancelSwapOut
		err  error
	}{
		{
			name: "valid address",
			msg: *NewMsgCancelSwapOut(
				sample.RandAccAddress(),
				1,
				[]uint32{},
			),
		},
		{
			name: "invalid address",
			msg: MsgCancelSwapOut{
				StorageProvider:            "invalid_address",
				GlobalVirtualGroupFamilyId: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid gvg groups",
			msg: MsgCancelSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{1, 2},
			},
			err: gnfderrors.ErrInvalidMessage,
		},
		{
			name: "invalid gvg groups",
			msg: MsgCancelSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
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
