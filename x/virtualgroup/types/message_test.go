package types

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/types/common"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
)

func TestMsgDeposit_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeposit
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgDeposit{
				StorageProvider:      "invalid_address",
				GlobalVirtualGroupId: 1,
				Deposit: types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(1),
				},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid deposit amount",
			msg: MsgDeposit{
				StorageProvider:      sample.RandAccAddressHex(),
				GlobalVirtualGroupId: 1,
				Deposit: types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(0),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "valid case",
			msg: *NewMsgDeposit(
				sample.RandAccAddress(),
				1,
				types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(1),
				},
			),
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

func TestMsgWithdraw_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgWithdraw
		err  error
	}{
		{
			name: "invalid address",
			msg: MsgWithdraw{
				StorageProvider:      "invalid_address",
				GlobalVirtualGroupId: 1,
				Withdraw: types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(1),
				},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid deposit amount",
			msg: MsgWithdraw{
				StorageProvider:      sample.RandAccAddressHex(),
				GlobalVirtualGroupId: 1,
				Withdraw: types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(0),
				},
			},
			err: sdkerrors.ErrInvalidRequest,
		},
		{
			name: "valid case",
			msg: *NewMsgWithdraw(
				sample.RandAccAddress(),
				1,
				types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(1),
				},
			),
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

func TestMsgSwapOut_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSwapOut
		err  error
	}{
		{
			name: "valid case",
			msg: MsgSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 0,
				GlobalVirtualGroupIds:      []uint32{1, 2, 3},
				SuccessorSpId:              100,
				SuccessorSpApproval: &common.Approval{
					ExpiredHeight:              100,
					GlobalVirtualGroupFamilyId: 1,
					Sig:                        []byte("sig"),
				},
			},
		},
		{
			name: "valid case",
			msg: MsgSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{},
				SuccessorSpId:              100,
				SuccessorSpApproval: &common.Approval{
					ExpiredHeight:              100,
					GlobalVirtualGroupFamilyId: 1,
					Sig:                        []byte("sig"),
				},
			},
		},
		{
			name: "invalid address",
			msg: MsgSwapOut{
				StorageProvider:            "invalid address",
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{},
				SuccessorSpId:              100,
				SuccessorSpApproval: &common.Approval{
					ExpiredHeight:              100,
					GlobalVirtualGroupFamilyId: 1,
					Sig:                        []byte("sig"),
				},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid virtual group family",
			msg: MsgSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{1},
				SuccessorSpId:              100,
				SuccessorSpApproval: &common.Approval{
					ExpiredHeight:              100,
					GlobalVirtualGroupFamilyId: 1,
					Sig:                        []byte("sig"),
				},
			},
			err: gnfderrors.ErrInvalidMessage,
		},
		{
			name: "invalid virtual group family",
			msg: MsgSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 0,
				GlobalVirtualGroupIds:      []uint32{},
				SuccessorSpId:              100,
				SuccessorSpApproval: &common.Approval{
					ExpiredHeight:              100,
					GlobalVirtualGroupFamilyId: 1,
					Sig:                        []byte("sig"),
				},
			},
			err: gnfderrors.ErrInvalidMessage,
		},
		{
			name: "invalid successor sp id",
			msg: MsgSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{},
				SuccessorSpId:              0,
				SuccessorSpApproval: &common.Approval{
					ExpiredHeight:              100,
					GlobalVirtualGroupFamilyId: 1,
					Sig:                        []byte("sig"),
				},
			},
			err: gnfderrors.ErrInvalidMessage,
		},
		{
			name: "invalid successor sp approval",
			msg: MsgSwapOut{
				StorageProvider:            sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 1,
				GlobalVirtualGroupIds:      []uint32{},
				SuccessorSpId:              1,
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

func TestMsgCreateGlobalVirtualGroup_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateGlobalVirtualGroup
		err  error
	}{
		{
			name: "valid case",
			msg: *NewMsgCreateGlobalVirtualGroup(
				sample.RandAccAddress(),
				1,
				[]uint32{2, 3, 4},
				types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(1),
				},
			),
		},
		{
			name: "invalid address",
			msg: MsgCreateGlobalVirtualGroup{
				StorageProvider: "invalid_address",
				FamilyId:        1,
				SecondarySpIds:  []uint32{2, 3, 4},
				Deposit: types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(1),
				},
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid deposit coin",
			msg: MsgCreateGlobalVirtualGroup{
				StorageProvider: "invalid_address",
				FamilyId:        1,
				SecondarySpIds:  []uint32{2, 3, 4},
				Deposit: types.Coin{
					Denom:  "denom",
					Amount: types.NewInt(0),
				},
			},
			err: sdkerrors.ErrInvalidAddress,
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

func TestMsgDeleteGlobalVirtualGroup_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeleteGlobalVirtualGroup
		err  error
	}{
		{
			name: "valid case",
			msg: *NewMsgDeleteGlobalVirtualGroup(
				sample.RandAccAddress(),
				1,
			),
		},
		{
			name: "invalid address",
			msg: MsgDeleteGlobalVirtualGroup{
				StorageProvider:      "invalid_address",
				GlobalVirtualGroupId: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
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

func TestMsgSettle_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgSettle
		err  error
	}{
		{
			name: "valid case",
			msg: *NewMsgSettle(
				sample.RandAccAddress(),
				1,
				[]uint32{1, 2, 3, 4},
			),
		},
		{
			name: "invalid address",
			msg: MsgSettle{
				Submitter:                  "invalid_address",
				GlobalVirtualGroupFamilyId: 1,
			},
			err: sdkerrors.ErrInvalidAddress,
		},
		{
			name: "invalid gvg ids",
			msg: MsgSettle{
				Submitter:                  sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 0,
			},
			err: ErrInvalidGVGCount,
		},
		{
			name: "invalid gvg ids",
			msg: MsgSettle{
				Submitter:                  sample.RandAccAddressHex(),
				GlobalVirtualGroupFamilyId: 0,
				GlobalVirtualGroupIds:      []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11},
			},
			err: ErrInvalidGVGCount,
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
