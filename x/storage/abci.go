package storage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	k "github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func BeginBlocker(ctx sdk.Context, keeper k.Keeper) {
	blockHeight := uint64(ctx.BlockHeight())
	discontinueWindow := keeper.DiscontinueRequestWindow(ctx)
	if blockHeight > 0 && blockHeight%discontinueWindow == 0 {
		keeper.ClearDiscontinueRequestCount(ctx)
	}
}

func EndBlocker(ctx sdk.Context, keeper k.Keeper) {
	blockHeight := uint64(ctx.BlockHeight())
	discontinueRequests := keeper.GetDiscontinueRequests(ctx, blockHeight)
	if len(discontinueRequests) == 0 {
		return
	}
	for operator, objectIds := range discontinueRequests {
		operatorAcc := sdk.AccAddress(operator)
		for _, objectId := range objectIds.ObjectId {
			err := keeper.ForceDeleteObject(ctx, operatorAcc, objectId, k.DeleteObjectOptions{SourceType: types.SOURCE_TYPE_ORIGIN})
			ctx.Logger().Error("failed to delete object", "id", objectId, "err", err.Error())
		}
	}
	keeper.ClearDiscontinueRequests(ctx, blockHeight)
}
