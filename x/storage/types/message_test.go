package types

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
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
				Creator:                    sample.AccAddress(),
				BucketName:                 testBucketName,
				IsPublic:                   true,
				PaymentAddress:             sample.AccAddress(),
				PrimarySpAddress:           sample.AccAddress(),
				PrimarySpApprovalSignature: []byte(""),
			},
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:                    sample.AccAddress(),
				BucketName:                 "TestBucket",
				IsPublic:                   true,
				PaymentAddress:             sample.AccAddress(),
				PrimarySpAddress:           sample.AccAddress(),
				PrimarySpApprovalSignature: []byte(""),
			},
			err: ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:                    sample.AccAddress(),
				BucketName:                 "Test-Bucket",
				IsPublic:                   true,
				PaymentAddress:             sample.AccAddress(),
				PrimarySpAddress:           sample.AccAddress(),
				PrimarySpApprovalSignature: []byte(""),
			},
			err: ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:                    sample.AccAddress(),
				BucketName:                 "ss",
				IsPublic:                   true,
				PaymentAddress:             sample.AccAddress(),
				PrimarySpAddress:           sample.AccAddress(),
				PrimarySpApprovalSignature: []byte(""),
			},
			err: ErrInvalidBucketName,
		}, {
			name: "invalid bucket name",
			msg: MsgCreateBucket{
				Creator:                    sample.AccAddress(),
				BucketName:                 string(testInvalidBucketNameWithLongLength[:]),
				IsPublic:                   true,
				PaymentAddress:             sample.AccAddress(),
				PrimarySpAddress:           sample.AccAddress(),
				PrimarySpApprovalSignature: []byte(""),
			},
			err: ErrInvalidBucketName,
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
			err: ErrInvalidBucketName,
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
				Operator:       sample.AccAddress(),
				BucketName:     testBucketName,
				PaymentAddress: sample.AccAddress(),
				ReadQuota:      READ_QUOTA_FREE,
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
				Creator:                    sample.AccAddress(),
				BucketName:                 testBucketName,
				ObjectName:                 testObjectName,
				PayloadSize:                1024,
				IsPublic:                   false,
				ContentType:                "content-type",
				PrimarySpApprovalSignature: sample.Checksum(),
				ExpectChecksums:            [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
				ExpectSecondarySpAddresses: []string{sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress()},
			},
		}, {
			name: "invalid object name",
			msg: MsgCreateObject{
				Creator:                    sample.AccAddress(),
				BucketName:                 testBucketName,
				ObjectName:                 "",
				PayloadSize:                1024,
				IsPublic:                   false,
				ContentType:                "content-type",
				PrimarySpApprovalSignature: sample.Checksum(),
				ExpectChecksums:            [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
				ExpectSecondarySpAddresses: []string{sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress()},
			},
			err: ErrInvalidObjectName,
		}, {
			name: "invalid object name",
			msg: MsgCreateObject{
				Creator:                    sample.AccAddress(),
				BucketName:                 testBucketName,
				ObjectName:                 "../object",
				PayloadSize:                1024,
				IsPublic:                   false,
				ContentType:                "content-type",
				PrimarySpApprovalSignature: sample.Checksum(),
				ExpectChecksums:            [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
				ExpectSecondarySpAddresses: []string{sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress()},
			},
			err: ErrInvalidObjectName,
		}, {
			name: "invalid object name",
			msg: MsgCreateObject{
				Creator:                    sample.AccAddress(),
				BucketName:                 testBucketName,
				ObjectName:                 "//object",
				PayloadSize:                1024,
				IsPublic:                   false,
				ContentType:                "content-type",
				PrimarySpApprovalSignature: sample.Checksum(),
				ExpectChecksums:            [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
				ExpectSecondarySpAddresses: []string{sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress()},
			},
			err: ErrInvalidObjectName,
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
			name: "normal",
			msg: MsgSealObject{
				Operator:              sample.AccAddress(),
				BucketName:            testBucketName,
				ObjectName:            testObjectName,
				SecondarySpAddresses:  []string{sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress(), sample.AccAddress()},
				SecondarySpSignatures: [][]byte{sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum(), sample.Checksum()},
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
