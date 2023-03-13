package types

import (
	"cosmossdk.io/math"
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
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

type SpKeeper interface {
	GetStorageProvider(ctx sdk.Context, addr sdk.AccAddress) (sp sptypes.StorageProvider, found bool)
	GetStorageProviderBySealAddr(ctx sdk.Context, sealAddr sdk.AccAddress) (sp sptypes.StorageProvider, found bool)
	IsStorageProviderExistAndInService(ctx sdk.Context, addr sdk.AccAddress) error
}

type PaymentKeeper interface {
	GetParams(ctx sdk.Context) paymenttypes.Params
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
		action permtypes.ActionType, resource *string) permtypes.Effect
	AddGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress) error
	RemoveGroupMember(ctx sdk.Context, groupID math.Uint, member sdk.AccAddress)
	GetPolicyByID(ctx sdk.Context, policyID math.Uint) (*permtypes.Policy, bool)
	GetPolicyForAccount(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType, addr sdk.AccAddress) (policy *permtypes.Policy, isFound bool)
	GetPolicyForGroup(ctx sdk.Context, resourceID math.Uint, resourceType resource.ResourceType,
		groupID math.Uint) (policy *permtypes.Policy, isFound bool)
}
