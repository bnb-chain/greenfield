package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	"github.com/bnb-chain/greenfield/x/virtualgroup/keeper"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func SimulateMsgCompleteStorageProviderExit(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgCompleteStorageProviderExit{
			StorageProvider: simAccount.Address.String(),
		}

		// TODO: Handling the CompleteStorageProviderExit simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "CompleteStorageProviderExit simulation not implemented"), nil, nil
	}
}
