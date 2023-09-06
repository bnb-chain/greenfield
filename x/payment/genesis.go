package payment

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/payment/keeper"
	v1 "github.com/bnb-chain/greenfield/x/payment/types/v1"
)

// InitGenesis initializes the module's state from a provided genesis state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState v1.GenesisState) {
	// Set all the streamRecord
	for _, elem := range genState.StreamRecordList {
		k.SetStreamRecord(ctx, &elem)
	}
	// Set all the paymentAccountCount
	for _, elem := range genState.PaymentAccountCountList {
		k.SetPaymentAccountCount(ctx, &elem)
	}
	// Set all the paymentAccount
	for _, elem := range genState.PaymentAccountList {
		k.SetPaymentAccount(ctx, &elem)
	}
	// Set all the autoSettleRecord
	for _, elem := range genState.AutoSettleRecordList {
		k.SetAutoSettleRecord(ctx, &elem)
	}
	err := k.SetV1Params(ctx, genState.Params)
	if err != nil {
		panic(err)
	}
}

// ExportGenesis returns the module's exported genesis
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *v1.GenesisState {
	genesis := v1.DefaultGenesis()
	genesis.Params = k.GetV1Params(ctx)

	genesis.StreamRecordList = k.GetAllStreamRecord(ctx)
	genesis.PaymentAccountCountList = k.GetAllPaymentAccountCount(ctx)
	genesis.PaymentAccountList = k.GetAllPaymentAccount(ctx)
	genesis.AutoSettleRecordList = k.GetAllAutoSettleRecord(ctx)
	// this line is used by starport scaffolding # genesis/module/export

	return genesis
}
