package keeper

import (
	"context"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) StorageProviderExit(goCtx context.Context, msg *types.MsgStorageProviderExit) (*types.MsgStorageProviderExitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.OperatorAddress)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	sp.Status = sptypes.STATUS_GRACEFUL_EXITING

	k.spKeeper.SetStorageProvider(ctx, sp)

	return &types.MsgStorageProviderExitResponse{}, nil
}
