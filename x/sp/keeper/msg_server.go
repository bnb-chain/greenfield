package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/bnb-chain/greenfield/x/sp/types"
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
	// TODO: check if a valid endpoint
	ctx := sdk.UnwrapSDKContext(goCtx)

	signers := msg.GetSigners()
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

	sealAcc, err := sdk.AccAddressFromHexUnsafe(msg.SealAddress)
	if err != nil {
		return nil, err
	}

	approvalAcc, err := sdk.AccAddressFromHexUnsafe(msg.ApprovalAddress)
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

	depositDenom := k.DepositDenomForSP(ctx)
	if depositDenom != msg.Deposit.GetDenom() {
		return nil, errors.Wrapf(types.ErrInvalidDepositDenom, "invalid coin denomination: got %s, expected %s", msg.Deposit.Denom, depositDenom)
	}

	// check the deposit authorization from the fund address to gov module account
	if ctx.BlockHeader().Height != 0 {
		err = k.CheckDepositAuthorization(
			ctx,
			k.authKeeper.GetModuleAddress(gov.ModuleName),
			fundingAcc,
			types.NewMsgDeposit(fundingAcc, spAcc, msg.Deposit))
		if err != nil {
			return nil, err
		}
	}
	// deposit coins to module account. move coins from sp address account to module account.
	// Requires FeeGrant module authorization
	coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForSP(ctx), msg.Deposit.Amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fundingAcc, types.ModuleName, coins); err != nil {
		return nil, err
	}

	sp, err := types.NewStorageProvider(spAcc, fundingAcc, sealAcc, approvalAcc, msg.Deposit.Amount, msg.Endpoint, msg.Description)
	if err != nil {
		return nil, err
	}

	k.SetStorageProvider(ctx, sp)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateStorageProvider{
		SpAddress:       spAcc.String(),
		FundingAddress:  fundingAcc.String(),
		SealAddress:     sealAcc.String(),
		ApprovalAddress: approvalAcc.String(),
		Endpoint:        msg.Endpoint,
		TotalDeposit:    msg.Deposit.String(),
	}); err != nil {
		return nil, err
	}
	return &types.MsgCreateStorageProviderResponse{}, nil
}

// EditStorageProvider defines a method for editing a existing storage provider
func (k msgServer) EditStorageProvider(goCtx context.Context, msg *types.MsgEditStorageProvider) (*types.MsgEditStorageProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return nil, err
	}

	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	oldEndpoint := sp.Endpoint
	if _, err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	description, err := sp.Description.UpdateDescription(msg.Description)
	if err != nil {
		return nil, err
	}

	sp.Description = description

	k.SetStorageProvider(ctx, sp)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventEditStorageProvider{
		OldEndpoint: oldEndpoint,
		NewEndpoint: sp.Endpoint,
	}); err != nil {
		return nil, err
	}
	return &types.MsgEditStorageProviderResponse{}, nil
}

// Deposit defines a method for deposit token from fund address.
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc, err := sdk.AccAddressFromHexUnsafe(msg.SpAddress)
	if err != nil {
		return nil, err
	}

	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	// Only funding address has permission to deposit tokens for SP
	if msg.Creator != sp.FundingAddress {
		return nil, types.ErrDepositAccountNotAllowed
	}

	depositDenom := k.DepositDenomForSP(ctx)
	if depositDenom != msg.Deposit.GetDenom() {
		return nil, errors.Wrapf(types.ErrInvalidDepositDenom, "invalid coin denomination: got %s, expected %s", msg.Deposit.Denom, depositDenom)
	}
	// deposit the deposit token to module account.
	coins := sdk.NewCoins(sdk.NewCoin(depositDenom, msg.Deposit.Amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sp.GetFundingAccAddress(), types.ModuleName, coins); err != nil {
		return nil, err
	}

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
