package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	"github.com/bnb-chain/greenfield/types/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	gnfdtypes "github.com/bnb-chain/greenfield/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
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
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.GetAuthority() != req.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	// Some parameters cannot be modified
	originParams := k.GetParams(ctx)
	if req.Params.GvgStakingPerBytes != originParams.GvgStakingPerBytes || req.Params.DepositDenom != originParams.DepositDenom {
		return nil, errors.ErrInvalidParameter.Wrap("GvgStakingPerBytes and depositDenom are not allow to update")
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) CreateGlobalVirtualGroup(goCtx context.Context, req *types.MsgCreateGlobalVirtualGroup) (*types.MsgCreateGlobalVirtualGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	var gvgStatisticsWithinSPs []*types.GVGStatisticsWithinSP

	spOperatorAddr := sdk.MustAccAddressFromHex(req.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spOperatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
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

	gvgFamily, err := k.GetOrCreateEmptyGVGFamily(ctx, req.FamilyId, sp.Id)
	if err != nil {
		return nil, err
	}

	// Each family supports only a limited number of GVGS
	if k.MaxGlobalVirtualGroupNumPerFamily(ctx) < uint32(len(gvgFamily.GlobalVirtualGroupIds)) {
		return nil, types.ErrLimitationExceed.Wrapf("The gvg number within the family exceeds the limit.")
	}

	// deposit enough tokens for oncoming objects
	coins := sdk.NewCoins(sdk.NewCoin(k.DepositDenomForGVG(ctx), req.Deposit.Amount))
	err = k.bankKeeper.SendCoinsFromAccountToModule(ctx, sdk.MustAccAddressFromHex(sp.FundingAddress), types.ModuleName, coins)
	if err != nil {
		return nil, err
	}

	gvgID := k.GenNextGVGID(ctx)
	gvg := &types.GlobalVirtualGroup{
		Id:                    gvgID,
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

	spOperatorAddr := sdk.MustAccAddressFromHex(req.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spOperatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
	}

	err := k.DeleteGVG(ctx, sp, req.GlobalVirtualGroupId)
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

	addr := sdk.MustAccAddressFromHex(req.StorageProvider)

	var sp *sptypes.StorageProvider
	found := false
	sp, found = k.spKeeper.GetStorageProviderByOperatorAddr(ctx, addr)
	if !found {
		sp, found = k.spKeeper.GetStorageProviderByFundingAddr(ctx, addr)
		if !found {
			return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
		}
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

	addr := sdk.MustAccAddressFromHex(req.StorageProvider)
	var sp *sptypes.StorageProvider
	found := false
	sp, found = k.spKeeper.GetStorageProviderByOperatorAddr(ctx, addr)
	if !found {
		sp, found = k.spKeeper.GetStorageProviderByFundingAddr(ctx, addr)
		if !found {
			return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
		}
	}

	gvg, found := k.GetGVG(ctx, req.GlobalVirtualGroupId)
	if !found {
		return nil, types.ErrGVGNotExist
	}

	if gvg.PrimarySpId != sp.Id {
		return nil, types.ErrWithdrawFailed.Wrapf("the withdrawer(spID: %d) is not the primary sp(ID:%d) of gvg.", sp.Id, gvg.PrimarySpId)
	}

	depositDenom := k.DepositDenomForGVG(ctx)
	if req.Withdraw.Denom != depositDenom {
		return nil, sdkerrors.Wrapf(types.ErrInvalidDenom, "invalid coin denomination: got %s, expected %s", req.Withdraw.Denom, k.DepositDenomForGVG(ctx))
	}

	var withdrawTokens math.Int

	availableTokens := k.GetAvailableStakingTokens(ctx, gvg)
	if availableTokens.IsNegative() {
		panic("the available tokens is negative when withdraw")
	}
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

func (k msgServer) SwapOut(goCtx context.Context, msg *types.MsgSwapOut) (*types.MsgSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
	}

	successorSP, found := k.spKeeper.GetStorageProvider(ctx, msg.SuccessorSpId)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("successor sp not found.")
	}

	// verify the approval
	err := gnfdtypes.VerifySignature(sdk.MustAccAddressFromHex(successorSP.ApprovalAddress), sdk.Keccak256(msg.GetApprovalBytes()), msg.SuccessorSpApproval.Sig)
	if err != nil {
		return nil, err
	}

	err = k.SetSwapOutInfo(ctx, msg.GlobalVirtualGroupFamilyId, msg.GlobalVirtualGroupIds, sp.Id, successorSP.Id)
	if err != nil {
		return nil, err
	}

	if err = ctx.EventManager().EmitTypedEvents(&types.EventSwapOut{
		StorageProviderId:          sp.Id,
		GlobalVirtualGroupFamilyId: msg.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupIds:      msg.GlobalVirtualGroupIds,
		SuccessorSpId:              successorSP.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgSwapOutResponse{}, nil
}

func (k msgServer) CancelSwapOut(goCtx context.Context, msg *types.MsgCancelSwapOut) (*types.MsgCancelSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
	}

	err := k.DeleteSwapOutInfo(ctx, msg.GlobalVirtualGroupFamilyId, msg.GlobalVirtualGroupIds, sp.Id)
	if err != nil {
		return nil, err
	}

	if err = ctx.EventManager().EmitTypedEvents(&types.EventCancelSwapOut{
		StorageProviderId:          sp.Id,
		GlobalVirtualGroupFamilyId: msg.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupIds:      msg.GlobalVirtualGroupIds,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCancelSwapOutResponse{}, nil
}

func (k msgServer) CompleteSwapOut(goCtx context.Context, msg *types.MsgCompleteSwapOut) (*types.MsgCompleteSwapOutResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
	}

	err := k.Keeper.CompleteSwapOut(ctx, msg.GlobalVirtualGroupFamilyId, msg.GlobalVirtualGroupIds, sp)
	if err != nil {
		return nil, err
	}

	if err = ctx.EventManager().EmitTypedEvents(&types.EventCompleteSwapOut{
		StorageProviderId:          sp.Id,
		GlobalVirtualGroupFamilyId: msg.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupIds:      msg.GlobalVirtualGroupIds,
	}); err != nil {
		return nil, err
	}
	return &types.MsgCompleteSwapOutResponse{}, nil
}

func (k msgServer) Settle(goCtx context.Context, req *types.MsgSettle) (*types.MsgSettleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := sdk.MustAccAddressFromHex(req.StorageProvider)
	var sp *sptypes.StorageProvider
	found := false
	sp, found = k.spKeeper.GetStorageProviderByOperatorAddr(ctx, addr)
	if !found {
		sp, found = k.spKeeper.GetStorageProviderByFundingAddr(ctx, addr)
		if !found {
			return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
		}
	}

	if req.GlobalVirtualGroupFamilyId != types.NoSpecifiedFamilyId {
		family, found := k.GetGVGFamily(ctx, sp.Id, req.GlobalVirtualGroupFamilyId)
		if !found {
			return nil, types.ErrGVGFamilyNotExist
		}

		err := k.SettleAndDistributeGVGFamily(ctx, sp, family)
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

	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
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

	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
	}

	if sp.Status != sptypes.STATUS_GRACEFUL_EXITING {
		return nil, sptypes.ErrStorageProviderExitFailed.Wrapf(
			"sp(id : %d, operator address: %s) not in the process of exiting", sp.Id, sp.OperatorAddress)
	}

	err := k.StorageProviderExitable(ctx, sp.Id)
	if err != nil {
		return nil, err
	}

	// send back the total deposit
	coins := sdk.NewCoins(sdk.NewCoin(k.spKeeper.DepositDenomForSP(ctx), sp.TotalDeposit))
	err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, sptypes.ModuleName, sdk.MustAccAddressFromHex(sp.FundingAddress), coins)
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
