package types

import (
	context "context"
	"math/big"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/types/resource"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
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
	GetStorageProvider(ctx sdk.Context, addr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	GetStorageProviderBySealAddr(ctx sdk.Context, sealAddr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	GetStorageProviderByGcAddr(ctx sdk.Context, gcAddr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	IsStorageProviderExistAndInService(ctx sdk.Context, addr sdk.AccAddress) error
	SetSpStoragePrice(ctx sdk.Context, SpStoragePrice sptypes.SpStoragePrice)
	SetSecondarySpStorePrice(ctx sdk.Context, secondarySpStorePrice sptypes.SecondarySpStorePrice)
	GetSpStoragePriceByTime(ctx sdk.Context, spAddr sdk.AccAddress, time int64) (val sptypes.SpStoragePrice, err error)
}

type PaymentKeeper interface {
	GetParams(ctx sdk.Context) paymenttypes.Params
	IsPaymentAccountOwner(ctx sdk.Context, addr, owner sdk.AccAddress) bool
	GetStoragePrice(ctx sdk.Context, params paymenttypes.StoragePriceParams) (price paymenttypes.StoragePrice, err error)
	ApplyUserFlowsList(ctx sdk.Context, userFlows []paymenttypes.UserFlows) (err error)
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
	// GetStreamRecord
	GetStreamRecord(ctx sdk.Context, account sdk.AccAddress) (val *paymenttypes.StreamRecord, found bool)
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
}

type CrossChainKeeper interface {
	CreateRawIBCPackageWithFee(ctx sdk.Context, channelID sdk.ChannelID, packageType sdk.CrossChainPackageType,
		packageLoad []byte, relayerFee *big.Int, ackRelayerFee *big.Int,
	) (uint64, error)

	RegisterChannel(name string, id sdk.ChannelID, app sdk.CrossChainApplication) error
}
