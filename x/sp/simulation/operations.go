package simulation

import (
	"math/rand"

	"github.com/bnb-chain/bfs/x/sp/keeper"
	"github.com/bnb-chain/bfs/x/sp/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// SimulateMsgCreateStorageProvider generates a MsgCreateStorageProvider with random values
func SimulateMsgCreateStorageProvider(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCreateStorageProvider{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the CreateStorageProvider simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CreateStorageProvider simulation not implemented"), nil, nil
	}
}

// SimulateMsgEditStorageProvider generates a MsgEditStorageProvider with random values
func SimulateMsgEditStorageProvider(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgEditStorageProvider{
			SpAddress: simAccount.Address.String(),
		}

		// TODO: Handling the EditStorageProvider simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "EditStorageProvider simulation not implemented"), nil, nil
	}
}

// SimulateMsgDeposit generates a MsgStaking with random values
func SimulateMsgDeposit(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDeposit{
			Creator: simAccount.Address.String(),
		}

		// TODO: Handling the Deposit simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "Deposit simulation not implemented"), nil, nil
	}
}
