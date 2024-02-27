package types

import (
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	types2 "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/common"
	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
	"github.com/bnb-chain/greenfield/x/permission/types"
)

var (
	testBucketName                      = "testbucket"
	testObjectName                      = "testobject"
	testGroupName                       = "testgroup"
	testInvalidBucketNameWithLongLength = [68]byte{}
)

func TestMsgCreateBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateBucket
		err  error
	}{
		{
			name: "normal",
			msg: MsgCreateBucket{
				Creator:           sample.RandAccAddressHex(),
				BucketName:        testBucketName,
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.RandAccAddressHex(),
				PrimarySpAddress:  sample.RandAccAddressHex(),
				PrimarySpApproval: &common.Approval{},
			},
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.RandAccAddressHex(),
				BucketName:        "TestBucket",
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.RandAccAddressHex(),
				PrimarySpAddress:  sample.RandAccAddressHex(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.RandAccAddressHex(),
				BucketName:        "Test-Bucket",
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.RandAccAddressHex(),
				PrimarySpAddress:  sample.RandAccAddressHex(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.RandAccAddressHex(),
				BucketName:        "ss",
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.RandAccAddressHex(),
				PrimarySpAddress:  sample.RandAccAddressHex(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.RandAccAddressHex(),
				BucketName:        string(testInvalidBucketNameWithLongLength[:]),
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.RandAccAddressHex(),
				PrimarySpAddress:  sample.RandAccAddressHex(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
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
			name: "normal",
			msg: MsgDeleteBucket{
				Operator:   sample.RandAccAddressHex(),
				BucketName: testBucketName,
			},
		}, {
			name: "invalid bucket name",
			msg: MsgDeleteBucket{
				Operator:   sample.RandAccAddressHex(),
				BucketName: "testBucket",
			},
			err: gnfderrors.ErrInvalidBucketName,
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

func TestMsgUpdateBucketInfo_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateBucketInfo
		err  error
	}{
		{
			name: "basic",
			msg: MsgUpdateBucketInfo{
				Operator:         sample.RandAccAddressHex(),
				BucketName:       testBucketName,
				PaymentAddress:   sample.RandAccAddressHex(),
				ChargedReadQuota: &common.UInt64Value{Value: 10000},
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
			name: "normal",
			msg: MsgCreateGroup{
				Creator:   sample.RandAccAddressHex(),
				GroupName: testGroupName,
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
			name: "normal",
			msg: MsgDeleteGroup{
				Operator:  sample.RandAccAddressHex(),
				GroupName: testGroupName,
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
			name: "normal",
			msg: MsgLeaveGroup{
				Member:     sample.RandAccAddressHex(),
				GroupOwner: sample.RandAccAddressHex(),
				GroupName:  testGroupName,
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
			name: "normal",
			msg: MsgUpdateGroupMember{
				Operator:   sample.RandAccAddressHex(),
				GroupOwner: sample.RandAccAddressHex(),
				GroupName:  testGroupName,
				MembersToAdd: []*MsgGroupMember{
					{
						Member: sample.RandAccAddressHex(),
					},
					{
						Member: sample.RandAccAddressHex(),
					},
				},
				MembersToDelete: []string{sample.RandAccAddressHex(), sample.RandAccAddressHex()},
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

func TestMsgUpdateGroupExtra_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateGroupExtra
		err  error
	}{
		{
			name: "normal",
			msg: MsgUpdateGroupExtra{
				Operator:   sample.RandAccAddressHex(),
				GroupOwner: sample.RandAccAddressHex(),
				GroupName:  testGroupName,
				Extra:      "testExtra",
			},
		},
		{
			name: "extra field is too long",
			msg: MsgUpdateGroupExtra{
				Operator:   sample.RandAccAddressHex(),
				GroupOwner: sample.RandAccAddressHex(),
				GroupName:  testGroupName,
				Extra:      strings.Repeat("abcdefg", 80),
			},
			err: gnfderrors.ErrInvalidParameter,
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

func TestMsgPutPolicy_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgPutPolicy
		err  error
	}{
		{
			name: "normal",
			msg: MsgPutPolicy{
				Operator:  sample.RandAccAddressHex(),
				Resource:  types2.NewBucketGRN(testBucketName).String(),
				Principal: types.NewPrincipalWithAccount(sdk.MustAccAddressFromHex(sample.RandAccAddressHex())),
				Statements: []*types.Statement{{
					Effect:  types.EFFECT_ALLOW,
					Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
				}},
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

func TestMsgDeletePolicy_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgDeletePolicy
		err  error
	}{
		{
			name: "valid address",
			msg: MsgDeletePolicy{
				Operator:  sample.RandAccAddressHex(),
				Resource:  types2.NewBucketGRN(testBucketName).String(),
				Principal: types.NewPrincipalWithAccount(sdk.MustAccAddressFromHex(sample.RandAccAddressHex())),
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

func TestMsgMirrorBucket_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMirrorBucket
		err  error
	}{
		{
			name: "normal",
			msg: MsgMirrorBucket{
				Operator: sample.RandAccAddressHex(),
				Id:       math.NewUint(1),
			},
		}, {
			name: "invalid account name",
			msg: MsgMirrorBucket{
				Operator: "wrong address",
				Id:       math.NewUint(1),
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

func TestMsgMirrorGroup_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMirrorGroup
		err  error
	}{
		{
			name: "normal",
			msg: MsgMirrorGroup{
				Operator: sample.RandAccAddressHex(),
				Id:       math.NewUint(1),
			},
		},
		{
			name: "invalid address",
			msg: MsgMirrorGroup{
				Operator: "wrong address",
				Id:       math.NewUint(1),
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

func TestMsgRenewGroupMember_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRenewGroupMember
		err  error
	}{
		{
			name: "normal",
			msg: MsgRenewGroupMember{
				Operator:   sample.RandAccAddressHex(),
				GroupOwner: sample.RandAccAddressHex(),
				GroupName:  testGroupName,
				Members: []*MsgGroupMember{
					{
						Member: sample.RandAccAddressHex(),
					},
					{
						Member: sample.RandAccAddressHex(),
					},
				},
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

func TestMsgSealObjetbytes(t *testing.T) {
	object := MsgSealObject{
		Operator:                    "abc",
		BucketName:                  "abc",
		ObjectName:                  "O",
		GlobalVirtualGroupId:        32,
		SecondarySpBlsAggSignatures: []byte{'a'},
	}
	fmt.Println(hex.EncodeToString(object.GetSignBytes()))
	//7b226275636b65745f6e616d65223a22616263222c22676c6f62616c5f7669727475616c5f67726f75705f6964223a33322c226f626a6563745f6e616d65223a224f222c226f70657261746f72223a22616263222c227365636f6e646172795f73705f626c735f6167675f7369676e617475726573223a2259513d3d227d
	//7b226275636b65745f6e616d65223a22616263222c226578706563745f636865636b73756d73223a5b5d2c22676c6f62616c5f7669727475616c5f67726f75705f6964223a33322c226f626a6563745f6e616d65223a224f222c226f70657261746f72223a22616263222c227365636f6e646172795f73705f626c735f6167675f7369676e617475726573223a2259513d3d227d
	//7b226275636b65745f6e616d65223a22616263222c226578706563745f636865636b73756d73223a5b5d2c22676c6f62616c5f7669727475616c5f67726f75705f6964223a33322c226f626a6563745f6e616d65223a224f222c226f70657261746f72223a22616263222c227365636f6e646172795f73705f626c735f6167675f7369676e617475726573223a2259513d3d227d
}
