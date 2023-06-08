package keeper

import (
	"context"

	errorsmod "cosmossdk.io/errors"
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
		return nil, errorsmod.Wrapf(govtypes.ErrInvalidSigner, "invalid authority; expected %s, got %s", k.GetAuthority(), req.Authority)
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := k.SetParams(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}

func (k msgServer) CreateGlobalVirtualGroup(goCtx context.Context, req *types.MsgCreateGlobalVirtualGroup) (*types.MsgCreateGlobalVirtualGroupResponse, error) {
	return &types.MsgCreateGlobalVirtualGroupResponse{}, nil
}

func (k msgServer) DeleteGlobalVirtualGroup(goCtx context.Context, req *types.MsgDeleteGlobalVirtualGroup) (*types.MsgDeleteGlobalVirtualGroupResponse, error) {
	return &types.MsgDeleteGlobalVirtualGroupResponse{}, nil
}

func (k msgServer) Deposit(goCtx context.Context, req *types.MsgDeposit) (*types.MsgDepositResponse, error) {
	return &types.MsgDepositResponse{}, nil
}

func (k msgServer) Withdraw(goCtx context.Context, req *types.MsgWithdraw) (*types.MsgWithdrawResponse, error) {
	return &types.MsgWithdrawResponse{}, nil
}

func (k msgServer) SwapOut(goCtx context.Context, req *types.MsgSwapOut) (*types.MsgSwapOutResponse, error) {
	return &types.MsgSwapOutResponse{}, nil
}
