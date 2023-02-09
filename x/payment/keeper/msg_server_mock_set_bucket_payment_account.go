package keeper

import (
	"context"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

func (k msgServer) MockSetBucketPaymentAccount(goCtx context.Context, msg *types.MsgMockSetBucketPaymentAccount) (*types.MsgMockSetBucketPaymentAccountResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	bucketMeta, _ := k.GetMockBucketMeta(ctx, msg.BucketName)
	if bucketMeta.Owner != msg.Operator {
		return nil, fmt.Errorf("not bucket owner")
	}
	var readPaymentAccount *string
	var storePaymentAccount *string
	if msg.ReadPaymentAccount != "" && msg.ReadPaymentAccount != bucketMeta.ReadPaymentAccount {
		// change read payment account
		// check permission
		if !k.IsPaymentAccountOwner(ctx, msg.ReadPaymentAccount, msg.Operator) {
			return nil, fmt.Errorf("no permission to use read payment account")
		}
		readPaymentAccount = &msg.ReadPaymentAccount
	}
	if msg.StorePaymentAccount != "" && msg.StorePaymentAccount != bucketMeta.StorePaymentAccount {
		if !k.IsPaymentAccountOwner(ctx, msg.StorePaymentAccount, msg.Operator) {
			return nil, fmt.Errorf("no permission to use store payment account")
		}
		storePaymentAccount = &msg.StorePaymentAccount
	}
	if readPaymentAccount != nil || storePaymentAccount != nil {
		err := k.ChargeUpdatePaymentAccount(ctx, &bucketMeta, readPaymentAccount, storePaymentAccount)
		if err != nil {
			return nil, fmt.Errorf("charge update payment account failed: %w", err)
		}
		k.SetMockBucketMeta(ctx, bucketMeta)
	}
	return &types.MsgMockSetBucketPaymentAccountResponse{}, nil
}
