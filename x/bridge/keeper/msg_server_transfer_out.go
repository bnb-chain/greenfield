package keeper

import (
	"context"
	"fmt"

	"cosmossdk.io/errors"
	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/bsc/rlp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
)

func (k msgServer) TransferOut(goCtx context.Context, msg *types.MsgTransferOut) (*types.MsgTransferOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	bondDenom := k.stakingKeeper.BondDenom(ctx)
	if msg.Amount.Denom != bondDenom {
		return nil, errors.Wrapf(types.ErrUnsupportedDenom, fmt.Sprintf("denom is not supported"))
	}

	relayFee := sdk.Coin{
		Denom:  bondDenom,
		Amount: types.CrossTransferOutRelayFee,
	}
	transferAmount := sdk.Coins{*msg.Amount}.Add(relayFee)

	fromAddress := msg.GetSigners()[0]
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fromAddress, crosschaintypes.ModuleName, transferAmount)
	if err != nil {
		return nil, err
	}

	toAddress, err := sdk.ETHAddressFromHexUnsafe(msg.To)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidAddress, fmt.Sprintf("to address is not invalid"))
	}

	transferPackage := types.TransferOutSynPackage{
		RefundAddress: fromAddress.Bytes(),
		Recipient:     toAddress,
		Amount:        msg.Amount.Amount.BigInt(),
	}

	encodedPackage, err := rlp.EncodeToBytes(transferPackage)
	if err != nil {
		return nil, errors.Wrapf(types.ErrInvalidPackage, fmt.Sprintf("encode transfer out package error"))
	}

	sendSeq, err := k.crossChainKeeper.CreateRawIBCPackageWithFee(ctx, k.DestChainId, types.TransferOutChannelID, sdk.SynCrossChainPackageType,
		encodedPackage, *relayFee.Amount.BigInt())
	if err != nil {
		return nil, err
	}

	// emit event
	transferOutEvent := types.EventCrossTransferOut{
		From:       fromAddress.String(),
		To:         toAddress.String(),
		Amount:     msg.Amount,
		RelayerFee: &relayFee,
		Sequence:   sendSeq,
	}
	ctx.EventManager().EmitTypedEvent(&transferOutEvent)

	return &types.MsgTransferOutResponse{}, nil
}
