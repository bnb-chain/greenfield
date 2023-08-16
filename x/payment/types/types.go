package types

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
)

var (
	// GovernanceAddress used to receive fee of storage system, and pay for the potential debt from late forced settlement
	GovernanceAddress       = sdk.AccAddress(address.Module(ModuleName, []byte("governance"))[:sdk.EthAddressLength])
	ValidatorTaxPoolAddress = sdk.AccAddress(address.Module(ModuleName, []byte("validator-tax-pool"))[:sdk.EthAddressLength])
)

const (
	ForceUpdateStreamRecordKey = "force_update_stream_record"
)

const (
	// GovernanceAddressLackBalanceLabel is the metrics label to notify that the governance account has no enough balance
	GovernanceAddressLackBalanceLabel = "governance_address_lack_balance"
)

type StreamRecordChange struct {
	Addr                sdk.AccAddress
	RateChange          sdkmath.Int
	StaticBalanceChange sdkmath.Int
	LockBalanceChange   sdkmath.Int
	FrozenRateChange    sdkmath.Int
}

func NewDefaultStreamRecordChangeWithAddr(addr sdk.AccAddress) *StreamRecordChange {
	return &StreamRecordChange{
		Addr:                addr,
		RateChange:          sdkmath.ZeroInt(),
		StaticBalanceChange: sdkmath.ZeroInt(),
		LockBalanceChange:   sdkmath.ZeroInt(),
		FrozenRateChange:    sdkmath.ZeroInt(),
	}
}

func (change *StreamRecordChange) WithRateChange(rateChange sdkmath.Int) *StreamRecordChange {
	change.RateChange = rateChange
	return change
}

func (change *StreamRecordChange) WithStaticBalanceChange(staticBalanceChange sdkmath.Int) *StreamRecordChange {
	change.StaticBalanceChange = staticBalanceChange
	return change
}

func (change *StreamRecordChange) WithLockBalanceChange(lockBalanceChange sdkmath.Int) *StreamRecordChange {
	change.LockBalanceChange = lockBalanceChange
	return change
}

func (change *StreamRecordChange) WithFrozenRateChange(frozenRateChange sdkmath.Int) *StreamRecordChange {
	change.FrozenRateChange = frozenRateChange
	return change
}

type UserFlows struct {
	From  sdk.AccAddress
	Flows []OutFlow
}
