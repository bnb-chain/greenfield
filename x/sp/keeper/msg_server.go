package keeper

import (
	"context"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc := sdk.MustAccAddressFromHex(msg.SpAddress)
	fundingAcc := sdk.MustAccAddressFromHex(msg.FundingAddress)

	fundingAccount := k.accountKeeper.GetAccount(ctx, fundingAcc)
	if fundingAccount == nil {
		return nil, status.Errorf(codes.NotFound, "account %s not found", msg.FundingAddress)
	}

	sealAcc := sdk.MustAccAddressFromHex(msg.SealAddress)
	approvalAcc := sdk.MustAccAddressFromHex(msg.ApprovalAddress)
	gcAcc := sdk.MustAccAddressFromHex(msg.GcAddress)

	signers := msg.GetSigners()
	if ctx.BlockHeight() == 0 {
		if len(signers) != 1 || !signers[0].Equals(spAcc) {
			return nil, types.ErrSignerNotSPOperator
		}
	} else {
		if len(signers) != 1 || !signers[0].Equals(k.accountKeeper.GetModuleAddress(gov.ModuleName)) {
			return nil, types.ErrSignerNotGovModule
		}
	}

	if _, found := k.GetStorageProvider(ctx, spAcc); found {
		return nil, types.ErrStorageProviderOwnerExists
	}

	// check to see if the funding address has been registered before
	if _, found := k.GetStorageProviderByFundingAddr(ctx, fundingAcc); found {
		return nil, types.ErrStorageProviderFundingAddrExists
	}

	// check to see if the seal address has been registered before
	if _, found := k.GetStorageProviderBySealAddr(ctx, sealAcc); found {
		return nil, types.ErrStorageProviderSealAddrExists
	}

	// check to see if the approval address has been registered before
	if _, found := k.GetStorageProviderByApprovalAddr(ctx, approvalAcc); found {
		return nil, types.ErrStorageProviderApprovalAddrExists
	}

	// check to see if the gc address has been registered before
	if _, found := k.GetStorageProviderByGcAddr(ctx, gcAcc); found {
		return nil, types.ErrStorageProviderGcAddrExists
	}

	if err := msg.Description.EnsureLength(); err != nil {
		return nil, err
	}

	if msg.Deposit.Amount.LT(k.MinDeposit(ctx)) {
		return nil, types.ErrInsufficientDepositAmount
	}

	depositDenom := k.DepositDenomForSP(ctx)
	if depositDenom != msg.Deposit.GetDenom() {
		return nil, errors.Wrapf(types.ErrInvalidDenom, "invalid coin denomination: got %s, expected %s", msg.Deposit.Denom, depositDenom)
	}

	// check the deposit authorization from the fund address to gov module account
	if ctx.BlockHeight() != 0 {
		err := k.CheckDepositAuthorization(
			ctx,
			k.accountKeeper.GetModuleAddress(gov.ModuleName),
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

	sp, err := types.NewStorageProvider(spAcc, fundingAcc, sealAcc, approvalAcc, gcAcc,
		msg.Deposit.Amount, msg.Endpoint, msg.Description)
	if err != nil {
		return nil, err
	}

	k.SetStorageProvider(ctx, &sp)
	k.SetStorageProviderByApprovalAddr(ctx, &sp)
	k.SetStorageProviderByFundingAddr(ctx, &sp)
	k.SetStorageProviderBySealAddr(ctx, &sp)
	k.SetStorageProviderByGcAddr(ctx, &sp)

	// set initial sp storage price
	spStoragePrice := types.SpStoragePrice{
		SpAddress:     spAcc.String(),
		UpdateTimeSec: ctx.BlockTime().Unix(),
		ReadPrice:     msg.ReadPrice,
		StorePrice:    msg.StorePrice,
		FreeReadQuota: msg.FreeReadQuota,
	}
	k.SetSpStoragePrice(ctx, spStoragePrice)
	err = k.UpdateSecondarySpStorePrice(ctx)
	if err != nil {
		return nil, err
	}

	if err = ctx.EventManager().EmitTypedEvents(&types.EventCreateStorageProvider{
		SpAddress:       spAcc.String(),
		FundingAddress:  fundingAcc.String(),
		SealAddress:     sealAcc.String(),
		ApprovalAddress: approvalAcc.String(),
		GcAddress:       gcAcc.String(),
		Endpoint:        msg.Endpoint,
		TotalDeposit:    &msg.Deposit,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCreateStorageProviderResponse{}, nil
}

// EditStorageProvider defines a method for editing a existing storage provider
func (k msgServer) EditStorageProvider(goCtx context.Context, msg *types.MsgEditStorageProvider) (*types.MsgEditStorageProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spAcc := sdk.MustAccAddressFromHex(msg.SpAddress)

	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	changed := false

	// replace endpoint
	if len(msg.Endpoint) != 0 {
		oldEndpoint := sp.Endpoint
		sp.Endpoint = msg.Endpoint
		changed = true
		// TODO(chris): when we need more fields to EventEditStorageProvider, it should be put before calling the SetStorageProvider method.
		if err := ctx.EventManager().EmitTypedEvents(&types.EventEditStorageProvider{
			OldEndpoint: oldEndpoint,
			NewEndpoint: sp.Endpoint,
		}); err != nil {
			return nil, err
		}
	}

	if msg.Description != nil {
		description, err := sp.Description.UpdateDescription(msg.Description)
		if err != nil {
			return nil, err
		}
		sp.Description = *description
		changed = true
	}

	if !changed {
		return nil, types.ErrStorageProviderNotChanged
	}
	k.SetStorageProvider(ctx, sp)

	return &types.MsgEditStorageProviderResponse{}, nil
}

// Deposit defines a method for deposit token from fund address.
func (k msgServer) Deposit(goCtx context.Context, msg *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fundAcc := sdk.MustAccAddressFromHex(msg.Creator)

	sp, found := k.GetStorageProviderByFundingAddr(ctx, fundAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	depositDenom := k.DepositDenomForSP(ctx)
	if depositDenom != msg.Deposit.GetDenom() {
		return nil, errors.Wrapf(types.ErrInvalidDenom, "invalid coin denomination: got %s, expected %s", msg.Deposit.Denom, depositDenom)
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
		FundingAddress: msg.Creator,
		Deposit:        msg.Deposit.String(),
		TotalDeposit:   sp.TotalDeposit.String(),
	}); err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{}, nil
}

func (k msgServer) UpdateSpStoragePrice(goCtx context.Context, msg *types.MsgUpdateSpStoragePrice) (*types.MsgUpdateSpStoragePriceResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	spAcc := sdk.MustAccAddressFromHex(msg.SpAddress)
	err := k.IsStorageProviderExistAndInService(ctx, spAcc)
	if err != nil {
		return nil, errors.Wrapf(err, "IsStorageProviderExistAndInService return err")
	}

	current := ctx.BlockTime().Unix()
	spStorePrice := types.SpStoragePrice{
		UpdateTimeSec: current,
		SpAddress:     spAcc.String(),
		ReadPrice:     msg.ReadPrice,
		StorePrice:    msg.StorePrice,
		FreeReadQuota: msg.FreeReadQuota,
	}
	k.SetSpStoragePrice(ctx, spStorePrice)
	err = k.UpdateSecondarySpStorePrice(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "update secondary sp store price failed")
	}
	return &types.MsgUpdateSpStoragePriceResponse{}, nil
}
