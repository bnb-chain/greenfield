package keeper

import (
	"context"

	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) CompleteStorageProviderExit(goCtx context.Context, msg *types.MsgCompleteStorageProviderExit) (*types.MsgCompleteStorageProviderExitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddress := sdk.MustAccAddressFromHex(msg.OperatorAddress)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddress)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	if sp.Status == sptypes.STATUS_GRACEFUL_EXITING {
		return nil, sptypes.ErrStorageProviderExitFailed.Wrapf(
			"sp(id : %d, operator address: %s) not in the process of exiting", sp.Id, sp.OperatorAddress)
	}

	err := k.IsStorageProviderCanExit(ctx, msg.OriginStorageProviderId)
	if err != nil {
		return nil, err
	}

	err = k.spKeeper.Exit(ctx, sp)
	if err != nil {
		return nil, err
	}

	return &types.MsgCompleteStorageProviderExitResponse{}, nil
}
