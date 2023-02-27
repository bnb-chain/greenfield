package keeper

import "github.com/bnb-chain/greenfield/x/storage/types"

type CreateBucketOptions struct {
	IsPublic bool
	types.SourceType
	ReadQuota         types.ReadQuota
	PaymentAddress    string
	PrimarySpApproval *types.Approval
	ApprovalMsgBytes  []byte
}

type DeleteBucketOptions struct {
	types.SourceType
}

type UpdateBucketOptions struct {
	types.SourceType
	types.ReadQuota
	PaymentAddress string
}

type CreateObjectOptions struct {
	IsPublic    bool
	ContentType string
	types.SourceType
	types.RedundancyType
	Checksums            [][]byte
	SecondarySpAddresses []string
	PrimarySpApproval    *types.Approval
	ApprovalMsgBytes     []byte
}

type CancelCreateObjectOptions struct {
	types.SourceType
}

type DeleteObjectOptions struct {
	types.SourceType
}

type CopyObjectOptions struct {
	types.SourceType
	IsPublic          bool
	PrimarySpApproval *types.Approval
	ApprovalMsgBytes  []byte
}
type CreateGroupOptions struct {
	Members []string
	types.SourceType
}
type LeaveGroupOptions struct {
	types.SourceType
}

type UpdateGroupMemberOptions struct {
	types.SourceType
	MembersToAdd    []string
	MembersToDelete []string
}
