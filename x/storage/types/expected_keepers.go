package types

import (
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
}

type PaymentKeeper interface {
	GetVersionedParamsWithTs(ctx sdk.Context, time int64) (paymenttypes.VersionedParams, error)
	IsPaymentAccountOwner(ctx sdk.Context, addr, owner sdk.AccAddress) bool
	GetStoragePrice(ctx sdk.Context, params paymenttypes.StoragePriceParams) (price paymenttypes.StoragePrice, err error)
	ApplyUserFlowsList(ctx sdk.Context, userFlows []paymenttypes.UserFlows) (err error)
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
}

type PermissionKeeper interface {
	PutPolicy(ctx sdk.Context, policy *permtypes.Policy) (math.Uint, error)
	DeletePolicy(ctx sdk.Context, principal *permtypes.Principal, resourceType resource.ResourceType,
		resourceID math.Uint) (math.Uint, error)
	VerifyPolicy(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType, operator sdk.AccAddress,
		action permtypes.ActionType, opts *permtypes.VerifyOptions) permtypes.Effect
	AddGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (math.Uint, error)
	AddGroupMemberWithExpiration(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress, expiration time.Time) error
	UpdateGroupMemberExpiration(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress, memberID math.Uint, expiration time.Time)
	RemoveGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error
	RemoveGroupMemberExtra(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error
	GetPolicyByID(ctx sdk.Context, policyID math.Uint) (*permtypes.Policy, bool)
	GetPolicyForAccount(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) (policy *permtypes.Policy, isFound bool)
	GetPolicyForGroup(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType,
		groupID math.Uint) (policy *permtypes.Policy, isFound bool)
	GetGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (*permtypes.GroupMember, bool)
	GetGroupMemberByID(ctx sdk.Context, groupMemberID math.Uint) (*permtypes.GroupMember, bool)
	GetGroupMemberExtra(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) (*permtypes.GroupMemberExtra, bool)
	GetGroupMemberExtraByID(ctx sdk.Context, groupMemberID math.Uint) (*permtypes.GroupMemberExtra, bool)
	ForceDeleteAccountPolicyForResource(ctx sdk.Context, maxDelete, deletedCount uint64, resourceType resource.ResourceType, resourceID math.Uint) (uint64, bool)
	ForceDeleteGroupPolicyForResource(ctx sdk.Context, maxDelete, deletedCount uint64, resourceType resource.ResourceType, resourceID math.Uint) (uint64, bool)
	ForceDeleteGroupMembers(ctx sdk.Context, maxDelete, deletedTotal uint64, groupId math.Uint) (uint64, bool)
	ExistAccountPolicyForResource(ctx sdk.Context, resourceType resource.ResourceType, resourceID math.Uint) bool
	ExistGroupPolicyForResource(ctx sdk.Context, resourceType resource.ResourceType, resourceID math.Uint) bool
	ExistGroupMemberForGroup(ctx sdk.Context, groupId math.Uint) bool
}

type CrossChainKeeper interface {
	GetDestBscChainID() sdk.ChainID
	CreateRawIBCPackageWithFee(ctx sdk.Context, chainID sdk.ChainID, channelID sdk.ChannelID, packageType sdk.CrossChainPackageType,
		packageLoad []byte, relayerFee *big.Int, ackRelayerFee *big.Int,
	) (uint64, error)

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
}
