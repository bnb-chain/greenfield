package types

import (
	sp "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"
)

type SpKeeper interface {
	GetStorageProvider(ctx sdk.Context, addr sdk.AccAddress) (sp sp.StorageProvider, found bool)
	Slash(ctx sdk.Context, spAcc sdk.AccAddress, rewardInfos []sp.RewardInfo) error
}

type StakingKeeper interface {
	GetLastValidators(ctx sdk.Context) (validators []staking.Validator)
	GetHistoricalInfo(ctx sdk.Context, height int64) (staking.HistoricalInfo, bool)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
