package types

import (
	"testing"

	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgCompleteMigrateBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCompleteMigrateBucket
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCompleteMigrateBucket{
				Operator:   "invalid_address",
				BucketName: "bucketname",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCompleteMigrateBucket{
				Operator:                   sample.AccAddress(),
				BucketName:                 "bucketname",
				GlobalVirtualGroupFamilyId: 1,
				GvgMappings:                []*GVGMapping{{1, 2, []byte("xxxxxxxxxxx")}},
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
