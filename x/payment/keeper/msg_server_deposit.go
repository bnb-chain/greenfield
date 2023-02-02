package keeper

import (
	"context"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// bank transfer
	creator, _ := sdk.AccAddressFromHexUnsafe(msg.Creator)
	coins := sdk.NewCoins(sdk.NewCoin(k.GetParams(ctx).FeeDenom, msg.Amount))
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, creator, types.ModuleName, coins)
	if err != nil {
		return nil, err
	}
	// change payment record
	streamRecord, found := k.Keeper.GetStreamRecord(ctx, msg.To)
	if !found {
		streamRecord.Account = msg.To
		streamRecord.CrudTimestamp = ctx.BlockTime().Unix()
		streamRecord.StaticBalance = msg.Amount
		k.Keeper.SetStreamRecord(ctx, streamRecord)
		return &types.MsgDepositResponse{}, nil
	} else {
		// TODO:
		// 1. check if the stream should be forced settled
		// 2. if the account is frozen, assume it
		change := types.NewDefaultStreamRecordChangeWithAddr(msg.To).WithStaticBalanceChange(msg.Amount)
		err := k.UpdateStreamRecord(ctx, &streamRecord, change)
		return &types.MsgDepositResponse{}, err
	}
}
