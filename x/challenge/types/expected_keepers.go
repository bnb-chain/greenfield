package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"

	sp "github.com/bnb-chain/greenfield/x/sp/types"
	storage "github.com/bnb-chain/greenfield/x/storage/types"
)

type SpKeeper interface {
	GetStorageProvider(ctx sdk.Context, addr sdk.AccAddress) (sp sp.StorageProvider, found bool)
	DepositDenomForSP(ctx sdk.Context) (res string)
	Slash(ctx sdk.Context, spAcc sdk.AccAddress, rewardInfos []sp.RewardInfo) error
}

type StakingKeeper interface {
	GetLastValidators(ctx sdk.Context) (validators []staking.Validator)
	GetHistoricalInfo(ctx sdk.Context, height int64) (staking.HistoricalInfo, bool)
}

type StorageKeeper interface {
	GetObjectInfo(ctx sdk.Context, bucketName string, objectName string) (objectInfo storage.ObjectInfo, found bool)
	GetObjectInfoById(ctx sdk.Context, objectId sdkmath.Uint) (objectInfo storage.ObjectInfo, found bool)
	GetObjectInfoCount(ctx sdk.Context) sdkmath.Uint
	GetBucketInfo(ctx sdk.Context, bucketName string) (bucketInfo storage.BucketInfo, found bool)
	MaxSegmentSize(ctx sdk.Context) (res uint64)
}

type PaymentKeeper interface {
	QueryValidatorRewards(ctx sdk.Context) (amount sdkmath.Int)
	TransferValidatorRewards(ctx sdk.Context, toAddr sdk.AccAddress, amount sdkmath.Int) error
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
