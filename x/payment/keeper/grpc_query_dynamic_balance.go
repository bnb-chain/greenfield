package keeper

import (
	"context"

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
	streamRecord, found := k.GetStreamRecord(
		ctx,
		account,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "payment account not found")
	}
	currentTimestamp := ctx.BlockTime().Unix()
	flowDelta := streamRecord.NetflowRate.MulRaw(currentTimestamp - streamRecord.CrudTimestamp)
	dynamicBalance := streamRecord.StaticBalance.Add(flowDelta)
	return &types.QueryDynamicBalanceResponse{
		DynamicBalance:   dynamicBalance,
		StreamRecord:     *streamRecord,
		CurrentTimestamp: currentTimestamp,
	}, nil
}
