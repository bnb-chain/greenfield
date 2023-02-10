package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) MockDeleteObject(goCtx context.Context, msg *types.MsgMockDeleteObject) (*types.MsgMockDeleteObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	objectInfo, found := k.GetMockObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, fmt.Errorf("object not found")
	}
	bucketMeta, _ := k.GetMockBucketMeta(ctx, msg.BucketName)
	if objectInfo.ObjectState == types.OBJECT_STATE_INIT {
		err := k.UnlockStoreFee(ctx, &bucketMeta, &objectInfo)
		if err != nil {
			return nil, fmt.Errorf("unlock store fee failed: %w", err)
		}
	} else {
		err := k.ChargeDeleteObject(ctx, &bucketMeta, &objectInfo)
		if err != nil {
			return nil, fmt.Errorf("charge delete fee failed: %w", err)
		}
	}
	k.RemoveMockObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	k.SetMockBucketMeta(ctx, bucketMeta)
	return &types.MsgMockDeleteObjectResponse{}, nil
}
