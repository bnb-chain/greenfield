package keeper

import (
	"context"
	"fmt"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MockSealObject(goCtx context.Context, msg *types.MsgMockSealObject) (*types.MsgMockSealObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	bucketMeta, found := k.GetMockBucketMeta(ctx, msg.BucketName)
	if !found {
		return nil, fmt.Errorf("bucket not found")
	}
	objectInfo, found := k.GetMockObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	if !found {
		return nil, fmt.Errorf("object not found")
	}
	if objectInfo.ObjectState != types.OBJECT_STATE_INIT {
		return nil, fmt.Errorf("object state is not init")
	}
	if objectInfo.Owner != msg.Operator {
		return nil, fmt.Errorf("object owner is not the same as msg owner")
	}
	objectInfo.ObjectState = types.OBJECT_STATE_IN_SERVICE
	for _, secondarySP := range msg.SecondarySPs {
		objectInfo.SecondarySPs = append(objectInfo.SecondarySPs, &types.StorageProviderInfo{
			Id: secondarySP,
		})
	}
	err := k.UnlockAndChargeStoreFee(ctx, &bucketMeta, &objectInfo)
	if err != nil {
		return nil, fmt.Errorf("unlock and charge store fee failed: %w", err)
	}
	objectInfo.LockedBalance = nil
	k.SetMockObjectInfo(ctx, objectInfo)
	k.SetMockBucketMeta(ctx, bucketMeta)
	return &types.MsgMockSealObjectResponse{}, nil
}
