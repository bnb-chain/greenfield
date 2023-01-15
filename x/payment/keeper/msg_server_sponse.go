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
	change := types.NewDefaultStreamRecordChangeWithAddr(msg.Creator).WithRateChange(msg.Rate.Neg()).WithAutoTransfer(true)
	err := k.Keeper.UpdateStreamRecord(ctx, &fromStream, &change)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "update stream record by rate failed, creator: %s", msg.Creator)
	}
	change = types.NewDefaultStreamRecordChangeWithAddr(msg.To).WithRateChange(msg.Rate).WithAutoTransfer(false)
	err = k.Keeper.UpdateStreamRecord(ctx, &toStream, &change)
	if err != nil {
		return nil, errorsmod.Wrapf(err, "update stream record by rate failed, to: %s", msg.To)
	}
	return &types.MsgSponseResponse{}, nil
}
