package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
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
			Operator: simAccount.Address.String(),
		}

		// TODO: Handling the DeleteBucket simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "DeleteBucket simulation not implemented"), nil, nil
	}
}

func SimulateMsgUpdateBucketInfo(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgUpdateBucketInfo{
			Operator: simAccount.Address.String(),
		}

		// TODO: Handling the UpdateBucketInfo simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UpdateBucketInfo simulation not implemented"), nil, nil
	}
}

func SimulateMsgCreateObject(
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

func SimulateMsgCancelCreateObject(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCancelCreateObject{
			Operator: simAccount.Address.String(),
		}

		// TODO: Handling the CancelCreateObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CancelCreateObject simulation not implemented"), nil, nil
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
			Operator: simAccount.Address.String(),
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
			Operator: simAccount.Address.String(),
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
		msg := &types.MsgRejectSealObject{
			Operator: simAccount.Address.String(),
		}

		// TODO: Handling the RejectSealObject simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "RejectSealObject simulation not implemented"), nil, nil
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
			Operator: simAccount.Address.String(),
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
			Operator: simAccount.Address.String(),
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
			Member: simAccount.Address.String(),
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
			Operator: simAccount.Address.String(),
		}

		// TODO: Handling the UpdateGroupMember simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "UpdateGroupMember simulation not implemented"), nil, nil
	}
}
