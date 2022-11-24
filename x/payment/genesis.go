package payment

import (
	"github.com/bnb-chain/bfs/x/payment/keeper"
	"github.com/bnb-chain/bfs/x/payment/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// Set all the streamRecord
	for _, elem := range genState.StreamRecordList {
		k.SetStreamRecord(ctx, elem)
	}
	// Set all the paymentAccountCount
	for _, elem := range genState.PaymentAccountCountList {
		k.SetPaymentAccountCount(ctx, elem)
	}
	// Set all the paymentAccount
	for _, elem := range genState.PaymentAccountList {
		k.SetPaymentAccount(ctx, elem)
	}
	// this line is used by starport scaffolding # genesis/module/init
	k.SetParams(ctx, genState.Params)
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()
	genesis.Params = k.GetParams(ctx)

	genesis.StreamRecordList = k.GetAllStreamRecord(ctx)
	genesis.PaymentAccountCountList = k.GetAllPaymentAccountCount(ctx)
	genesis.PaymentAccountList = k.GetAllPaymentAccount(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
