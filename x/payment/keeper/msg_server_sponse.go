package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	"fmt"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Sponse(goCtx context.Context, msg *types.MsgSponse) (*types.MsgSponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fromStream, found := k.Keeper.GetStreamRecord(ctx, msg.Creator)
	if !found {
		return nil, fmt.Errorf("creator stream record not found")
	}
	if fromStream.Status != types.StreamPaymentAccountStatusNormal {
		return nil, fmt.Errorf("creator stream record status is not normal")
	}
	toStream, found := k.Keeper.GetStreamRecord(ctx, msg.To)
	if !found {
		return nil, fmt.Errorf("to stream record not found")
	}
	if toStream.Status != types.StreamPaymentAccountStatusNormal {
		return nil, fmt.Errorf("to stream record status is not normal")
	}
	err := k.Keeper.UpdateStreamRecord(ctx, &fromStream, msg.Rate.Neg(), sdkmath.ZeroInt(), true)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "update stream record by rate failed, creator: %s", msg.Creator)
	}
	err = k.Keeper.UpdateStreamRecord(ctx, &toStream, msg.Rate, sdkmath.ZeroInt(), false)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "update stream record by rate failed, to: %s", msg.To)
	}
	return &types.MsgSponseResponse{}, nil
}
