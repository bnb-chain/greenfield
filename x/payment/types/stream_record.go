package types

import (
	sdkmath "cosmossdk.io/math"
)

func NewStreamRecord(account string, crudTimestamp int64) *StreamRecord {
	return &StreamRecord{
		Account:       account,
		CrudTimestamp: crudTimestamp,
		StaticBalance: sdkmath.ZeroInt(),
		BufferBalance: sdkmath.ZeroInt(),
		NetflowRate:   sdkmath.ZeroInt(),
		LockBalance:   sdkmath.ZeroInt(),
	}
}
