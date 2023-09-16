package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	staking "github.com/cosmos/cosmos-sdk/x/staking/types"

	sp "github.com/bnb-chain/greenfield/x/sp/types"
	storage "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type SpKeeper interface {
	GetStorageProvider(ctx sdk.Context, id uint32) (*sp.StorageProvider, bool)
	GetStorageProviderByOperatorAddr(ctx sdk.Context, opAddr sdk.AccAddress) (sp *sp.StorageProvider, found bool)
	DepositDenomForSP(ctx sdk.Context) (res string)
	Slash(ctx sdk.Context, spID uint32, rewardInfos []sp.RewardInfo) error
}

type StakingKeeper interface {
	GetLastValidators(ctx sdk.Context) (validators []staking.Validator)
	GetHistoricalInfo(ctx sdk.Context, height int64) (staking.HistoricalInfo, bool)
}

type StorageKeeper interface {
	GetObjectInfo(ctx sdk.Context, bucketName string, objectName string) (*storage.ObjectInfo, bool)
	GetObjectInfoById(ctx sdk.Context, objectId sdkmath.Uint) (*storage.ObjectInfo, bool)
	GetObjectInfoCount(ctx sdk.Context) sdkmath.Uint
	GetBucketInfo(ctx sdk.Context, bucketName string) (*storage.BucketInfo, bool)
	MaxSegmentSize(ctx sdk.Context) (res uint64)
	MaxSegmentSizeAtTime(ctx sdk.Context, timestamp int64) (res uint64, err error)
	GetObjectGVG(ctx sdk.Context, bucketID sdkmath.Uint, lvgID uint32) (*types.GlobalVirtualGroup, bool)
	MustGetPrimarySPForBucket(ctx sdk.Context, bucketInfo *storage.BucketInfo) *sp.StorageProvider
}

type PaymentKeeper interface {
	QueryDynamicBalance(ctx sdk.Context, addr sdk.AccAddress) (amount sdkmath.Int, err error)
	Withdraw(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, amount sdkmath.Int) error
}

// AccountKeeper defines the expected account keeper used for simulations (noalias)
type AccountKeeper interface {
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) authtypes.AccountI
}

// BankKeeper defines the expected interface needed to retrieve account balances.
type BankKeeper interface {
	SpendableCoins(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}
