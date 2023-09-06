package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
	v1 "github.com/bnb-chain/greenfield/x/payment/types/v1"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.ParamsKey)

	oldParams := &v1.Params{}
	cdc.MustUnmarshal(bz, oldParams)

	newParams := types.NewParams(
		oldParams.VersionedParams.ReserveTime,
		oldParams.VersionedParams.ValidatorTaxRate,
		oldParams.ForcedSettleTime,
		oldParams.PaymentAccountCountLimit,
		oldParams.MaxAutoSettleFlowCount,
		oldParams.MaxAutoResumeFlowCount,
		oldParams.FeeDenom,
		types.DefaultWithdrawTimeLockThreshold,
		types.DefaultWithdrawTimeLockDuration)

	store.Set(types.ParamsKey, cdc.MustMarshal(&newParams))

	return nil
}
