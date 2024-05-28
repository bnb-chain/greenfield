package types

import (
	"context"
	"math/big"
	time "time"

	"cosmossdk.io/math"
	sdkmath "cosmossdk.io/math"
	"github.com/cometbft/cometbft/libs/log"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bnb-chain/greenfield/types/resource"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	GetModuleAddress(name string) sdk.AccAddress
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	// Methods imported from bank should be defined here
}

type SpKeeper interface {
	GetStorageProvider(ctx sdk.Context, id uint32) (*sptypes.StorageProvider, bool)
	MustGetStorageProvider(ctx sdk.Context, id uint32) *sptypes.StorageProvider
	GetStorageProviderByOperatorAddr(ctx sdk.Context, addr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	GetStorageProviderBySealAddr(ctx sdk.Context, sealAddr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	GetStorageProviderByGcAddr(ctx sdk.Context, gcAddr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	GetGlobalSpStorePriceByTime(ctx sdk.Context, time int64) (val sptypes.GlobalSpStorePrice, err error)
}

type PaymentKeeper interface {
	GetVersionedParamsWithTs(ctx sdk.Context, time int64) (paymenttypes.VersionedParams, error)
	IsPaymentAccountOwner(ctx sdk.Context, addr, owner sdk.AccAddress) bool
	ApplyUserFlowsList(ctx sdk.Context, userFlows []paymenttypes.UserFlows) (err error)
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
	GetStreamRecord(ctx sdk.Context, account sdk.AccAddress) (ret *paymenttypes.StreamRecord, found bool)
	MergeOutFlows(flows []paymenttypes.OutFlow) []paymenttypes.OutFlow
	GetAllStreamRecord(ctx sdk.Context) (list []paymenttypes.StreamRecord)
	GetOutFlows(ctx sdk.Context, addr sdk.AccAddress) []paymenttypes.OutFlow
}

type PermissionKeeper interface {
	PutPolicy(ctx sdk.Context, policy *permtypes.Policy) (math.Uint, error)
	DeletePolicy(ctx sdk.Context, principal *permtypes.Principal, resourceType resource.ResourceType,
		resourceID math.Uint) (math.Uint, error)
	AddGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress, expiration *time.Time) error
	UpdateGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress, memberID math.Uint, expiration *time.Time)
	MustGetPolicyByID(ctx sdk.Context, policyID math.Uint) *permtypes.Policy
	GetPolicyGroupForResource(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType) (*permtypes.PolicyGroup, bool)
	RemoveGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error
	GetPolicyByID(ctx sdk.Context, policyID math.Uint) (*permtypes.Policy, bool)
	GetPolicyForAccount(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) (policy *permtypes.Policy, isFound bool)
	GetPolicyForGroup(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType,
		groupID math.Uint) (policy *permtypes.Policy, isFound bool)
	GetGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (*permtypes.GroupMember, bool)
	GetGroupMemberByID(ctx sdk.Context, groupMemberID math.Uint) (*permtypes.GroupMember, bool)
	ForceDeleteAccountPolicyForResource(ctx sdk.Context, maxDelete, deletedCount uint64, resourceType resource.ResourceType, resourceID math.Uint) (uint64, bool)
	ForceDeleteGroupPolicyForResource(ctx sdk.Context, maxDelete, deletedCount uint64, resourceType resource.ResourceType, resourceID math.Uint) (uint64, bool)
	ForceDeleteGroupMembers(ctx sdk.Context, maxDelete, deletedTotal uint64, groupId math.Uint) (uint64, bool)
	ExistAccountPolicyForResource(ctx sdk.Context, resourceType resource.ResourceType, resourceID math.Uint) bool
	ExistGroupPolicyForResource(ctx sdk.Context, resourceType resource.ResourceType, resourceID math.Uint) bool
	ExistGroupMemberForGroup(ctx sdk.Context, groupId math.Uint) bool
}

type CrossChainKeeper interface {
	GetDestBscChainID() sdk.ChainID
	GetDestOpChainID() sdk.ChainID

	CreateRawIBCPackageWithFee(ctx sdk.Context, chainID sdk.ChainID, channelID sdk.ChannelID, packageType sdk.CrossChainPackageType,
		packageLoad []byte, relayerFee *big.Int, ackRelayerFee *big.Int,
	) (uint64, error)

	IsDestChainSupported(chainID sdk.ChainID) bool

	RegisterChannel(name string, id sdk.ChannelID, app sdk.CrossChainApplication) error
}

type VirtualGroupKeeper interface {
	SetGVGAndEmitUpdateEvent(ctx sdk.Context, gvg *types.GlobalVirtualGroup) error
	GetGVGFamily(ctx sdk.Context, familyID uint32) (*types.GlobalVirtualGroupFamily, bool)
	GetGVG(ctx sdk.Context, gvgID uint32) (*types.GlobalVirtualGroup, bool)
	SettleAndDistributeGVGFamily(ctx sdk.Context, sp *sptypes.StorageProvider, family *types.GlobalVirtualGroupFamily) error
	SettleAndDistributeGVG(ctx sdk.Context, gvg *types.GlobalVirtualGroup) error
	GetAndCheckGVGFamilyAvailableForNewBucket(ctx sdk.Context, familyID uint32) (*types.GlobalVirtualGroupFamily, error)
	GetGlobalVirtualGroupIfAvailable(ctx sdk.Context, gvgID uint32, expectedStoreSize uint64) (*types.GlobalVirtualGroup, error)
	GetSwapInInfo(ctx sdk.Context, familyID, gvgID uint32) (*types.SwapInInfo, bool)
}

// StorageKeeper used by the cross-chain applications
type StorageKeeper interface {
	Logger(ctx sdk.Context) log.Logger
	GetBucketInfoById(ctx sdk.Context, bucketId sdkmath.Uint) (*BucketInfo, bool)
	SetBucketInfo(ctx sdk.Context, bucketInfo *BucketInfo)
	CreateBucket(
		ctx sdk.Context, ownerAcc sdk.AccAddress, bucketName string,
		primarySpAcc sdk.AccAddress, opts *CreateBucketOptions) (sdkmath.Uint, error)
	DeleteBucket(ctx sdk.Context, operator sdk.AccAddress, bucketName string, opts DeleteBucketOptions) error
	GetGroupInfoById(ctx sdk.Context, groupId sdkmath.Uint) (*GroupInfo, bool)
	GetGroupInfo(ctx sdk.Context, ownerAddr sdk.AccAddress, groupName string) (*GroupInfo, bool)
	DeleteGroup(ctx sdk.Context, operator sdk.AccAddress, groupName string, opts DeleteGroupOptions) error
	CreateGroup(
		ctx sdk.Context, owner sdk.AccAddress,
		groupName string, opts CreateGroupOptions) (sdkmath.Uint, error)
	SetGroupInfo(ctx sdk.Context, groupInfo *GroupInfo)
	UpdateGroupMember(ctx sdk.Context, operator sdk.AccAddress, groupInfo *GroupInfo, opts UpdateGroupMemberOptions) error
	RenewGroupMember(ctx sdk.Context, operator sdk.AccAddress, groupInfo *GroupInfo, opts RenewGroupMemberOptions) error
	GetObjectInfoById(ctx sdk.Context, objectId sdkmath.Uint) (*ObjectInfo, bool)
	SetObjectInfo(ctx sdk.Context, objectInfo *ObjectInfo)
	DeleteObject(
		ctx sdk.Context, operator sdk.AccAddress, bucketName, objectName string, opts DeleteObjectOptions) error
	GetSourceTypeByChainId(ctx sdk.Context, chainId sdk.ChainID) (SourceType, error)

	NormalizePrincipal(ctx sdk.Context, principal *permtypes.Principal)
	ValidatePrincipal(ctx sdk.Context, resOwner sdk.AccAddress, principal *permtypes.Principal) error
}

type PaymentMsgServer interface {
	CreatePaymentAccount(context.Context, *paymenttypes.MsgCreatePaymentAccount) (*paymenttypes.MsgCreatePaymentAccountResponse, error)
	Deposit(context.Context, *paymenttypes.MsgDeposit) (*paymenttypes.MsgDepositResponse, error)
	Withdraw(context.Context, *paymenttypes.MsgWithdraw) (*paymenttypes.MsgWithdrawResponse, error)
	DisableRefund(context.Context, *paymenttypes.MsgDisableRefund) (*paymenttypes.MsgDisableRefundResponse, error)
}

type StorageMsgServer interface {
	UpdateBucketInfo(context.Context, *MsgUpdateBucketInfo) (*MsgUpdateBucketInfoResponse, error)
	ToggleSPAsDelegatedAgent(context.Context, *MsgToggleSPAsDelegatedAgent) (*MsgToggleSPAsDelegatedAgentResponse, error)
	CopyObject(context.Context, *MsgCopyObject) (*MsgCopyObjectResponse, error)
	UpdateObjectInfo(context.Context, *MsgUpdateObjectInfo) (*MsgUpdateObjectInfoResponse, error)
	UpdateGroupExtra(context.Context, *MsgUpdateGroupExtra) (*MsgUpdateGroupExtraResponse, error)
	MigrateBucket(context.Context, *MsgMigrateBucket) (*MsgMigrateBucketResponse, error)
	CancelMigrateBucket(context.Context, *MsgCancelMigrateBucket) (*MsgCancelMigrateBucketResponse, error)
	SetTag(context.Context, *MsgSetTag) (*MsgSetTagResponse, error)
	SetBucketFlowRateLimit(context.Context, *MsgSetBucketFlowRateLimit) (*MsgSetBucketFlowRateLimitResponse, error)
}
