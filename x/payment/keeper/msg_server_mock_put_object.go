package keeper

import (
	"context"
	"fmt"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MockPutObject(goCtx context.Context, msg *types.MsgMockPutObject) (*types.MsgMockPutObjectResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	bucketMeta, found := k.GetMockBucketMeta(ctx, msg.BucketName)
	if bucketMeta.Owner != msg.Owner {
		return nil, fmt.Errorf("bucket owner is not the same as msg owner")
	}
	objectInfo, found := k.GetMockObjectInfo(ctx, msg.BucketName, msg.ObjectName)
	if found {
		return nil, fmt.Errorf("object already exists")
	}
	objectInfo = types.MockObjectInfo{
		BucketName:  msg.BucketName,
		ObjectName:  msg.ObjectName,
		Owner:       msg.Owner,
		Id:          msg.BucketName + "/" + msg.ObjectName,
		Size_:       msg.Size_,
		CreateAt:    uint64(ctx.BlockTime().Unix()),
		ObjectState: types.OBJECT_STATE_INIT,
	}
	// lock store fee
	err := k.LockStoreFee(ctx, &bucketMeta, &objectInfo)
	if err != nil {
		return nil, fmt.Errorf("lock store fee failed: %w", err)
	}
	k.SetMockObjectInfo(ctx, objectInfo)
	return &types.MsgMockPutObjectResponse{}, nil
}
