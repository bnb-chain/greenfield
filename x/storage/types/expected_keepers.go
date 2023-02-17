package types

import (
	sdkmath "cosmossdk.io/math"
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
}

type PaymentKeeper interface {
	IsPaymentAccountOwner(ctx sdk.Context, addr string, owner string) bool
	GetReadPrice(ctx sdk.Context, spAddr string, readQuota uint64, priceTime int64) (sdkmath.Int, error)
	ApplyUSDFlowChanges(ctx sdk.Context, from string, flowChanges []paymenttypes.OutFlowInUSD) (err error)
	GetParams(ctx sdk.Context) paymenttypes.Params
	UpdateStreamRecordByAddr(ctx sdk.Context, change *paymenttypes.StreamRecordChange) (ret *paymenttypes.StreamRecord, err error)
	GetStorePrice(ctx sdk.Context, bucketInfo *BucketInfo, objectInfo *ObjectInfo) paymenttypes.StorePrice
	//// TODO(owen): add a thin wrapper to storage module and only provide basic interface here.
	//ChargeUpdatePaymentAccount(ctx sdk.Context, bucketInfo *BucketInfo, paymentAddress *string) error
	//LockStoreFee(ctx sdk.Context, bucketInfo *BucketInfo, objectInfo *ObjectInfo) error
	//ChargeDeleteObject(ctx sdk.Context, bucketInfo *BucketInfo, objectInfo *ObjectInfo) error
	//UnlockAndChargeStoreFee(ctx sdk.Context, bucketInfo *BucketInfo, objectInfo *ObjectInfo) error
	//ChargeUpdateReadQuota(ctx sdk.Context, bucketInfo *BucketInfo, newReadPacket ReadQuota) error
	//UnlockStoreFee(ctx sdk.Context, bucketInfo *BucketInfo, objectInfo *ObjectInfo) error
	//ChargeInitialReadFee(ctx sdk.Context, bucketInfo *BucketInfo) error
}
