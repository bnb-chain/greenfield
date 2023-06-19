package types

import (
	"strings"
	"testing"

	"github.com/prysmaticlabs/prysm/crypto/bls"

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
				Creator:           sample.AccAddress(),
				BucketName:        testBucketName,
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.AccAddress(),
				PrimarySpAddress:  sample.AccAddress(),
				PrimarySpApproval: &common.Approval{},
			},
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.AccAddress(),
				BucketName:        "TestBucket",
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.AccAddress(),
				PrimarySpAddress:  sample.AccAddress(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.AccAddress(),
				BucketName:        "Test-Bucket",
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.AccAddress(),
				PrimarySpAddress:  sample.AccAddress(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.AccAddress(),
				BucketName:        "ss",
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.AccAddress(),
				PrimarySpAddress:  sample.AccAddress(),
				PrimarySpApproval: &common.Approval{},
			},
			err: gnfderrors.ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:           sample.AccAddress(),
				BucketName:        string(testInvalidBucketNameWithLongLength[:]),
				Visibility:        VISIBILITY_TYPE_PUBLIC_READ,
				PaymentAddress:    sample.AccAddress(),
				PrimarySpAddress:  sample.AccAddress(),
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
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
			},
		}, {
			name: "invalid bucket name",
			msg: MsgDeleteBucket{
				Operator:   sample.AccAddress(),
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
				Operator:         sample.AccAddress(),
				BucketName:       testBucketName,
				PaymentAddress:   sample.AccAddress(),
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

func TestMsgCreateObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCreateObject
		err  error
	}{
		{
			name: "normal",
			msg: MsgCreateObject{
				Creator:           sample.AccAddress(),
				BucketName:        testBucketName,
				ObjectName:        testObjectName,
				PayloadSize:       1024,
				Visibility:        VISIBILITY_TYPE_PRIVATE,
				ContentType:       "content-type",
				PrimarySpApproval: &common.Approval{},
				ExpectChecksums:   [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
			},
		}, {
			name: "invalid object name",
			msg: MsgCreateObject{
				Creator:           sample.AccAddress(),
				BucketName:        testBucketName,
				ObjectName:        "",
				PayloadSize:       1024,
				Visibility:        VISIBILITY_TYPE_PRIVATE,
				ContentType:       "content-type",
				PrimarySpApproval: &common.Approval{},
				ExpectChecksums:   [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
			},
			err: gnfderrors.ErrInvalidObjectName,
		}, {
			name: "invalid object name",
			msg: MsgCreateObject{
				Creator:           sample.AccAddress(),
				BucketName:        testBucketName,
				ObjectName:        "../object",
				PayloadSize:       1024,
				Visibility:        VISIBILITY_TYPE_PRIVATE,
				ContentType:       "content-type",
				PrimarySpApproval: &common.Approval{},
				ExpectChecksums:   [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
			},
			err: gnfderrors.ErrInvalidObjectName,
		}, {
			name: "invalid object name",
			msg: MsgCreateObject{
				Creator:           sample.AccAddress(),
				BucketName:        testBucketName,
				ObjectName:        "//object",
				PayloadSize:       1024,
				Visibility:        VISIBILITY_TYPE_PRIVATE,
				ContentType:       "content-type",
				PrimarySpApproval: &common.Approval{},
				ExpectChecksums:   [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
			},
			err: gnfderrors.ErrInvalidObjectName,
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

func TestMsgCancelCreateObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgCancelCreateObject
		err  error
	}{
		{
			name: "basic",
			msg: MsgCancelCreateObject{
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
				ObjectName: testObjectName,
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
			name: "normal",
			msg: MsgDeleteObject{
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
				ObjectName: testObjectName,
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
			name: "valid address",
			msg: MsgCopyObject{
				Operator:      sample.AccAddress(),
				SrcBucketName: testBucketName,
				SrcObjectName: testObjectName,
				DstBucketName: "dst" + testBucketName,
				DstObjectName: "dst" + testObjectName,
				DstPrimarySpApproval: &common.Approval{
					ExpiredHeight: 100,
					Sig:           []byte("xxx"),
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

func TestMsgSealObject_ValidateBasic(t *testing.T) {
	checksums := [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()}
	blsSignDoc := NewSecondarySpSealObjectSignDoc(math.NewUint(1), 1, GenerateHash(checksums[:])).GetSignBytes()
	blsPrivKey, _ := bls.RandKey()
	aggSig := blsPrivKey.Sign(blsSignDoc[:]).Marshal()
	tests := []struct {
		name string
		msg  MsgSealObject
		err  error
	}{
		{
			name: "normal",
			msg: MsgSealObject{
				Operator:                    sample.AccAddress(),
				BucketName:                  testBucketName,
				ObjectName:                  testObjectName,
				SecondarySpBlsAggSignatures: aggSig,
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

func TestMsgRejectSealObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgRejectSealObject
		err  error
	}{
		{
			name: "normal",
			msg: MsgRejectSealObject{
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
				ObjectName: testObjectName,
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

func TestMsgUpdateObjectInfo_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgUpdateObjectInfo
		err  error
	}{
		{
			name: "normal",
			msg: MsgUpdateObjectInfo{
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
				ObjectName: testObjectName,
				Visibility: VISIBILITY_TYPE_INHERIT,
			},
		},
		{
			name: "abnormal",
			msg: MsgUpdateObjectInfo{
				Operator:   sample.AccAddress(),
				BucketName: testBucketName,
				ObjectName: testObjectName,
				Visibility: VISIBILITY_TYPE_UNSPECIFIED,
			},
			err: ErrInvalidVisibility,
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
				Creator:   sample.AccAddress(),
				GroupName: testGroupName,
				Members:   []string{sample.AccAddress(), sample.AccAddress()},
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
				Operator:  sample.AccAddress(),
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
				Member:     sample.AccAddress(),
				GroupOwner: sample.AccAddress(),
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
				Operator:        sample.AccAddress(),
				GroupOwner:      sample.AccAddress(),
				GroupName:       testGroupName,
				MembersToAdd:    []string{sample.AccAddress(), sample.AccAddress()},
				MembersToDelete: []string{sample.AccAddress(), sample.AccAddress()},
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
				Operator:   sample.AccAddress(),
				GroupOwner: sample.AccAddress(),
				GroupName:  testGroupName,
				Extra:      "testExtra",
			},
		},
		{
			name: "extra field is too long",
			msg: MsgUpdateGroupExtra{
				Operator:   sample.AccAddress(),
				GroupOwner: sample.AccAddress(),
				GroupName:  testGroupName,
				Extra:      strings.Repeat("abcdefg", 80),
			},
			err: ErrInvalidParameter,
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
				Operator:  sample.AccAddress(),
				Resource:  types2.NewBucketGRN(testBucketName).String(),
				Principal: types.NewPrincipalWithAccount(sdk.MustAccAddressFromHex(sample.AccAddress())),
				Statements: []*types.Statement{{Effect: types.EFFECT_ALLOW,
					Actions: []types.ActionType{types.ACTION_DELETE_BUCKET}}},
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
				Operator:  sample.AccAddress(),
				Resource:  types2.NewBucketGRN(testBucketName).String(),
				Principal: types.NewPrincipalWithAccount(sdk.MustAccAddressFromHex(sample.AccAddress())),
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
				Operator: sample.AccAddress(),
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

func TestMsgMirrorObject_ValidateBasic(t *testing.T) {
	tests := []struct {
		name string
		msg  MsgMirrorObject
		err  error
	}{
		{
			name: "normal",
			msg: MsgMirrorObject{
				Operator: sample.AccAddress(),
				Id:       math.NewUint(1),
			},
		},
		{
			name: "invalid address",
			msg: MsgMirrorObject{
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
				Operator: sample.AccAddress(),
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
