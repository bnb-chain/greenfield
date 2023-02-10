package types

import (
	sdkmath "cosmossdk.io/math"
)

const (
	StreamPaymentAccountStatusNormal = 0
	StreamPaymentAccountStatusFrozen = 1
)

func NewStreamRecord(account string, crudTimestamp int64) StreamRecord {
	return StreamRecord{
		Account:       account,
		CrudTimestamp: crudTimestamp,
		StaticBalance: sdkmath.ZeroInt(),
		BufferBalance: sdkmath.ZeroInt(),
		NetflowRate:   sdkmath.ZeroInt(),
		LockBalance:   sdkmath.ZeroInt(),
	}
}
