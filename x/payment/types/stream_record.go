package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func NewStreamRecord(account sdk.AccAddress, crudTimestamp int64) *StreamRecord {
	return &StreamRecord{
		Account:       account.String(),
		CrudTimestamp: crudTimestamp,
		StaticBalance: sdkmath.ZeroInt(),
		BufferBalance: sdkmath.ZeroInt(),
		NetflowRate:   sdkmath.ZeroInt(),
		LockBalance:   sdkmath.ZeroInt(),
		Status:        STREAM_ACCOUNT_STATUS_ACTIVE,
	}
}
