package keeper

import (
	"context"
	"encoding/hex"
	"time"

	"cosmossdk.io/errors"
	errorsmod "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	gnfderrors "github.com/bnb-chain/greenfield/types/errors"
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
	maintenanceAcc, err := sdk.AccAddressFromHexUnsafe(msg.MaintenanceAddress)
	if err != nil {
		return nil, err
	}

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

	if _, found := k.GetStorageProviderByOperatorAddr(ctx, spAcc); found {
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

	// check if the bls pubkey has been registered before
	blsPk, err := hex.DecodeString(msg.BlsKey)
	if err != nil || len(blsPk) != sdk.BLSPubKeyLength {
		return nil, types.ErrStorageProviderInvalidBlsKey
	}
	if _, found := k.GetStorageProviderByBlsKey(ctx, blsPk); found {
		return nil, types.ErrStorageProviderBlsKeyExists
	}
	if err = k.checkBlsProof(blsPk, msg.BlsProof); err != nil {
		return nil, err
	}
	if err = msg.Description.EnsureLength(); err != nil {
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

	sp, err := types.NewStorageProvider(k.GetNextSpID(ctx), spAcc, fundingAcc, sealAcc, approvalAcc, gcAcc, maintenanceAcc,
		msg.Deposit.Amount, msg.Endpoint, msg.Description, msg.BlsKey)
	if err != nil {
		return nil, err
	}

	// external sp default to be in STATUS_IN_MAINTENANCE after the proposal passed
	if ctx.BlockHeight() != 0 {
		sp.Status = types.STATUS_IN_MAINTENANCE
	}

	k.SetStorageProvider(ctx, &sp)
	k.SetStorageProviderByOperatorAddr(ctx, &sp)
	k.SetStorageProviderByApprovalAddr(ctx, &sp)
	k.SetStorageProviderByFundingAddr(ctx, &sp)
	k.SetStorageProviderBySealAddr(ctx, &sp)
	k.SetStorageProviderByGcAddr(ctx, &sp)
	k.SetStorageProviderByBlsKey(ctx, &sp)

	if err = ctx.EventManager().EmitTypedEvents(&types.EventCreateStorageProvider{
		SpId:               sp.Id,
		SpAddress:          spAcc.String(),
		FundingAddress:     fundingAcc.String(),
		SealAddress:        sealAcc.String(),
		ApprovalAddress:    approvalAcc.String(),
		GcAddress:          gcAcc.String(),
		MaintenanceAddress: maintenanceAcc.String(),
		Endpoint:           msg.Endpoint,
		TotalDeposit:       &msg.Deposit,
		Status:             sp.Status,
		Description:        sp.Description,
		BlsKey:             hex.EncodeToString(sp.BlsKey),
	}); err != nil {
		return nil, err
	}

	// set initial sp storage price
	spStoragePrice := types.SpStoragePrice{
		SpId:          sp.Id,
		UpdateTimeSec: ctx.BlockTime().Unix(),
		ReadPrice:     msg.ReadPrice,
		StorePrice:    msg.StorePrice,
		FreeReadQuota: msg.FreeReadQuota,
	}
	k.SetSpStoragePrice(ctx, spStoragePrice)

	return &types.MsgCreateStorageProviderResponse{}, nil
}

// EditStorageProvider defines a method for editing a existing storage provider
func (k msgServer) EditStorageProvider(goCtx context.Context, msg *types.MsgEditStorageProvider) (*types.MsgEditStorageProviderResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.SpAddress)

	sp, found := k.GetStorageProviderByOperatorAddr(ctx, operatorAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	changed := false

	// replace endpoint
	if len(msg.Endpoint) != 0 {
		sp.Endpoint = msg.Endpoint
		changed = true
	}

	if msg.Description != nil {
		description, err := sp.Description.UpdateDescription(msg.Description)
		if err != nil {
			return nil, err
		}
		sp.Description = *description
		changed = true
	}

	if msg.SealAddress != "" {
		sealAcc := sdk.MustAccAddressFromHex(msg.SealAddress)
		sp.SealAddress = sealAcc.String()
		changed = true
	}

	if msg.ApprovalAddress != "" {
		approvalAcc := sdk.MustAccAddressFromHex(msg.ApprovalAddress)
		sp.ApprovalAddress = approvalAcc.String()
		changed = true
	}

	if msg.GcAddress != "" {
		gcAcc := sdk.MustAccAddressFromHex(msg.GcAddress)
		sp.GcAddress = gcAcc.String()
		changed = true
	}
	if msg.MaintenanceAddress != "" {
		testAcc := sdk.MustAccAddressFromHex(msg.MaintenanceAddress)
		sp.MaintenanceAddress = testAcc.String()
		changed = true
	}
	if msg.BlsKey != "" && len(msg.BlsProof) != 0 {
		blsPk, err := hex.DecodeString(msg.BlsKey)
		if err != nil || len(blsPk) != sdk.BLSPubKeyLength {
			return nil, types.ErrStorageProviderInvalidBlsKey
		}
		if err = k.checkBlsProof(blsPk, msg.BlsProof); err != nil {
			return nil, err
		}
		sp.BlsKey = blsPk
		changed = true
	}

	if !changed {
		return nil, types.ErrStorageProviderNotChanged
	}

	k.SetStorageProvider(ctx, sp)
	k.SetStorageProviderByFundingAddr(ctx, sp)
	k.SetStorageProviderBySealAddr(ctx, sp)
	k.SetStorageProviderByApprovalAddr(ctx, sp)
	k.SetStorageProviderByGcAddr(ctx, sp)
	k.SetStorageProviderByBlsKey(ctx, sp)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventEditStorageProvider{
		SpId:               sp.Id,
		SpAddress:          operatorAcc.String(),
		Endpoint:           sp.Endpoint,
		Description:        sp.Description,
		ApprovalAddress:    sp.ApprovalAddress,
		SealAddress:        sp.SealAddress,
		GcAddress:          sp.GcAddress,
		MaintenanceAddress: sp.MaintenanceAddress,
		BlsKey:             hex.EncodeToString(sp.BlsKey),
	}); err != nil {
		return nil, err
	}

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

	if !sdk.MustAccAddressFromHex(sp.OperatorAddress).Equals(sdk.MustAccAddressFromHex(msg.SpAddress)) {
		return nil, types.ErrDepositAccountNotAllowed.Wrap("the sp address mismatch")
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

	sp, found := k.GetStorageProviderByOperatorAddr(ctx, spAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	if sp.Status != types.STATUS_IN_SERVICE {
		return nil, types.ErrStorageProviderNotInService
	}

	params := k.GetParams(ctx)
	if params.UpdateGlobalPriceInterval == 0 { // update price by month
		blockTime := ctx.BlockTime().UTC()
		days := params.UpdatePriceDisallowedDays
		if IsLastDaysOfTheMonth(blockTime, int(days)) {
			return nil, errors.Wrapf(types.ErrStorageProviderPriceUpdateNotAllow, "price cannot be updated in the last %d days of the month", days)
		}
	}

	current := ctx.BlockTime().Unix()
	spStorePrice := types.SpStoragePrice{
		UpdateTimeSec: current,
		SpId:          sp.Id,
		ReadPrice:     msg.ReadPrice,
		StorePrice:    msg.StorePrice,
		FreeReadQuota: msg.FreeReadQuota,
	}
	k.SetSpStoragePrice(ctx, spStorePrice)

	return &types.MsgUpdateSpStoragePriceResponse{}, nil
}

func IsLastDaysOfTheMonth(now time.Time, days int) bool {
	now = now.UTC()
	year, month, _ := now.Date()
	nextMonth := time.Date(year, month+1, 1, 0, 0, 0, 0, time.FixedZone("UTC", 0))
	daysBack := nextMonth.AddDate(0, 0, -1*days)
	return now.After(daysBack)
}

func (k msgServer) UpdateParams(goCtx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	if k.GetAuthority() != req.Authority {
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	if ctx.IsUpgraded(upgradetypes.Nagqu) {
		params := k.GetParams(ctx)
		_ = ctx.EventManager().EmitTypedEvents(&params)
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

// checkBlsProof checks the BLS signature of the Storage Provider
func (k msgServer) checkBlsProof(blsPk []byte, sig string) error {
	// check to see if the bls proof is signed from sp
	blsProof, err := hex.DecodeString(sig)
	if err != nil {
		return gnfderrors.ErrInvalidBlsSignature
	}
	if len(blsProof) != sdk.BLSSignatureLength {
		return gnfderrors.ErrInvalidBlsSignature
	}
	signature, err := bls.SignatureFromBytes(blsProof)
	if err != nil {
		return gnfderrors.ErrInvalidBlsSignature
	}
	blsPubKey, err := bls.PublicKeyFromBytes(blsPk)
	if err != nil {
		return types.ErrStorageProviderInvalidBlsKey
	}
	if !signature.Verify(blsPubKey, tmhash.Sum(blsPk)) {
		return sdkerrors.ErrorInvalidSigner.Wrapf("check bls proof failed.")
	}
	return nil
}

// UpdateSpStatus only allow SP to update status between STATUS_MAINTENANCE and STATUS_IN_SERVICE for now.
func (k msgServer) UpdateSpStatus(goCtx context.Context, msg *types.MsgUpdateStorageProviderStatus) (*types.MsgUpdateStorageProviderStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	operatorAcc := sdk.MustAccAddressFromHex(msg.SpAddress)

	sp, found := k.GetStorageProviderByOperatorAddr(ctx, operatorAcc)
	if !found {
		return nil, types.ErrStorageProviderNotFound
	}

	curStatus := sp.Status
	newStatus := msg.GetStatus()

	if curStatus == newStatus {
		return nil, types.ErrStorageProviderNotChanged
	}

	switch curStatus {
	case types.STATUS_IN_SERVICE:
		if newStatus != types.STATUS_IN_MAINTENANCE {
			return nil, types.ErrStorageProviderStatusUpdateNotAllow
		}
		err := k.UpdateToInMaintenance(ctx, sp, msg.GetDuration())
		if err != nil {
			return nil, err
		}
	case types.STATUS_IN_MAINTENANCE:
		if newStatus != types.STATUS_IN_SERVICE {
			return nil, types.ErrStorageProviderStatusUpdateNotAllow
		}
		k.UpdateToInService(ctx, sp)
	case types.STATUS_IN_JAILED, types.STATUS_GRACEFUL_EXITING, types.STATUS_FORCED_EXITING:
		return nil, types.ErrStorageProviderStatusUpdateNotAllow
	}
	k.SetStorageProvider(ctx, sp)

	if err := ctx.EventManager().EmitTypedEvents(&types.EventUpdateStorageProviderStatus{
		SpId:      sp.Id,
		SpAddress: operatorAcc.String(),
		PreStatus: curStatus.String(),
		NewStatus: newStatus.String(),
	}); err != nil {
		return nil, err
	}
	return &types.MsgUpdateStorageProviderStatusResponse{}, nil
}
