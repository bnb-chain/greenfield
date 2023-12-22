package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
)

func TestMsgStorageProviderForcedExit_ValidateBasic(t *testing.T) {
	tests := []struct {
		name      string
		msg       MsgStorageProviderForcedExit
		expErr    bool
		expErrMsg string
	}{
		{
			name: "invalid authority",
			msg: MsgStorageProviderForcedExit{
				Authority: "invalid authority address",
			},
			expErr:    true,
			expErrMsg: "invalid authority address",
		},
		{
			name: "invalid address",
			msg: MsgStorageProviderForcedExit{
				Authority:       "0xaE4F00015B40eE402a7f05E46757c18Df86E49E1",
				StorageProvider: "invalid_address",
			},
			expErr:    true,
			expErrMsg: "invalid address",
		}, {
			name: "valid address",
			msg:  *NewMsgStorageProviderForcedExit(sample.RandAccAddress().String(), sample.RandAccAddress()),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.msg.ValidateBasic()
			if tt.expErr {
				require.Contains(t, err.Error(), tt.expErrMsg)
				return
			}
			require.NoError(t, err)
		})
	}
}
