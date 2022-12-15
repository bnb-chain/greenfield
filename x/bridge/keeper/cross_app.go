package keeper

import (
	"encoding/hex"
	"math/big"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/bfs/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
)

var _ sdk.CrossChainApplication = &TransferOutApp{}

type TransferOutApp struct {
	bridgeKeeper Keeper
}

func (app *TransferOutApp) checkPackage(refundPackage *types.TransferOutRefundPackage) error {
	if refundPackage.RefundAddr.Empty() {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "refund address is empty")
	}

	if refundPackage.RefundAmount.Cmp(big.NewInt(0)) < 0 {
		return errors.Wrapf(types.ErrInvalidAmount, "amount to refund should not be negative")
	}
	return nil
}

func (app *TransferOutApp) ExecuteAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	if len(payload) == 0 {
		return sdk.ExecuteResult{}
	}

	app.bridgeKeeper.Logger(ctx).Info("receive transfer out refund ack package")

	refundPackage, err := types.DeserializeTransferOutRefundPackage(payload)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("unmarshal transfer out refund claim error", "err", err.Error(), "claim", hex.EncodeToString(payload))
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	err = app.checkPackage(refundPackage)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("check transfer out refund package error", "err", err.Error(), "claim", hex.EncodeToString(payload))
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	symbol := types.BytesToSymbol(refundPackage.TokenSymbol)
	err = app.bridgeKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, crosschaintypes.ModuleName, refundPackage.RefundAddr,
		sdk.Coins{
			sdk.Coin{
				Denom:  symbol,
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

	return sdk.ExecuteResult{}
}

func (app *TransferOutApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Info("received transfer out fail ack package")

	transferOutPackage, err := types.DeserializeTransferOutSynPackage(payload)
	if err != nil {
		return sdk.ExecuteResult{
			Err: err,
		}
	}

	symbol := types.BytesToSymbol(transferOutPackage.TokenSymbol)
	err = app.bridgeKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, crosschaintypes.ModuleName, transferOutPackage.RefundAddress,
		sdk.Coins{
			sdk.Coin{
				Denom:  symbol,
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

	return sdk.ExecuteResult{}
}

func (app *TransferOutApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, _ int64) sdk.ExecuteResult {
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

func (app *TransferInApp) checkTransferInSynPackage(transferInPackage *types.TransferInSynPackage) error {
	if len(transferInPackage.Amounts) == 0 {
		return errors.Wrapf(types.ErrInvalidLength, "length of Amounts should not be 0")
	}

	if len(transferInPackage.RefundAddresses) != len(transferInPackage.ReceiverAddresses) ||
		len(transferInPackage.RefundAddresses) != len(transferInPackage.Amounts) {
		return errors.Wrapf(types.ErrInvalidLength, "length of RefundAddresses, ReceiverAddresses, Amounts should be the same")
	}

	for _, addr := range transferInPackage.RefundAddresses {
		if addr.IsEmpty() {
			return errors.Wrapf(types.ErrInvalidAddress, "refund address should not be empty")

		}
	}

	for _, addr := range transferInPackage.ReceiverAddresses {
		if addr.Empty() {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "refund address is empty")
		}
	}

	for _, amount := range transferInPackage.Amounts {
		if amount.Cmp(big.NewInt(0)) < 0 {
			return errors.Wrapf(types.ErrInvalidAmount, "amount to refund should not be negative")
		}
	}

	return nil
}

func (app *TransferInApp) ExecuteAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Error("received transfer in ack package", "payload", hex.EncodeToString(payload))
	return sdk.ExecuteResult{}
}

func (app *TransferInApp) ExecuteFailAckPackage(ctx sdk.Context, payload []byte) sdk.ExecuteResult {
	app.bridgeKeeper.Logger(ctx).Error("received transfer in fail ack package", "payload", hex.EncodeToString(payload))
	return sdk.ExecuteResult{}
}

func (app *TransferInApp) ExecuteSynPackage(ctx sdk.Context, payload []byte, relayerFee int64) sdk.ExecuteResult {
	transferInPackage, err := types.DeserializeTransferInSynPackage(payload)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("unmarshal transfer in claim error", "err", err.Error(), "claim", string(payload))
		panic("unmarshal transfer in claim error")
	}

	err = app.checkTransferInSynPackage(transferInPackage)
	if err != nil {
		app.bridgeKeeper.Logger(ctx).Error("check transfer in package error", "err", err.Error(), "claim", string(payload))
		panic(err)
	}

	symbol := types.BytesToSymbol(transferInPackage.TokenSymbol)
	bondDenom := app.bridgeKeeper.stakingKeeper.BondDenom(ctx)

	// only support bond denom
	if symbol != bondDenom {
		refundPackage, err := app.bridgeKeeper.GetRefundTransferInPayload(transferInPackage, types.UnsupportedSymbol)
		if err != nil {
			app.bridgeKeeper.Logger(ctx).Error("get refund transfer in payload error", "err", err)
			panic(err)
		}
		return sdk.ExecuteResult{
			Payload: refundPackage,
			Err:     types.ErrUnsupportedDenom,
		}
	}

	if int64(transferInPackage.ExpireTime) < ctx.BlockHeader().Time.Unix() {
		refundPackage, sdkErr := app.bridgeKeeper.GetRefundTransferInPayload(transferInPackage, types.Timeout)
		if sdkErr != nil {
			app.bridgeKeeper.Logger(ctx).Error("refund transfer in error", "err", sdkErr.Error())
			panic(sdkErr)
		}
		return sdk.ExecuteResult{
			Payload: refundPackage,
			Err:     errors.Wrapf(types.ErrInvalidLength, "transfer in package is expired"),
		}
	}

	for idx, receiverAddr := range transferInPackage.ReceiverAddresses {
		amount := sdk.NewCoin(symbol, sdk.NewIntFromBigInt(transferInPackage.Amounts[idx]))
		err = app.bridgeKeeper.bankKeeper.SendCoinsFromModuleToAccount(ctx, crosschaintypes.ModuleName, receiverAddr, sdk.Coins{amount})
		if err != nil {
			app.bridgeKeeper.Logger(ctx).Error("send coins error", "err", err.Error())
			refundPackage, err := app.bridgeKeeper.GetRefundTransferInPayload(transferInPackage, types.InsufficientBalance)
			if err != nil {
				app.bridgeKeeper.Logger(ctx).Error("get refund transfer in payload error", "err", err.Error())
				panic(err)
			}
			return sdk.ExecuteResult{
				Payload: refundPackage,
				Err:     errors.Wrapf(types.ErrInvalidLength, "balance of cross chain module is insufficient"),
			}
		}
	}

	return sdk.ExecuteResult{}
}
