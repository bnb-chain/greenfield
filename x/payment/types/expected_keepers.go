package types

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
	HasAccount(ctx context.Context, addr sdk.AccAddress) bool
	GetModuleAddress(name string) sdk.AccAddress
	GetModuleAccount(ctx context.Context, moduleName string) sdk.ModuleAccountI
	SetModuleAccount(context.Context, sdk.ModuleAccountI)
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// SpKeeper defines the expected interface needed to retrieve storage provider.
type SpKeeper interface {
	GetSpStoragePriceByTime(ctx sdk.Context, spAddr sdk.AccAddress, time int64) (val sptypes.SpStoragePrice, err error)
	GetSecondarySpStorePriceByTime(ctx sdk.Context, time int64) (val sptypes.SecondarySpStorePrice, err error)
}
