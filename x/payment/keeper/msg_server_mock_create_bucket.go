package keeper

import (
	"context"
	"fmt"

	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

func (k msgServer) MockCreateBucket(goCtx context.Context, msg *types.MsgMockCreateBucket) (*types.MsgMockCreateBucketResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// TODO: msg verification
	bucketMeta, found := k.GetMockBucketMeta(ctx, msg.BucketName)
	if found {
		return nil, fmt.Errorf("bucket already exists")
	}
	readPacket := types.ReadPacket(msg.ReadPacket)
	// compose bucket meta
	bucketMeta = types.MockBucketMeta{
		BucketName: msg.BucketName,
		Owner:      msg.Operator,
		SpAddress:  msg.SpAddress,
		ReadPacket: readPacket,
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
	if readPacket != types.ReadPacketFree {
		err := k.ChargeInitialReadFee(ctx, &bucketMeta)
		if err != nil {
			return nil, fmt.Errorf("charge initial read fee failed: %w", err)
		}
	}
	// save bucket meta
	k.SetMockBucketMeta(ctx, bucketMeta)
	return &types.MsgMockCreateBucketResponse{}, nil
}
