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

	_, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddress)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	err := k.IsStorageProviderCanExit(ctx, msg.OriginStorageProviderId)
	if err != nil {
		return nil, err
	}
	return &types.MsgCompleteStorageProviderExitResponse{}, nil
}
