package types

import (
	"testing"

	"github.com/bnb-chain/greenfield/testutil/sample"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"
)

func TestMsgCreateBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateBucket
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateBucket{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateBucket{
				Creator: sample.AccAddress(),
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

func TestMsgDeleteBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteBucket
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeleteBucket{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgDeleteBucket{
				Creator: sample.AccAddress(),
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

func TestMsgCreateObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateObject
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateObject{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateObject{
				Creator: sample.AccAddress(),
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

func TestMsgDeleteObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteObject
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeleteObject{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgDeleteObject{
				Creator: sample.AccAddress(),
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

func TestMsgCopyObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCopyObject
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCopyObject{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCopyObject{
				Creator: sample.AccAddress(),
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

func TestMsgSealObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSealObject
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgSealObject{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgSealObject{
				Creator: sample.AccAddress(),
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

func TestMsgRejectUnsealedObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRejectUnsealedObject
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgRejectUnsealedObject{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgRejectUnsealedObject{
				Creator: sample.AccAddress(),
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

func TestMsgCreateGroup_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateGroup
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgCreateGroup{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgCreateGroup{
				Creator: sample.AccAddress(),
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

func TestMsgDeleteGroup_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteGroup
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeleteGroup{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgDeleteGroup{
				Creator: sample.AccAddress(),
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

func TestMsgLeaveGroup_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgLeaveGroup
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgLeaveGroup{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgLeaveGroup{
				Creator: sample.AccAddress(),
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

func TestMsgUpdateGroupMember_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateGroupMember
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgUpdateGroupMember{
				Creator: "invalid_address",
			},
			err: sdkerrors.ErrInvalidAddress,
		}, {
			name: "valid address",
			msg: MsgUpdateGroupMember{
				Creator: sample.AccAddress(),
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
