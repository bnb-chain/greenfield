package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
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
	IsPaymentAccountOwner(ctx sdk.Context, addr string, owner string) bool
	GetStoragePrice(ctx sdk.Context, params paymenttypes.StoragePriceParams) (price paymenttypes.StoragePrice, err error)
	ApplyFlowChanges(ctx sdk.Context, from string, flowChanges []paymenttypes.OutFlow) (err error)
	ApplyUserFlowsList(ctx sdk.Context, userFlows []paymenttypes.UserFlows) (err error)
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
}
