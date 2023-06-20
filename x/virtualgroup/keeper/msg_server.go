package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	gnfdtypes "github.com/bnb-chain/greenfield/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) CreateGlobalVirtualGroup(goCtx context.Context, req *types.MsgCreateGlobalVirtualGroup) (*types.MsgCreateGlobalVirtualGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var gvgStatisticsWithinSPs []*types.GVGStatisticsWithinSP

	spOperatorAddr := sdk.MustAccAddressFromHex(req.PrimarySpAddress)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spOperatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}
	var secondarySpIds []uint32
	for _, id := range req.SecondarySpIds {
		ssp, found := k.spKeeper.GetStorageProvider(ctx, id)
		if !found {
			return nil, sdkerrors.Wrapf(sptypes.ErrStorageProviderNotFound, "secondary sp not found, ID: %d", id)
		}
		secondarySpIds = append(secondarySpIds, ssp.Id)
		gvgStatisticsWithinSP := k.GetOrCreateGVGStatisticsWithinSP(ctx, ssp.Id)
		gvgStatisticsWithinSP.SecondaryCount++
		gvgStatisticsWithinSPs = append(gvgStatisticsWithinSPs, gvgStatisticsWithinSP)
	}

	// TODO(fynn): add some limit for gvgs in a family
	gvgFamily, err := k.GetOrCreateEmptyGVGFamily(ctx, req.FamilyId, sp.Id)
	if err != nil {
		return nil, err
	}

	gvgID := k.GenNextGVGID(ctx)
	if gvgID == 0 {
		return nil, sdkerrors.Wrapf(types.ErrGenSequenceIDError, "wrong next gvg id.")
	}

	// TODO: Add gvg number limit for a family
	// deposit enough tokens for oncoming objects
	coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForGVG(ctx), req.Deposit.Amount))
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromHex(sp.FundingAddress), types.ModuleName, coins)
	if err != nil {
		return nil, err
	}

	gvg := &types.GlobalVirtualGroup{
		Id:                    k.GenNextGVGID(ctx),
		FamilyId:              gvgFamily.Id,
		PrimarySpId:           sp.Id,
		SecondarySpIds:        secondarySpIds,
		StoredSize:            0,
		VirtualPaymentAddress: k.DeriveVirtualPaymentAccount(types.GVGVirtualPaymentAccountName, gvgID).String(),
		TotalDeposit:          req.Deposit.Amount,
	}

	gvgFamily.AppendGVG(gvg.Id)

	k.SetGVG(ctx, gvg)
	k.SetGVGFamily(ctx, gvg.PrimarySpId, gvgFamily)
	k.BatchSetGVGStatisticsWithinSP(ctx, gvgStatisticsWithinSPs)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateGlobalVirtualGroup{
		Id:                    gvg.Id,
		FamilyId:              gvg.FamilyId,
		PrimarySpId:           gvg.PrimarySpId,
		SecondarySpIds:        gvg.SecondarySpIds,
		StoredSize:            gvg.StoredSize,
		VirtualPaymentAddress: gvg.VirtualPaymentAddress,
		TotalDeposit:          gvg.TotalDeposit,
	}); err != nil {
		return nil, err
	}
	if req.FamilyId == types.NoSpecifiedFamilyId {
		if err := ctx.EventManager().EmitTypedEvents(&types.EventCreateGlobalVirtualGroupFamily{
			Id:                    gvg.Id,
			VirtualPaymentAddress: gvgFamily.VirtualPaymentAddress,
		}); err != nil {
			return nil, err
		}
	}
	return &types.MsgCreateGlobalVirtualGroupResponse{}, nil
}

func (k msgServer) DeleteGlobalVirtualGroup(goCtx context.Context, req *types.MsgDeleteGlobalVirtualGroup) (*types.MsgDeleteGlobalVirtualGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	spOperatorAddr := sdk.MustAccAddressFromHex(req.PrimarySpAddress)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spOperatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	err := k.DeleteGVG(ctx, sp.Id, req.GlobalVirtualGroupId)
	if err != nil {
		return nil, err
	}

	if err = ctx.EventManager().EmitTypedEvents(&types.EventDeleteGlobalVirtualGroup{
		Id: req.GlobalVirtualGroupId,
	}); err != nil {
		return nil, err
	}
	return &types.MsgDeleteGlobalVirtualGroupResponse{}, nil
}

func (k msgServer) Deposit(goCtx context.Context, req *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	funcAcc := sdk.MustAccAddressFromHex(req.FundingAddress)

	sp, found := k.spKeeper.GetStorageProviderByFundingAddr(ctx, funcAcc)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	gvg, found := k.GetGVG(ctx, req.GlobalVirtualGroupId)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	depositDenom := k.DepositDenomForGVG(ctx)
	if depositDenom != req.Deposit.GetDenom() {
		return nil, sdkerrors.Wrapf(types.ErrInvalidDenom, "invalid coin denomination: got %s, expected %s", req.Deposit.Denom, depositDenom)
	}

	// deposit the deposit token to module account.
	coins := sdk.NewCoins(sdk.NewCoin(depositDenom, req.Deposit.Amount))
	if err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, sp.GetFundingAccAddress(), types.ModuleName, coins); err != nil {
		return nil, err
	}

	gvg.TotalDeposit = gvg.TotalDeposit.Add(req.Deposit.Amount)
	k.SetGVG(ctx, gvg)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGlobalVirtualGroup{
		Id:           req.GlobalVirtualGroupId,
		StoreSize:    gvg.StoredSize,
		TotalDeposit: gvg.TotalDeposit,
	}); err != nil {
		return nil, err
	}
	return &types.MsgDepositResponse{}, nil
}

func (k msgServer) Withdraw(goCtx context.Context, req *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	funcAcc := sdk.MustAccAddressFromHex(req.FundingAddress)

	sp, found := k.spKeeper.GetStorageProviderByFundingAddr(ctx, funcAcc)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	gvg, found := k.GetGVG(ctx, req.GlobalVirtualGroupId)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	depositDenom := k.DepositDenomForGVG(ctx)
	if req.Withdraw.Denom != depositDenom {
		return nil, sdkerrors.Wrapf(types.ErrInvalidDenom, "invalid coin denomination: got %s, expected %s", req.Withdraw.Denom, k.DepositDenomForGVG(ctx))
	}

	var withdrawTokens math.Int

	availableTokens := k.GetAvailableStakingTokens(ctx, gvg)
	if req.Withdraw.Amount.IsZero() {
		withdrawTokens = availableTokens
	} else {
		if availableTokens.LT(req.Withdraw.Amount) {
			return nil, types.ErrWithdrawAmountTooLarge
		}
		withdrawTokens = req.Withdraw.Amount
	}

	// withdraw the deposit token from module account to funding account.
	coins := sdk.NewCoins(sdk.NewCoin(depositDenom, withdrawTokens))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sp.GetFundingAccAddress(), coins); err != nil {
		return nil, err
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGlobalVirtualGroup{
		Id:           req.GlobalVirtualGroupId,
		StoreSize:    gvg.StoredSize,
		TotalDeposit: gvg.TotalDeposit,
	}); err != nil {
		return nil, err
	}

	return &types.MsgWithdrawResponse{}, nil
}

func (k msgServer) SwapOut(goCtx context.Context, req *types.MsgSwapOut) (*types.MsgSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	operatorAddr := sdk.MustAccAddressFromHex(req.OperatorAddress)
	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	successorSP, found := k.spKeeper.GetStorageProvider(ctx, req.SuccessorSpId)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("successor sp not found.")
	}

	// verify the approval
	err := gnfdtypes.VerifySignature(sdk.MustAccAddressFromHex(successorSP.ApprovalAddress), sdk.Keccak256(req.GetApprovalBytes()), req.SuccessorSpApproval.Sig)
	if err != nil {
		return nil, err
	}

	if req.GlobalVirtualGroupFamilyId == types.NoSpecifiedFamilyId {
		// if the family id is not specified, it means that the SP will swap out as a secondary SP.
		err := k.SwapOutAsSecondarySP(ctx, sp.Id, successorSP.Id, req.GlobalVirtualGroupIds)
		if err != nil {
			return nil, err
		}
	} else {
		// if the family id is specified, it means that the SP will swap out as a primary SP and the successor sp will
		// take over all the gvg of this family
		err := k.SwapOutAsPrimarySP(ctx, sp, successorSP, req.GlobalVirtualGroupFamilyId)
		if err != nil {
			return nil, err
		}
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventSwapOut{
		StorageProviderId:          sp.Id,
		GlobalVirtualGroupFamilyId: req.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupIds:      req.GlobalVirtualGroupIds,
		SuccessorSpId:              successorSP.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgSwapOutResponse{}, nil
}

func (k msgServer) Settle(goCtx context.Context, req *types.MsgSettle) (*types.MsgSettleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	funcAcc := sdk.MustAccAddressFromHex(req.FundingAddress)

	sp, found := k.spKeeper.GetStorageProviderByFundingAddr(ctx, funcAcc)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	if req.GlobalVirtualGroupFamilyId != types.NoSpecifiedFamilyId {
		family, found := k.GetGVGFamily(ctx, sp.Id, req.GlobalVirtualGroupFamilyId)
		if !found {
			return nil, types.ErrGVGFamilyNotExist
		}

		err := k.SettleAndDistributeGVGFamily(ctx, sp.Id, family)
		if err != nil {
			return nil, types.ErrSettleFailed
		}
	} else {
		m := make(map[uint32]struct{})
		for _, gvgID := range req.GlobalVirtualGroupIds {
			m[gvgID] = struct{}{}
		}
		for gvgID := range m {
			gvg, found := k.GetGVG(ctx, gvgID)
			if !found {
				return nil, types.ErrGVGNotExist
			}

			permitted := false
			for _, id := range gvg.SecondarySpIds {
				if id == sp.Id {
					permitted = true
					break
				}
			}
			if !permitted {
				return nil, sdkerrors.Wrapf(types.ErrSettleFailed, "storage provider %d is not in the group", sp.Id)
			}

			err := k.SettleAndDistributeGVG(ctx, gvg)
			if err != nil {
				return nil, types.ErrSettleFailed
			}
		}
	}

	return &types.MsgSettleResponse{}, nil
}

func (k msgServer) StorageProviderExit(goCtx context.Context, msg *types.MsgStorageProviderExit) (*types.MsgStorageProviderExitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.OperatorAddress)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	if sp.Status != sptypes.STATUS_IN_SERVICE {
		return nil, sptypes.ErrStorageProviderExitFailed.Wrapf("sp not in service, status: %s", sp.Status.String())
	}

	sp.Status = sptypes.STATUS_GRACEFUL_EXITING

	k.spKeeper.SetStorageProvider(ctx, sp)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventStorageProviderExit{
		StorageProviderId: sp.Id,
		OperatorAddress:   sp.OperatorAddress,
	}); err != nil {
		return nil, err
	}
	return &types.MsgStorageProviderExitResponse{}, nil
}

func (k msgServer) CompleteStorageProviderExit(goCtx context.Context, msg *types.MsgCompleteStorageProviderExit) (*types.MsgCompleteStorageProviderExitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddress := sdk.MustAccAddressFromHex(msg.OperatorAddress)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddress)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound
	}

	if sp.Status != sptypes.STATUS_GRACEFUL_EXITING {
		return nil, sptypes.ErrStorageProviderExitFailed.Wrapf(
			"sp(id : %d, operator address: %s) not in the process of exiting", sp.Id, sp.OperatorAddress)
	}

	err := k.IsStorageProviderCanExit(ctx, sp.Id)
	if err != nil {
		return nil, err
	}

	err = k.spKeeper.Exit(ctx, sp)
	if err != nil {
		return nil, err
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventCompleteStorageProviderExit{
		StorageProviderId: sp.Id,
		OperatorAddress:   sp.OperatorAddress,
		TotalDeposit:      sp.TotalDeposit,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCompleteStorageProviderExitResponse{}, nil
}
