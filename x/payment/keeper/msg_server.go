package keeper

import (
	"context"
	"github.com/bnb-chain/greenfield/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of the MsgServer interface
// for the provided Keeper.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

var _ types.MsgServer = msgServer{}

func (k msgServer) UpdateSpStoragePrice(goCtx context.Context, msg *types.MsgUpdateSpStoragePrice) (*types.MsgUpdateSpStoragePriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	current := ctx.BlockTime().Unix()
	if current > msg.ExpireTime {
		return nil, types.ErrSpStoragePriceExpired
	}
	spStorePrice := types.SpStoragePrice{
		SpAddress:      msg.SpAddress,
		UpdateTime:     current,
		ReadQuotaPrice: msg.ReadQuotaPrice,
		StorePrice:     msg.StorePrice,
	}
	k.SetSpStoragePrice(ctx, spStorePrice)
	return &types.MsgUpdateSpStoragePriceResponse{}, nil
}
