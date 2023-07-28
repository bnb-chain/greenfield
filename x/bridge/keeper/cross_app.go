package keeper

import (
	"encoding/hex"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"

	"github.com/bnb-chain/greenfield/x/bridge/types"
)

func RegisterCrossApps(keeper Keeper) {
	transferOutApp := NewTransferOutApp(keeper)
	err := keeper.crossChainKeeper.RegisterChannel(types.TransferOutChannel, types.TransferOutChannelID, transferOutApp)
	if err != nil {
		panic(err)
	}

	transferInApp := NewTransferInApp(keeper)
	err = keeper.crossChainKeeper.RegisterChannel(types.TransferInChannel, types.TransferInChannelID, transferInApp)
	if err != nil {
		panic(err)
	}
}

var _ sdk.CrossChainApplication = &TransferOutApp{}

type TransferOutApp struct {
	bridgeKeeper Keeper
}

func NewTransferOutApp(keeper Keeper) *TransferOutApp {
	return &TransferOutApp{
		bridgeKeeper: keeper,
	}
}

func (app *TransferOutApp) ExecuteAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		return sdk.ExecuteResult{}
	}

	app.bridgeKeeper.Logger(ctx).Info("receive transfer out refund ack package")

	refundPackage, err := types.DeserializeTransferOutRefundPackage(payload)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("decode transfer out refund claim error", "err", err.Error(), "claim", hex.EncodeToString(payload))
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	denom := app.bridgeKeeper.stakingKeeper.BondDenom(ctx) // only support native token so far
	err = app.bridgeKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, crosschaintypes.ModuleName, refundPackage.RefundAddress,
		sdk.Coins{
			sdk.Coin{
				Denom:  denom,
				Amount: sdk.NewIntFromBigInt(refundPackage.RefundAmount),
			},
		},
	)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("send coins error", "err", err.Error())
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventCrossTransferOutRefund{
		RefundAddress: refundPackage.RefundAddress.String(),
		Amount: &sdk.Coin{
			Denom:  denom,
			Amount: sdk.NewIntFromBigInt(refundPackage.RefundAmount),
		},
		RefundReason: types.RefundReason(refundPackage.RefundReason),
		Sequence:     appCtx.Sequence,
		DestChainId:  uint32(appCtx.SrcChainId),
	})
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("emit event error", "err", err.Error())
		panic(err)
	}

	return sdk.ExecuteResult{}
}

func (app *TransferOutApp) ExecuteFailAckPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Info("received transfer out fail ack package")

	transferOutPackage, err := types.DeserializeTransferOutSynPackage(payload)
	if err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	denom := app.bridgeKeeper.stakingKeeper.BondDenom(ctx) // only support native token so far
	err = app.bridgeKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, crosschaintypes.ModuleName, transferOutPackage.RefundAddress,
		sdk.Coins{
			sdk.Coin{
				Denom:  denom,
				Amount: sdk.NewIntFromBigInt(transferOutPackage.Amount),
			},
		},
	)

	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("send coins error", "err", err.Error())
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	err = ctx.EventManager().EmitTypedEvent(&types.EventCrossTransferOutRefund{
		RefundAddress: transferOutPackage.RefundAddress.String(),
		Amount: &sdk.Coin{
			Denom:  denom,
			Amount: sdk.NewIntFromBigInt(transferOutPackage.Amount),
		},
		RefundReason: types.REFUND_REASON_FAIL_ACK,
		Sequence:     appCtx.Sequence,
		DestChainId:  uint32(appCtx.SrcChainId),
	})
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("emit event error", "err", err.Error())
	}

	return sdk.ExecuteResult{}
}

func (app *TransferOutApp) ExecuteSynPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Error("received transfer out syn package ")
	return sdk.ExecuteResult{}
}

var _ sdk.CrossChainApplication = &TransferInApp{}

type TransferInApp struct {
	bridgeKeeper Keeper
}

func NewTransferInApp(bridgeKeeper Keeper) *TransferInApp {
	return &TransferInApp{
		bridgeKeeper: bridgeKeeper,
	}
}

func (app *TransferInApp) ExecuteAckPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Error("received transfer in ack package", "payload", hex.EncodeToString(payload))
	return sdk.ExecuteResult{}
}

func (app *TransferInApp) ExecuteFailAckPackage(ctx sdk.Context, header *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Error("received transfer in fail ack package", "payload", hex.EncodeToString(payload))
	return sdk.ExecuteResult{}
}

func (app *TransferInApp) ExecuteSynPackage(ctx sdk.Context, appCtx *sdk.CrossChainAppContext, payload []byte) sdk.ExecuteResult {
	transferInPackage, err := types.DeserializeTransferInSynPackage(payload)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("decode transfer in claim error", "err", err.Error(), "claim", string(payload))
		panic("decode transfer in claim error")
	}

	denom := app.bridgeKeeper.stakingKeeper.BondDenom(ctx)
	amount := sdk.NewCoin(denom, sdk.NewIntFromBigInt(transferInPackage.Amount))

	err = app.bridgeKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, crosschaintypes.ModuleName, transferInPackage.ReceiverAddress, sdk.Coins{amount})
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("send coins error", "err", err.Error())
		refundPackage, refundErr := app.bridgeKeeper.GetRefundTransferInPayload(transferInPackage, uint32(types.REFUND_REASON_INSUFFICIENT_BALANCE))
		if refundErr != nil {
			app.bridgeKeeper.Logger(ctx).Error("get refund transfer in payload error", "err", refundErr.Error())
			panic(refundErr)
		}
		return sdk.ExecuteResult{
			Payload: refundPackage,
			Err:     errors.Wrapf(types.ErrInvalidPackage, "send coins error: %s", err.Error()),
		}
	}

	// emit event
	err = ctx.EventManager().EmitTypedEvent(&types.EventCrossTransferIn{
		Amount:          &amount,
		ReceiverAddress: transferInPackage.ReceiverAddress.String(),
		RefundAddress:   transferInPackage.RefundAddress.String(),
		Sequence:        appCtx.Sequence,
		SrcChainId:      uint32(appCtx.SrcChainId),
	})
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("emit event error", "err", err.Error())
		panic(err)
	}

	return sdk.ExecuteResult{}
}
