package types

import (
	"math/big"

	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/bnb-chain/greenfield/types/resource"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
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
	IsStorageProviderExistAndInService(ctx sdk.Context, addr sdk.AccAddress) error
	SetSpStoragePrice(ctx sdk.Context, SpStoragePrice sptypes.SpStoragePrice)
	SetSecondarySpStorePrice(ctx sdk.Context, secondarySpStorePrice sptypes.SecondarySpStorePrice)
	GetSpStoragePriceByTime(ctx sdk.Context, spAddr sdk.AccAddress, time int64) (val sptypes.SpStoragePrice, err error)
}

type PaymentKeeper interface {
	GetParams(ctx sdk.Context) paymenttypes.Params
	GetVersionedParamsWithTs(ctx sdk.Context, time int64) (paymenttypes.VersionedParams, error)
	IsPaymentAccountOwner(ctx sdk.Context, addr, owner sdk.AccAddress) bool
	GetStoragePrice(ctx sdk.Context, params paymenttypes.StoragePriceParams) (price paymenttypes.StoragePrice, err error)
	ApplyUserFlowsList(ctx sdk.Context, userFlows []paymenttypes.UserFlows) (err error)
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
	GetStreamRecord(ctx sdk.Context, account sdk.AccAddress) (val *paymenttypes.StreamRecord, found bool)
	UpdateStreamRecord(ctx sdk.Context, streamRecord *paymenttypes.StreamRecord, change *paymenttypes.StreamRecordChange) error
}

type PermissionKeeper interface {
	PutPolicy(ctx sdk.Context, policy *permtypes.Policy) (math.Uint, error)
	DeletePolicy(ctx sdk.Context, principal *permtypes.Principal, resourceType resource.ResourceType,
		resourceID math.Uint) (math.Uint, error)
	VerifyPolicy(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType, operator sdk.AccAddress,
		action permtypes.ActionType, opts *permtypes.VerifyOptions) permtypes.Effect
	AddGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error
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
	CreateRawIBCPackageWithFee(ctx sdk.Context, channelID sdk.ChannelID, packageType sdk.CrossChainPackageType,
		packageLoad []byte, relayerFee *big.Int, ackRelayerFee *big.Int) (uint64, error)

	RegisterChannel(name string, id sdk.ChannelID, app sdk.CrossChainApplication) error
}

type VirtualGroupKeeper interface {
	BindingEmptyObjectToGVG(ctx sdk.Context, bucketID math.Uint, primarySPID, familyID uint32) (*types.LocalVirtualGroup, error)
	BindingObjectToGVG(ctx sdk.Context, bucketID math.Uint, primarySPID, familyID, gvgID uint32, payloadSize uint64) (*types.LocalVirtualGroup, error)
	UnBindingObjectFromLVG(ctx sdk.Context, bucketID math.Uint, lvgID uint32, payloadSize uint64) error
	UnBindingBucketFromGVG(ctx sdk.Context, bucketID math.Uint) error
	GetGVGFamily(ctx sdk.Context, spID, familyID uint32) (*types.GlobalVirtualGroupFamily, bool)
	GetGVG(ctx sdk.Context, gvgID uint32) (*types.GlobalVirtualGroup, bool)
	GetLVG(ctx sdk.Context, bucketID math.Uint, lvgID uint32) (*types.LocalVirtualGroup, bool)
	SettleAndDistributeGVGFamily(ctx sdk.Context, spID uint32, family *types.GlobalVirtualGroupFamily) error
	SettleAndDistributeGVG(ctx sdk.Context, gvg *types.GlobalVirtualGroup) error
	RebindingGVGsToBucket(ctx sdk.Context, bucketID math.Uint, dstSP *sptypes.StorageProvider, newLVGToGVGMappings []*LVGToGVGMapping) error
}
