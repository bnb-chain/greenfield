package simulation

import (
	"math/rand"

	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)


func SimulateMsgCreateBucket(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCreateBucket{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the CreateBucket simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CreateBucket simulation not implemented"), nil, nil
	}
}

func SimulateMsgDeleteBucket(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDeleteBucket{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the DeleteBucket simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "DeleteBucket simulation not implemented"), nil, nil
	}
}

func SimulateMsgPutObject(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCreateObject{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the CreateObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CreateObject simulation not implemented"), nil, nil
	}
}

func SimulateMsgDeleteObject(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDeleteObject{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the DeleteObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "DeleteObject simulation not implemented"), nil, nil
	}
}

func SimulateMsgSealObject(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgSealObject{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the SealObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "SealObject simulation not implemented"), nil, nil
	}
}

func SimulateMsgRejectUnsealedObject(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgRejectUnsealedObject{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the RejectUnsealedObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "RejectUnsealedObject simulation not implemented"), nil, nil
	}
}

func SimulateMsgCopyObject(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCopyObject{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the CopyObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CopyObject simulation not implemented"), nil, nil
	}
}


func SimulateMsgCreateGroup(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCreateGroup{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the CreateGroup simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CreateGroup simulation not implemented"), nil, nil
	}
}

func SimulateMsgDeleteGroup(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDeleteGroup{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the DeleteGroup simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "DeleteGroup simulation not implemented"), nil, nil
	}
}

func SimulateMsgLeaveGroup(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgLeaveGroup{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the LeaveGroup simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "LeaveGroup simulation not implemented"), nil, nil
	}
}

func SimulateMsgUpdateGroupMember(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUpdateGroupMember{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the UpdateGroupMember simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UpdateGroupMember simulation not implemented"), nil, nil
	}
}
