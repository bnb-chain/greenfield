package simulation

import (
	"math/rand"

	"github.com/bnb-chain/bfs/x/bridge/keeper"
	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

func SimulateMsgTransferOut(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		//simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgTransferOut{}

		// TODO: Handling the TransferOut simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "TransferOut simulation not implemented"), nil, nil
	}
}
