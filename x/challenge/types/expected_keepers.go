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
	Slash(ctx sdk.Context, spAcc sdk.AccAddress, rewardInfos []sp.RewardInfo) error
}

type StakingKeeper interface {
	GetLastValidators(ctx sdk.Context) (validators []staking.Validator)
	GetHistoricalInfo(ctx sdk.Context, height int64) (staking.HistoricalInfo, bool)
}

type StorageKeeper interface {
	GetObject(ctx sdk.Context, bucketName string, objectName string) (objectInfo storage.ObjectInfo, found bool)
	GetObjectWithKey(ctx sdk.Context, objectKey []byte) (objectInfo storage.ObjectInfo, found bool)
	GetObjectAfterKey(ctx sdk.Context, objectKey []byte) (objectInfo storage.ObjectInfo, found bool)
	GetBucket(ctx sdk.Context, bucketName string) (bucketInfo storage.BucketInfo, found bool)
	MaxSegmentSize(ctx sdk.Context) (res uint64)
}

type PaymentKeeper interface {
	QueryValidatorRewards(ctx sdk.Context) (res sdkmath.Int)
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) types.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	SendCoinsFromModuleToModule(ctx sdk.Context, senderModule, recipientModule string, amt sdk.Coins) error
}
