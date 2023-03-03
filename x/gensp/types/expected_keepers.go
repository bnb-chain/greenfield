package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	bankexported "github.com/cosmos/cosmos-sdk/x/bank/exported"
	abci "github.com/tendermint/tendermint/abci/types"

	"encoding/json"

	auth "github.com/cosmos/cosmos-sdk/x/auth/types"
)

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
	NewAccount(sdk.Context, auth.AccountI) auth.AccountI
	SetAccount(sdk.Context, auth.AccountI)
	IterateAccounts(ctx sdk.Context, process func(auth.AccountI) (stop bool))
	// Methods imported from account should be defined here
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	// Methods imported from bank should be defined here
}

// StakingKeeper defines the expected staking keeper (noalias)
type StakingKeeper interface {
	ApplyAndReturnValidatorSetUpdates(sdk.Context) (updates []abci.ValidatorUpdate, err error)
}

// GenesisAccountsIterator defines the expected iterating genesis accounts object (noalias)
type GenesisAccountsIterator interface {
	IterateGenesisAccounts(
		cdc *codec.LegacyAmino,
		appGenesis map[string]json.RawMessage,
		cb func(auth.AccountI) (stop bool),
	)
}

// GenesisAccountsIterator defines the expected iterating genesis accounts object (noalias)
type GenesisBalancesIterator interface {
	IterateGenesisBalances(
		cdc codec.JSONCodec,
		appGenesis map[string]json.RawMessage,
		cb func(bankexported.GenesisBalance) (stop bool),
	)
}
