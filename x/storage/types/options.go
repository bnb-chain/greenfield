package types

import (
	time "time"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/types/common"
)

type CreateBucketOptions struct {
	Visibility        VisibilityType
	SourceType        SourceType
	ChargedReadQuota  uint64
	PaymentAddress    string
	PrimarySpApproval *common.Approval
	ApprovalMsgBytes  []byte
}

type DeleteBucketOptions struct {
	SourceType SourceType
}

type UpdateBucketOptions struct {
	Visibility       VisibilityType
	SourceType       SourceType
	PaymentAddress   string
	ChargedReadQuota *uint64
}

type CreateObjectOptions struct {
	Visibility        VisibilityType
	ContentType       string
	SourceType        SourceType
	RedundancyType    RedundancyType
	Checksums         [][]byte
	PrimarySpApproval *common.Approval
	ApprovalMsgBytes  []byte
	Delegated         bool
	Creator           sdk.AccAddress
}

type CancelCreateObjectOptions struct {
	SourceType SourceType
}

type DeleteObjectOptions struct {
	SourceType SourceType
}

type CopyObjectOptions struct {
	SourceType        SourceType
	Visibility        VisibilityType
	PrimarySpApproval *common.Approval
	ApprovalMsgBytes  []byte
}

type CreateGroupOptions struct {
	SourceType SourceType
	Extra      string
}

type LeaveGroupOptions struct {
	SourceType SourceType
}

type UpdateGroupMemberOptions struct {
	SourceType             SourceType
	MembersToAdd           []string
	MembersExpirationToAdd []*time.Time
	MembersToDelete        []string
}

type RenewGroupMemberOptions struct {
	SourceType        SourceType
	Members           []string
	MembersExpiration []*time.Time
}

type DeleteGroupOptions struct {
	SourceType SourceType
}

type CreatePolicyOptions struct {
	SourceType SourceType
	Extra      string
	Tags       *ResourceTags
}

type DeletePolicyOptions struct {
	SourceType SourceType
}

type UpdateObjectOptions struct {
	ContentType string
	Checksums   [][]byte
	Delegated   bool
	Updater     sdk.AccAddress
}
