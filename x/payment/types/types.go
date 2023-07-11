package types

import (
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
