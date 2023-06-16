package types

import (
	"testing"

	"github.com/bnb-chain/greenfield/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgCompleteStorageProviderExit_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCompleteStorageProviderExit
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCompleteStorageProviderExit{
				OperatorAddress: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCompleteStorageProviderExit{
				OperatorAddress:         sample.AccAddress(),
				OriginStorageProviderId: 1,
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
