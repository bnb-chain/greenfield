package v2

import (
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	v1 "github.com/bnb-chain/greenfield/x/virtualgroup/keeper/v1"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func MigrateStore(ctx sdk.Context, storeKey storetypes.StoreKey, cdc codec.BinaryCodec) error {
	store := ctx.KVStore(storeKey)
	bz := store.Get(types.ParamsKey)
	oldParams := &v1.Params{}
	cdc.MustUnmarshal(bz, oldParams)

	newParams := types.NewParams(
		oldParams.DepositDenom,
		oldParams.GvgStakingPerBytes,
		oldParams.MaxGlobalVirtualGroupNumPerFamily,
		oldParams.MaxStoreSizePerFamily,
		types.DefaultSwapInValidityPeriod,
		types.DefaultSPConcurrentExitNum)
	store.Set(types.ParamsKey, cdc.MustMarshal(&newParams))

	return nil
}
