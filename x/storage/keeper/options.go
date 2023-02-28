package keeper

import (
	"github.com/bnb-chain/greenfield/x/storage/types"
)

type CreateBucketOptions struct {
	IsPublic          bool
	SourceType        types.SourceType
	ReadQuota         types.ReadQuota
	PaymentAddress    string
	PrimarySpApproval *types.Approval
	ApprovalMsgBytes  []byte
}

type DeleteBucketOptions struct {
	SourceType types.SourceType
}

type UpdateBucketOptions struct {
	SourceType     types.SourceType
	ReadQuota      types.ReadQuota
	PaymentAddress string
}

type CreateObjectOptions struct {
	IsPublic             bool
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
	IsPublic          bool
	PrimarySpApproval *types.Approval
	ApprovalMsgBytes  []byte
}
type CreateGroupOptions struct {
	Members    []string
	SourceType types.SourceType
}
type LeaveGroupOptions struct {
	SourceType types.SourceType
}

type UpdateGroupMemberOptions struct {
	SourceType      types.SourceType
	MembersToAdd    []string
	MembersToDelete []string
}
