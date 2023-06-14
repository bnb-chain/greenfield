package keeper

import (
	sdkmath "cosmossdk.io/math"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

type CreateBucketOptions struct {
	Visibility        types.VisibilityType
	SourceType        types.SourceType
	ChargedReadQuota  uint64
	PaymentAddress    string
	PrimarySpApproval *types.Approval
	ApprovalMsgBytes  []byte
}

type DeleteBucketOptions struct {
	SourceType types.SourceType
}

type UpdateBucketOptions struct {
	Visibility       types.VisibilityType
	SourceType       types.SourceType
	PaymentAddress   string
	ChargedReadQuota *uint64
}

type CreateObjectOptions struct {
	Visibility           types.VisibilityType
	ContentType          string
	SourceType           types.SourceType
	RedundancyType       types.RedundancyType
	Checksums            [][]byte
	SecondarySpAddresses []string
	PrimarySpApproval    *types.Approval
	ApprovalMsgBytes     []byte
}

type CancelCreateObjectOptions struct {
	SourceType types.SourceType
}

type DeleteObjectOptions struct {
	SourceType types.SourceType
}

type CopyObjectOptions struct {
	SourceType        types.SourceType
	Visibility        types.VisibilityType
	PrimarySpApproval *types.Approval
	ApprovalMsgBytes  []byte
}
type CreateGroupOptions struct {
	Members    []string
	SourceType types.SourceType
	Extra      string
}
type LeaveGroupOptions struct {
	SourceType types.SourceType
}

type UpdateGroupMemberOptions struct {
	SourceType      types.SourceType
	MembersToAdd    []string
	MembersToDelete []string
}

type InvokeExecutionOptions struct {
	InputObjectIds []sdkmath.Uint
	MaxGas         sdkmath.Uint
	Method         string
	Params         []byte
}
