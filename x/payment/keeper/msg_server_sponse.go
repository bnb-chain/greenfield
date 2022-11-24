package keeper

import (
	"context"
	errorsmod "cosmossdk.io/errors"
	"fmt"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Sponse(goCtx context.Context, msg *types.MsgSponse) (*types.MsgSponseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if msg.Rate <= 0 {
		return nil, fmt.Errorf("rate must be positive")
	}
	if msg.Creator == msg.To {
		return nil, fmt.Errorf("can not sponse to yourself")
	}
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
	err := k.Keeper.UpdateStreamRecordByRate(ctx, &fromStream, -msg.Rate)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "update stream record by rate failed, creator: %s", msg.Creator)
	}
	err = k.Keeper.UpdateStreamRecordByRate(ctx, &toStream, msg.Rate)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "update stream record by rate failed, to: %s", msg.To)
	}
	k.Keeper.SetStreamRecord(ctx, fromStream)
	k.Keeper.SetStreamRecord(ctx, toStream)
	return &types.MsgSponseResponse{}, nil
}
