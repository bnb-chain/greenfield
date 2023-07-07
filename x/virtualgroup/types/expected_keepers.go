package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

type SpKeeper interface {
	GetStorageProvider(ctx sdk.Context, id uint32) (sp *sptypes.StorageProvider, found bool)
	GetStorageProviderByOperatorAddr(ctx sdk.Context, addr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	GetStorageProviderByFundingAddr(ctx sdk.Context, sealAddr sdk.AccAddress) (sp *sptypes.StorageProvider, found bool)
	SetStorageProvider(ctx sdk.Context, sp *sptypes.StorageProvider)
	Exit(ctx sdk.Context, sp *sptypes.StorageProvider) error
	DepositDenomForSP(ctx sdk.Context) (res string)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
// Methods imported from account should be defined here
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
	IterateAccounts(ctx sdk.Context, process func(authtypes.AccountI) (stop bool))
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx sdk.Context, moduleName string) authtypes.ModuleAccountI
	SetModuleAccount(sdk.Context, authtypes.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
// Methods imported from bank should be defined here
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	LockedCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoins(ctx sdk.Context, fromAddr sdk.AccAddress, toAddr sdk.AccAddress, amt sdk.Coins) error
}

type PaymentKeeper interface {
	QueryDynamicBalance(ctx sdk.Context, addr sdk.AccAddress) (amount sdkmath.Int, err error)
	Withdraw(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amount sdkmath.Int) error
	IsEmptyNetFlow(ctx sdk.Context, account sdk.AccAddress) bool
}
