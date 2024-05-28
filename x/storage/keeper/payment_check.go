package keeper

import (
	sdkmath "cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/internal/sequence"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/pkg/errors"
)

// InitPaymentCheck initializes the payment check configuration.
func InitPaymentCheck(k Keeper, enabled bool, interval uint32) {
	k.cfg.Enabled = enabled
	k.cfg.Interval = interval
}

// RunPaymentCheck checks the payment data of all buckets and objects.
// It will compare the lock balance, net flow rate of users and gvg families/gvgs/validator tax pool.
func (k Keeper) RunPaymentCheck(ctx sdk.Context) error {
	ctx.Logger().Info("start checking payment data")

	type Detail struct {
		address string
		amount  sdkmath.Int
	}

	lockBalanceMap := make(map[string]sdkmath.Int)         // payment address -> lock balance
	lockBalanceDetailMap := make(map[string][]Detail)      // payment address -> {bucket name, lock balance}
	userFlowRateMap := make(map[string]sdkmath.Int)        // payment address -> net flow rate
	userFlowRateDetailMap := make(map[string][]Detail)     // payment address -> {bucket name, net flow rate}
	receiverFlowRateMap := make(map[string]sdkmath.Int)    // gvg family/gvg/validator tax pool address -> net flow rate
	receiverFlowRateDetailMap := make(map[string][]Detail) // gvg family/gvg/validator tax pool address -> {bucket name, net flow rate}

	store := ctx.KVStore(k.storeKey)
	bucketStore := prefix.NewStore(store, types.BucketByIDPrefix)
	bucketIt := bucketStore.Iterator(nil, nil)
	defer bucketIt.Close()

	// cache all stream records to speed up
	streamRecordMap := make(map[string]paymenttypes.StreamRecord)
	allStreamRecords := k.paymentKeeper.GetAllStreamRecord(ctx)
	for _, record := range allStreamRecords {
		streamRecordMap[record.Account] = record
	}

	var result error

Exit:
	for ; bucketIt.Valid(); bucketIt.Next() {
		var bucket types.BucketInfo
		k.cdc.MustUnmarshal(bucketIt.Value(), &bucket)
		if bucket.Id.IsZero() {
			continue
		}

		// get net flow rate
		internalBucketInfo, found := k.GetInternalBucketInfo(ctx, bucket.Id)
		if !found {
			result = errors.New("internal bucket info not found")
			ctx.Logger().Error("internal bucket info not found", "bucket", bucket.BucketName)
			continue Exit
		}
		userFlows, err := k.GetBucketReadStoreBill(ctx, &bucket, internalBucketInfo)
		if err != nil {
			result = errors.New("fail to get bucket read and store bill")
			ctx.Logger().Error("fail to get bucket read and store bill", "bucket", bucket.BucketName, "error", err)
			continue Exit
		}

		rateLimited := false
		paymentAddress := bucket.PaymentAddress
		rateLimitStatus, found := k.getBucketFlowRateLimitStatus(ctx, bucket.BucketName)
		if found {
			rateLimited = rateLimitStatus.IsBucketLimited
			paymentAddress = rateLimitStatus.PaymentAddress
		}

		if len(userFlows.Flows) > 0 && !rateLimited { // if rate limited, there should not payment bills
			expectedNetFlowRate := sdkmath.ZeroInt()
			for _, flow := range userFlows.Flows {
				expectedNetFlowRate = expectedNetFlowRate.Add(flow.Rate)
				// gvg family/gvg/validator tax pool
				_, ok := receiverFlowRateMap[flow.ToAddress]
				if !ok {
					receiverFlowRateMap[flow.ToAddress] = sdkmath.ZeroInt()
					receiverFlowRateDetailMap[flow.ToAddress] = []Detail{}
				}
				receiverFlowRateMap[flow.ToAddress] = receiverFlowRateMap[flow.ToAddress].Add(flow.Rate)
				receiverFlowRateDetailMap[flow.ToAddress] = append(receiverFlowRateDetailMap[flow.ToAddress],
					Detail{bucket.BucketName, flow.Rate})
			}

			// user payment account
			expectedNetFlowRate = expectedNetFlowRate.Neg()
			_, ok := userFlowRateMap[paymentAddress]
			if !ok {
				userFlowRateMap[paymentAddress] = sdkmath.ZeroInt()
				userFlowRateDetailMap[paymentAddress] = []Detail{}
			}
			userFlowRateMap[paymentAddress] = userFlowRateMap[paymentAddress].Add(expectedNetFlowRate)
			userFlowRateDetailMap[paymentAddress] = append(userFlowRateDetailMap[paymentAddress],
				Detail{bucket.BucketName, expectedNetFlowRate})
		}

		// get lock balance
		objectPrefixStore := prefix.NewStore(store, types.GetObjectKeyOnlyBucketPrefix(bucket.BucketName))
		it := objectPrefixStore.Iterator(nil, nil)
		defer it.Close()

		expectedLockBalance := sdkmath.ZeroInt()
		for ; it.Valid(); it.Next() {
			u256Seq := sequence.Sequence[sdkmath.Uint]{}
			objectInfo, found := k.GetObjectInfoById(ctx, u256Seq.DecodeSequence(it.Value()))
			if found && (objectInfo.ObjectStatus == types.OBJECT_STATUS_CREATED || objectInfo.IsUpdating) {
				priceTime := objectInfo.GetLatestUpdatedTime()
				payloadSize := objectInfo.PayloadSize
				if objectInfo.IsUpdating {
					shadowObject, found := k.GetShadowObjectInfo(ctx, bucket.BucketName, objectInfo.ObjectName)
					if !found {
						result = errors.New("shadow object not found")
						ctx.Logger().Error("shadow object not found", "bucket", bucket.BucketName, "object", objectInfo.ObjectName)
						continue Exit
					}
					priceTime = shadowObject.UpdatedAt
					payloadSize = shadowObject.PayloadSize
				}

				lockAmount, _, err := k.GetObjectLockFee(ctx, priceTime, payloadSize)
				if err != nil {
					result = errors.New("get object lock fee failed")
					ctx.Logger().Error("get object lock fee failed", "bucket", bucket.BucketName, "object", objectInfo.ObjectName, "error", err)
					continue Exit
				}
				expectedLockBalance = expectedLockBalance.Add(lockAmount)
			}
		}

		if expectedLockBalance.IsPositive() {
			_, ok := lockBalanceMap[bucket.PaymentAddress]
			if !ok {
				lockBalanceMap[bucket.PaymentAddress] = sdkmath.ZeroInt()
				lockBalanceDetailMap[bucket.PaymentAddress] = []Detail{}
			}
			lockBalanceMap[bucket.PaymentAddress] = lockBalanceMap[bucket.PaymentAddress].Add(expectedLockBalance)
			lockBalanceDetailMap[bucket.PaymentAddress] = append(lockBalanceDetailMap[bucket.PaymentAddress],
				Detail{bucket.BucketName, expectedLockBalance})
		}
	}

	if result != nil { // if already has error, do not check the following
		ctx.Logger().Info("stop checking payment data due to error")
		return result
	}

	// compare lock balance: expected -> actual side
	for address, expectedLockBalance := range lockBalanceMap {
		streamRecord, found := streamRecordMap[address]
		if !found {
			result = errors.New("comparing lock balance - stream record not found")
			ctx.Logger().Error("comparing lock balance - stream record not found", "address", address)
			continue // to print all errors if there are any
		}

		actualLockBalance := streamRecord.LockBalance
		if !expectedLockBalance.Equal(actualLockBalance) {
			if !k.isKnownLockBalanceIssue(ctx, address) {
				result = errors.New("lock balance not equal")
			}
			ctx.Logger().Error("lock balance not equal", "address", address, "expected", expectedLockBalance, "actual", actualLockBalance)
			details := lockBalanceDetailMap[address]
			for _, detail := range details {
				ctx.Logger().Error("lock balance detail", "bucket", detail.address, "amount", detail.amount)
			}
		}
	}

	// compare user net flow rate: expected -> actual side
	frozenReceiverFlowRateMap := make(map[string]sdkmath.Int)
	for address, expectedNetFlowRate := range userFlowRateMap {
		streamRecord, found := streamRecordMap[address]
		if !found {
			result = errors.New("comparing user net flow rate - stream record not found")
			ctx.Logger().Error("comparing user net flow rate - stream record not found", "address", address)
			continue // to print all errors if there are any
		}

		actualNetFlowRate := streamRecord.NetflowRate
		if streamRecord.Status == paymenttypes.STREAM_ACCOUNT_STATUS_FROZEN {
			actualNetFlowRate = actualNetFlowRate.Add(streamRecord.FrozenNetflowRate)

			// be noted, payment outflows can be in different status even if the stream account is frozen,
			// for we are force settling a stream account in multiple blocks
			outFlows := k.paymentKeeper.GetOutFlows(ctx, sdk.MustAccAddressFromHex(address))
			for _, outFlow := range outFlows {
				if outFlow.Status == paymenttypes.OUT_FLOW_STATUS_FROZEN {
					_, ok := frozenReceiverFlowRateMap[outFlow.ToAddress]
					if !ok {
						frozenReceiverFlowRateMap[outFlow.ToAddress] = sdkmath.ZeroInt()
					}
					frozenReceiverFlowRateMap[outFlow.ToAddress] = frozenReceiverFlowRateMap[outFlow.ToAddress].Add(outFlow.Rate)
				}
			}
		}

		if actualNetFlowRate.IsNegative() && streamRecord.OutFlowCount <= 0 {
			result = errors.New("user net flow rate invalid status or out flow count")
			ctx.Logger().Error("user net flow rate invalid flow rate or out flow count",
				"address", address, "status", actualNetFlowRate, "outflow count", streamRecord.OutFlowCount)
		}

		if !expectedNetFlowRate.Equal(actualNetFlowRate) {
			result = errors.New("user net flow rate not equal")
			ctx.Logger().Error("user net flow rate not equal", "address", address, "expected", expectedNetFlowRate, "actual", actualNetFlowRate)
			details := userFlowRateDetailMap[address]
			for _, detail := range details {
				ctx.Logger().Error("user net flow rate detail", "bucket", detail.address, "amount", detail.amount)
			}
		}
	}

	// compare receiver net flow rate: expected -> actual side
	for address, expectedNetFlowRate := range receiverFlowRateMap {
		streamRecord, found := streamRecordMap[address]
		if !found {
			result = errors.New("comparing receiver net flow rate - stream record not found")
			ctx.Logger().Error("comparing receiver net flow rate - stream record not found", "address", address)
			continue // to print all errors if there are any
		}

		if streamRecord.Status == paymenttypes.STREAM_ACCOUNT_STATUS_FROZEN || streamRecord.OutFlowCount > 0 {
			result = errors.New("receiver net flow rate invalid status or out flow count")
			ctx.Logger().Error("receiver net flow rate invalid status or out flow count",
				"address", address, "status", streamRecord.Status, "outflow count", streamRecord.OutFlowCount)
		}

		actualNetFlowRate := streamRecord.NetflowRate
		frozenRate, ok := frozenReceiverFlowRateMap[address]
		if ok {
			actualNetFlowRate = actualNetFlowRate.Add(frozenRate)
		}

		if !expectedNetFlowRate.Equal(actualNetFlowRate) {
			result = errors.New("receiver net flow rate not equal")
			ctx.Logger().Error("receiver net flow rate not equal", "address", address, "expected", expectedNetFlowRate, "actual", actualNetFlowRate)
			details := receiverFlowRateDetailMap[address]
			for _, detail := range details {
				ctx.Logger().Error("receiver net flow rate detail", "bucket", detail.address, "amount", detail.amount)
			}
		}
	}

	// compare lock balance: actual -> expected side
	// compare user net flow rate: actual -> expected side
	// compare receiver net flow rate: actual -> expected side
	for _, streamRecord := range streamRecordMap {
		if streamRecord.LockBalance.IsPositive() {
			_, found := lockBalanceMap[streamRecord.Account]
			if !found {
				if !k.isKnownLockBalanceIssue(ctx, streamRecord.Account) {
					result = errors.New("the stream record has lock balance which is not expected")
				}
				ctx.Logger().Error("the stream record has lock balance which is not expected", "address", streamRecord.Account)
			}
		}

		if streamRecord.NetflowRate.IsNegative() || streamRecord.FrozenNetflowRate.IsNegative() {
			_, found := userFlowRateMap[streamRecord.Account]
			if !found {
				result = errors.New("the stream record has negative flow rate which is not expected")
				ctx.Logger().Error("the stream record has negative flow rate which is not expected", "address", streamRecord.Account)
			}
		}

		if streamRecord.NetflowRate.IsPositive() {
			_, found := receiverFlowRateMap[streamRecord.Account]
			if !found {
				result = errors.New("the stream record has positive flow rate which is not expected")
				ctx.Logger().Error("the stream record has positive flow rate which is not expected", "address", streamRecord.Account)
			}
		}
	}

	ctx.Logger().Info("finish checking payment data")
	return result
}

// isKnownLockBalanceIssue checks if the address is the known addresses of the lock balance issue on testnet.
func (k Keeper) isKnownLockBalanceIssue(ctx sdk.Context, address string) bool {
	if ctx.ChainID() != upgradetypes.TestnetChainID {
		return false
	}
	if address == "0x8E15D16d6432166372Fb1e6f4A41840D71edd41F" || address == "0x9b825492966508C587536bA71425d61E822545C3" {
		return true
	}
	return false
}
