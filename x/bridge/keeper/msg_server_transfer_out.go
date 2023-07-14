package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"

	"github.com/bnb-chain/greenfield/x/bridge/types"
)

func (k msgServer) TransferOut(goCtx context.Context, msg *types.MsgTransferOut) (*types.MsgTransferOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, errors.Wrapf(types.ErrUnsupportedDenom, "denom is not supported")
	}

	relayerFeeAmount, ackRelayerFeeAmount := k.GetTransferOutRelayerFee(ctx)
	totalRelayerFee := relayerFeeAmount.Add(ackRelayerFeeAmount)

	relayerFee := sdk.Coin{
		Denom:  bondDenom,
		Amount: totalRelayerFee,
	}
	transferAmount := sdk.Coins{*msg.Amount}.Add(relayerFee)

	fromAddress := sdk.MustAccAddressFromHex(msg.From)
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fromAddress, crosschaintypes.ModuleName, transferAmount)
	if err != nil {
		return nil, err
	}

	toAddress := sdk.MustAccAddressFromHex(msg.To)

	transferPackage := types.TransferOutSynPackage{
		RefundAddress: fromAddress,
		Recipient:     toAddress,
		Amount:        msg.Amount.Amount.BigInt(),
	}

	encodedPackage, err := transferPackage.Serialize()
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidPackage, "encode transfer out package error")
	}

	sendSeq, err := k.crossChainKeeper.CreateRawIBCPackageWithFee(ctx, k.crossChainKeeper.GetDestBscChainID(), types.TransferOutChannelID, sdk.SynCrossChainPackageType,
		encodedPackage, relayerFeeAmount.BigInt(), ackRelayerFeeAmount.BigInt())
	if err != nil {
		return nil, err
	}

	// emit event
	transferOutEvent := types.EventCrossTransferOut{
		From:        fromAddress.String(),
		To:          toAddress.String(),
		Amount:      msg.Amount,
		RelayerFee:  &relayerFee,
		Sequence:    sendSeq,
		DestChainId: uint32(k.crossChainKeeper.GetDestBscChainID()),
	}
	err = ctx.EventManager().EmitTypedEvent(&transferOutEvent)
	if err != nil {
		return nil, err
	}

	return &types.MsgTransferOutResponse{}, nil
}
