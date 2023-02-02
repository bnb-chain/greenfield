package keeper

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MockUpdateBucketReadPacket(goCtx context.Context, msg *types.MsgMockUpdateBucketReadPacket) (*types.MsgMockUpdateBucketReadPacketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	bucketMeta, found := k.GetMockBucketMeta(ctx, msg.BucketName)
	if !found {
		return nil, fmt.Errorf("bucket not exists")
	}
	newReadPacket, err := types.ParseReadPacket(msg.ReadPacket)
	if err != nil {
		return nil, fmt.Errorf("parse read packet failed: %w", err)
	}
	if newReadPacket == bucketMeta.ReadPacket {
		return nil, fmt.Errorf("read packet is not changed")
	}
	if bucketMeta.Owner != msg.Operator {
		return nil, fmt.Errorf("not bucket owner")
	}
	// charge read packet fee if it's changed
	err = k.ChargeUpdateReadPacket(ctx, &bucketMeta, newReadPacket)
	if err != nil {
		return nil, fmt.Errorf("charge update read packet failed: %w", err)
	}
	// change bucket meta
	bucketMeta.PriceTime = ctx.BlockTime().Unix()
	bucketMeta.ReadPacket = newReadPacket
	k.SetMockBucketMeta(ctx, bucketMeta)
	return &types.MsgMockUpdateBucketReadPacketResponse{}, nil
}
