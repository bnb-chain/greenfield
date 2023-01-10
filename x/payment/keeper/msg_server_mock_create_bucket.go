package keeper

import (
	"context"
	"fmt"

	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MockCreateBucket(goCtx context.Context, msg *types.MsgMockCreateBucket) (*types.MsgMockCreateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: msg verification
	// compose bucket meta
	bucketMeta := types.MockBucketMeta{
		BucketName: msg.BucketName,
		Owner:      msg.Operator,
		SpAddress:  msg.SpAddress,
		ReadPacket: msg.ReadPacket,
		PriceTime:  ctx.BlockTime().Unix(),
	}
	if msg.ReadPaymentAccount == "" || msg.ReadPaymentAccount == msg.Operator {
		bucketMeta.ReadPaymentAccount = msg.Operator
	} else {
		if !k.IsPaymentAccountOwner(ctx, msg.ReadPaymentAccount, msg.Operator) {
			return nil, types.ErrNotPaymentAccountOwner
		}
		bucketMeta.ReadPaymentAccount = msg.ReadPaymentAccount
	}
	if msg.StorePaymentAccount == "" || msg.StorePaymentAccount == msg.Operator {
		bucketMeta.StorePaymentAccount = msg.Operator
	} else {
		if !k.IsPaymentAccountOwner(ctx, msg.StorePaymentAccount, msg.Operator) {
			return nil, types.ErrNotPaymentAccountOwner
		}
		bucketMeta.StorePaymentAccount = msg.StorePaymentAccount
	}
	// charge read packet fee if it's not free level
	readPacket := types.ReadPacket(msg.ReadPacket)
	if readPacket != types.ReadPacketLevelFree {
		err := k.ChargeInitialReadFee(goCtx, bucketMeta.ReadPaymentAccount, msg.SpAddress, readPacket)
		if err != nil {
			return nil, fmt.Errorf("charge initial read fee failed: %w", err)
		}
	}
	// save bucket meta
	k.SetMockBucketMeta(ctx, bucketMeta)
	return &types.MsgMockCreateBucketResponse{}, nil
}
