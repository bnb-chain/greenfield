package keeper

import (
	"context"

	sdkerrors "cosmossdk.io/errors"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	gnfdtypes "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/errors"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
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
	if req.Params.DepositDenom != originParams.DepositDenom {
		return nil, errors.ErrInvalidParameter.Wrapf("DepositDenom is not allow to update, current: %v, got: %v", originParams.DepositDenom, req.Params.DepositDenom)
	}

	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	if ctx.IsUpgraded(upgradetypes.Nagqu) {
		params := k.GetParams(ctx)
		_ = ctx.EventManager().EmitTypedEvents(&params)
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) CreateGlobalVirtualGroup(goCtx context.Context, req *types.MsgCreateGlobalVirtualGroup) (*types.MsgCreateGlobalVirtualGroupResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if ctx.IsUpgraded(upgradetypes.Pampas) {
		expectSecondarySPNum := int(k.storageKeeper.GetExpectSecondarySPNumForECObject(ctx, ctx.BlockTime().Unix()))
		if len(req.GetSecondarySpIds()) != expectSecondarySPNum {
			return nil, types.ErrInvalidSecondarySPCount.Wrapf("the number of secondary sp in the Global virtual group should be %d", expectSecondarySPNum)
		}
		spIdSet := make(map[uint32]struct{}, len(req.GetSecondarySpIds()))
		for _, spId := range req.GetSecondarySpIds() {
			if _, ok := spIdSet[spId]; ok {
				return nil, types.ErrDuplicateSecondarySP.Wrapf("the SP(id=%d) is duplicate in the Global virtual group.", spId)
			}
			spIdSet[spId] = struct{}{}
		}
	}

	var gvgStatisticsWithinSPs []*types.GVGStatisticsWithinSP

	spOperatorAddr := sdk.MustAccAddressFromHex(req.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spOperatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
	}

	if !sp.IsInService() && !sp.IsInMaintenance() {
		return nil, sptypes.ErrStorageProviderNotInService.Wrapf("sp is not in service or in maintenance, status: %s", sp.Status.String())
	}

	stat := k.GetOrCreateGVGStatisticsWithinSP(ctx, sp.Id)
	stat.PrimaryCount++
	gvgStatisticsWithinSPs = append(gvgStatisticsWithinSPs, stat)

	var secondarySpIds []uint32
	for _, id := range req.SecondarySpIds {
		ssp, found := k.spKeeper.GetStorageProvider(ctx, id)
		if !found {
			return nil, sdkerrors.Wrapf(sptypes.ErrStorageProviderNotFound, "secondary sp not found, ID: %d", id)
		}
		if !ssp.IsInService() && !ssp.IsInMaintenance() {
			return nil, sptypes.ErrStorageProviderNotInService.Wrapf("sp in GVG is not in service or in maintenance, status: %s", sp.Status.String())
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

	if ctx.IsUpgraded(upgradetypes.Manchurian) {
		for _, gvgID := range gvgFamily.GlobalVirtualGroupIds {
			gvg, found := k.GetGVG(ctx, gvgID)
			if !found {
				return nil, types.ErrGVGNotExist
			}
			for i, secondarySPId := range gvg.SecondarySpIds {
				if secondarySPId != secondarySpIds[i] {
					break
				}
				if i == len(secondarySpIds)-1 {
					return nil, types.ErrDuplicateGVG.Wrapf("the global virtual group family already has a GVG with same SP in same order")
				}
			}
		}
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
	k.SetGVGFamily(ctx, gvgFamily)
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
			Id:                    gvgFamily.Id,
			PrimarySpId:           gvgFamily.PrimarySpId,
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
		Id:          req.GlobalVirtualGroupId,
		PrimarySpId: sp.Id,
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
		Id:             gvg.Id,
		PrimarySpId:    gvg.PrimarySpId,
		StoreSize:      gvg.StoredSize,
		TotalDeposit:   gvg.TotalDeposit,
		SecondarySpIds: gvg.SecondarySpIds,
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

	if ctx.IsUpgraded(upgradetypes.Tundra) {
		gvg.TotalDeposit = gvg.TotalDeposit.Sub(withdrawTokens)
		k.SetGVG(ctx, gvg)
	}

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateGlobalVirtualGroup{
		Id:             req.GlobalVirtualGroupId,
		PrimarySpId:    gvg.PrimarySpId,
		StoreSize:      gvg.StoredSize,
		TotalDeposit:   gvg.TotalDeposit,
		SecondarySpIds: gvg.SecondarySpIds,
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

	if !successorSP.IsInService() {
		return nil, sptypes.ErrStorageProviderNotInService.Wrapf("successor sp is not in service, status: %s", sp.Status.String())
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

	return &types.MsgCompleteSwapOutResponse{}, nil
}

func (k msgServer) Settle(goCtx context.Context, req *types.MsgSettle) (*types.MsgSettleResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	addr := sdk.MustAccAddressFromHex(req.StorageProvider)
	var sp *sptypes.StorageProvider

	pampasUpgraded := ctx.IsUpgraded(upgradetypes.Pampas)
	if !pampasUpgraded {
		found := false
		sp, found = k.spKeeper.GetStorageProviderByOperatorAddr(ctx, addr)
		if !found {
			sp, found = k.spKeeper.GetStorageProviderByFundingAddr(ctx, addr)
			if !found {
				return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator/funding address of sp.")
			}
		}
	}

	if req.GlobalVirtualGroupFamilyId != types.NoSpecifiedFamilyId {
		family, found := k.GetGVGFamily(ctx, req.GlobalVirtualGroupFamilyId)
		if !found {
			return nil, types.ErrGVGFamilyNotExist
		}

		if pampasUpgraded {
			sp, found = k.spKeeper.GetStorageProvider(ctx, family.PrimarySpId)
			if !found {
				return nil, sptypes.ErrStorageProviderNotFound.Wrapf("Cannot find storage provider %d.", family.PrimarySpId)
			}
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

			if !pampasUpgraded {
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

	if ctx.IsUpgraded(upgradetypes.Hulunbeier) {
		stat, found := k.GetGVGStatisticsWithinSP(ctx, sp.Id)
		if found && stat.BreakRedundancyReqmtGvgCount != 0 {
			return nil, types.ErrSPCanNotExit.Wrapf("The SP has %d GVG that break the redundancy requirement, need to be resolved before exit.", stat.BreakRedundancyReqmtGvgCount)
		}

		// can only allow 1 sp exit at a time, a GVG can have only 1 SwapInInfo associated.
		exitingSPNum := uint32(0)
		sps := k.spKeeper.GetAllStorageProviders(ctx)
		maxSPExitingNum := k.SpConcurrentExitNum(ctx)

		for _, curSP := range sps {
			if curSP.Status == sptypes.STATUS_GRACEFUL_EXITING ||
				curSP.Status == sptypes.STATUS_FORCED_EXITING {
				exitingSPNum++
				if exitingSPNum >= maxSPExitingNum {
					return nil, sptypes.ErrStorageProviderExitFailed.Wrapf("There are %d SP exiting, only allow %d sp exit concurrently", exitingSPNum, maxSPExitingNum)
				}
			}
		}
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

	spAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be the operator address of sp.")
	}

	if sp.Status != sptypes.STATUS_GRACEFUL_EXITING && sp.Status != sptypes.STATUS_FORCED_EXITING {
		return nil, sptypes.ErrStorageProviderExitFailed.Wrapf(
			"sp(id : %d, operator address: %s) not in the process of exiting", sp.Id, sp.OperatorAddress)
	}

	err := k.StorageProviderExitable(ctx, sp.Id)
	if err != nil {
		return nil, err
	}

	var forcedExit bool
	if sp.Status == sptypes.STATUS_GRACEFUL_EXITING {
		// send back the total deposit
		coins := sdk.NewCoins(sdk.NewCoin(k.spKeeper.DepositDenomForSP(ctx), sp.TotalDeposit))
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, sptypes.ModuleName, sdk.MustAccAddressFromHex(sp.FundingAddress), coins)
		if err != nil {
			return nil, err
		}
	} else {
		forcedExit = true
		coins := sdk.NewCoins(sdk.NewCoin(k.spKeeper.DepositDenomForSP(ctx), sp.TotalDeposit))
		// the deposit will be transfer to the payment module gov addr stream record related bank account, when a stream record lack of
		// static balance, it will check for its related bank account
		err = k.bankKeeper.SendCoinsFromModuleToAccount(ctx, sptypes.ModuleName, paymenttypes.GovernanceAddress, coins)
		if err != nil {
			return nil, err
		}
	}
	err = k.spKeeper.Exit(ctx, sp)
	if err != nil {
		return nil, err
	}
	if ctx.IsUpgraded(upgradetypes.Hulunbeier) {
		if err = ctx.EventManager().EmitTypedEvents(&types.EventCompleteStorageProviderExit{
			StorageProviderId:      sp.Id,
			OperatorAddress:        msg.Operator,
			StorageProviderAddress: sp.OperatorAddress,
			TotalDeposit:           sp.TotalDeposit,
			ForcedExit:             forcedExit,
		}); err != nil {
			return nil, err
		}
	} else {
		if err = ctx.EventManager().EmitTypedEvents(&types.EventCompleteStorageProviderExit{
			StorageProviderId: sp.Id,
			OperatorAddress:   sp.OperatorAddress,
			TotalDeposit:      sp.TotalDeposit,
		}); err != nil {
			return nil, err
		}
	}
	return &types.MsgCompleteStorageProviderExitResponse{}, nil
}

func (k msgServer) ReserveSwapIn(goCtx context.Context, msg *types.MsgReserveSwapIn) (*types.MsgReserveSwapInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	successorSP, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be the operator address of sp.")
	}
	if successorSP.Id == msg.TargetSpId {
		return nil, types.ErrSwapInFailed.Wrapf("The SP(ID=%d) can not swap itself", successorSP.Id)
	}
	targetSP, found := k.spKeeper.GetStorageProvider(ctx, msg.TargetSpId)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("Target sp(ID=%d) try to swap not found.", msg.TargetSpId)
	}
	expirationTime := ctx.BlockTime().Unix() + int64(k.SwapInValidityPeriod(ctx))

	if err := k.Keeper.SwapIn(ctx, msg.GlobalVirtualGroupFamilyId, msg.GlobalVirtualGroupId, successorSP.Id, targetSP, expirationTime); err != nil {
		return nil, err
	}
	if err := ctx.EventManager().EmitTypedEvents(&types.EventReserveSwapIn{
		StorageProviderId:          successorSP.Id,
		GlobalVirtualGroupFamilyId: msg.GlobalVirtualGroupFamilyId,
		GlobalVirtualGroupId:       msg.GlobalVirtualGroupId,
		TargetSpId:                 msg.TargetSpId,
		ExpirationTime:             uint64(expirationTime),
	}); err != nil {
		return nil, err
	}
	return &types.MsgReserveSwapInResponse{}, nil
}

func (k msgServer) CancelSwapIn(goCtx context.Context, msg *types.MsgCancelSwapIn) (*types.MsgCancelSwapInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	successorSP, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
	}
	err := k.DeleteSwapInInfo(ctx, msg.GlobalVirtualGroupFamilyId, msg.GlobalVirtualGroupId, successorSP.Id)
	if err != nil {
		return nil, err
	}
	return &types.MsgCancelSwapInResponse{}, nil
}

func (k msgServer) CompleteSwapIn(goCtx context.Context, msg *types.MsgCompleteSwapIn) (*types.MsgCompleteSwapInResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)
	successorSP, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, operatorAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The address must be operator address of sp.")
	}
	err := k.Keeper.CompleteSwapIn(ctx, msg.GlobalVirtualGroupFamilyId, msg.GlobalVirtualGroupId, successorSP)
	if err != nil {
		return nil, err
	}
	return &types.MsgCompleteSwapInResponse{}, nil
}
func (k msgServer) StorageProviderForcedExit(goCtx context.Context, msg *types.MsgStorageProviderForcedExit) (*types.MsgStorageProviderForcedExitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if k.GetAuthority() != msg.Authority {
		return nil, sdkerrors.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), msg.Authority)
	}

	spAddr := sdk.MustAccAddressFromHex(msg.StorageProvider)

	sp, found := k.spKeeper.GetStorageProviderByOperatorAddr(ctx, spAddr)
	if !found {
		return nil, sptypes.ErrStorageProviderNotFound.Wrapf("The SP with operator address %s not found", msg.StorageProvider)
	}

	exitingSPNum := uint32(0)
	maxSPExitingNum := k.SpConcurrentExitNum(ctx)
	sps := k.spKeeper.GetAllStorageProviders(ctx)
	for _, curSP := range sps {
		if curSP.Status == sptypes.STATUS_GRACEFUL_EXITING ||
			curSP.Status == sptypes.STATUS_FORCED_EXITING {
			exitingSPNum++
			if exitingSPNum >= maxSPExitingNum {
				return nil, sptypes.ErrStorageProviderExitFailed.Wrapf("%d SP are exiting, allow %d sp exit concurrently", exitingSPNum, maxSPExitingNum)
			}
		}
	}

	// Governance can put an SP into force exiting status no matter what status it is in.
	sp.Status = sptypes.STATUS_FORCED_EXITING
	k.spKeeper.SetStorageProvider(ctx, sp)
	if err := ctx.EventManager().EmitTypedEvents(&types.EventStorageProviderForcedExit{
		StorageProviderId: sp.Id,
	}); err != nil {
		return nil, err
	}
	return &types.MsgStorageProviderForcedExitResponse{}, nil
}
