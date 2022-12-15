package keeper

import (
	"context"

	"github.com/bnb-chain/bfs/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
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

// CreateStorageProvider defines a method for creating a new storage provider
func (k msgServer) CreateStorageProvider(goCtx context.Context, msg *types.MsgCreateStorageProvider) (*types.MsgCreateStorageProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, types.ErrSignerEmpty
	}
	if len(signers) != 1 || !signers[0].Equals(k.authKeeper.GetModuleAddress(gov.ModuleName)) {
		return nil, types.ErrSignerNotGovModule
	}

	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return nil, err
	}

	fundingAcc, err := sdk.AccAddressFromHexUnsafe(msg.FundingAddress)
	if err != nil {
		return nil, err
	}

	if _, found := k.GetStorageProvider(ctx, spAcc); found {
		return nil, types.ErrStorageProviderOwnerExists
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	if msg.Deposit.Amount.LT(k.MinDeposit(ctx)) {
		return nil, types.ErrInsufficientDepositAmount
	}

	// check the deposit authorization from the fund address to gov module account
	if ctx.BlockHeader().Height != 0 {
		err = k.CheckDepositAuthorization(
			ctx,
			k.authKeeper.GetModuleAddress(gov.ModuleName),
			fundingAcc,
			types.NewMsgDeposit(msg.FundingAddress, msg.SpAddress, msg.Deposit))
		if err != nil {
			return nil, err
		}
	}

	sp, err := types.NewStorageProvider(spAcc, fundingAcc, msg.Description)
	if err != nil {
		return nil, err
	}

	k.SetStorageProvider(ctx, sp)

	// deposit coins to module account. move coins from sp address account to module account.
	// Requires FeeGrant module authorization
	coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForSP(ctx), msg.Deposit.Amount))
	k.bankKeeper.SendCoinsFromAccountToModule(ctx, fundingAcc, types.ModuleName, coins)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateStorageProvider{
		SpAddress:      spAcc.String(),
		FundingAddress: fundingAcc.String(),
		TotalDeposit:   msg.Deposit.String(),
	}); err != nil {
		return nil, err
	}
	return &types.MsgCreateStorageProviderResponse{}, nil
}

// EditStorageProvider defines a method for editing a existing storage provider
func (k msgServer) EditStorageProvider(goCtx context.Context, msg *types.MsgEditStorageProvider) (*types.MsgEditStorageProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	signer := msg.GetSigners()
	if len(signer) == 0 {
		return nil, types.ErrSignerEmpty
	}

	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return nil, err
	}

	storageProvider, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	description, err := storageProvider.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	storageProvider.Description = description

	k.SetStorageProvider(ctx, storageProvider)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventEditStorageProvider{}); err != nil {
		return nil, err
	}
	return &types.MsgEditStorageProviderResponse{}, nil
}

// Deposit defines a method for deposit token from fund address.
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	signers := msg.GetSigners()
	if len(signers) == 0 {
		return nil, types.ErrSignerEmpty
	}

	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return nil, err
	}

	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	// Only operator address and fund address has permission to deposit tokens for SP
	if msg.Creator == msg.SpAddress {
		// check the deposit authorization from the fund address to gov module account
		if ctx.BlockHeader().Height != 0 {
			err = k.CheckDepositAuthorization(
				ctx,
				k.authKeeper.GetModuleAddress(gov.ModuleName),
				sp.GetFundingAccAddress(),
				types.NewMsgDeposit(sp.FundingAddress, sp.GetOperatorAddress(), msg.Deposit))
			if err != nil {
				return nil, err
			}
		}
	} else if msg.Creator == sp.FundingAddress {
		// nothing todo
	} else {
		return nil, types.ErrDepositAccountNotAllowed
	}

	// deposit the deposit token to module account.
	coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForSP(ctx), msg.Deposit.Amount))
	k.bankKeeper.SendCoinsFromAccountToModule(ctx, sp.GetFundingAccAddress(), types.ModuleName, coins)

	// Add to storage provider's deposit tokens and update the storage provider.
	sp.TotalDeposit = sp.TotalDeposit.Add(msg.Deposit.Amount)
	k.SetStorageProvider(ctx, sp)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventDeposit{
		SpAddress:    msg.SpAddress,
		Deposit:      msg.Deposit.String(),
		TotalDeposit: sp.TotalDeposit.String(),
	}); err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{}, nil
}
