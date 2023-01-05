package keeper

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (k Keeper) DynamicBalance(goCtx context.Context, req *types.QueryDynamicBalanceRequest) (*types.QueryDynamicBalanceResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "invalid request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	streamRecord, found := k.GetStreamRecord(
		ctx,
		req.Account,
	)
	if !found {
		return nil, status.Error(codes.NotFound, "not found")
	}
	currentTimestamp := ctx.BlockTime().Unix()
	flowDelta := streamRecord.NetflowRate.MulRaw(currentTimestamp - streamRecord.CrudTimestamp)
	dynamicBalance := streamRecord.StaticBalance.Add(flowDelta)
	return &types.QueryDynamicBalanceResponse{
		DynamicBalance:   dynamicBalance,
		StreamRecord:     streamRecord,
		CurrentTimestamp: currentTimestamp,
	}, nil
}
