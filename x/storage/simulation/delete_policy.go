package simulation

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	types2 "github.com/bnb-chain/greenfield/types"
	permtypes "github.com/bnb-chain/greenfield/x/permission/types"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

func SimulateMsgDeletePolicy(
	ak types.AccountKeeper,
	bk types.BankKeeper,
	k keeper.Keeper,
) simtypes.Operation {
	return func(r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		simAccount, _ := simtypes.RandomAcc(r, accs)
		msg := &types.MsgDeletePolicy{
			Operator:  simAccount.Address.String(),
			Principal: permtypes.NewPrincipalWithAccount(simAccount.Address),
			Resource:  types2.NewBucketGRN("test-bucket").String(),
		}

		// TODO: Handling the DeletePolicy simulation

		return simtypes.NoOpMsg(types.ModuleName, msg.Type(), "DeletePolicy simulation not implemented"), nil, nil
	}
}
