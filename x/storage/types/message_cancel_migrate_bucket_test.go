package types

import (
	"testing"

	"github.com/bnb-chain/greenfield/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgCancelMigrateBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCancelMigrateBucket
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCancelMigrateBucket{
				Operator:   "invalid_address",
				BucketName: testBucketName,
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCancelMigrateBucket{
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
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
