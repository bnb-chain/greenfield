package keeper

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k Keeper) DynamicBalance(goCtx context.Context, req *types.QueryDynamicBalanceRequest) (*types.QueryDynamicBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	account, err := sdk.AccAddressFromHexUnsafe(req.Account)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid account")
	}
	streamRecord, _ := k.GetStreamRecord(
		ctx,
		account,
	)
	currentTimestamp := ctx.BlockTime().Unix()
	flowDelta := streamRecord.NetflowRate.MulRaw(currentTimestamp - streamRecord.CrudTimestamp)
	dynamicBalance := streamRecord.StaticBalance.Add(flowDelta)
	// get bank balance
	params := k.GetParams(ctx)
	bankBalance := math.ZeroInt()
	if k.accountKeeper.HasAccount(ctx, account) {
		bankBalance = k.bankKeeper.GetBalance(ctx, account, params.FeeDenom).Amount
	}
	// calc frontend fields
	lockedFee := streamRecord.LockBalance.Add(streamRecord.BufferBalance)
	availableBalance := bankBalance
	if streamRecord.StaticBalance.IsPositive() {
		availableBalance = availableBalance.Add(streamRecord.StaticBalance)
	} else {
		lockedFee = lockedFee.Add(streamRecord.StaticBalance)
	}
	return &types.QueryDynamicBalanceResponse{
		DynamicBalance:   dynamicBalance,
		StreamRecord:     *streamRecord,
		CurrentTimestamp: currentTimestamp,
		BankBalance:      bankBalance,
		AvailableBalance: availableBalance,
		LockedFee:        lockedFee,
		ChangeRate:       streamRecord.NetflowRate,
	}, nil
}
