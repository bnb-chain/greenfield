package payment

import (
	"math/rand"

	"github.com/bnb-chain/bfs/testutil/sample"
	paymentsimulation "github.com/bnb-chain/bfs/x/payment/simulation"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = paymentsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgCreatePaymentAccount = "op_weight_msg_create_payment_account"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreatePaymentAccount int = 100

	opWeightMsgDeposit = "op_weight_msg_deposit"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeposit int = 100

	opWeightMsgWithdraw = "op_weight_msg_withdraw"
	// TODO: Determine the simulation weight value
	defaultWeightMsgWithdraw int = 100

	opWeightMsgSponse = "op_weight_msg_sponse"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSponse int = 100

	opWeightMsgDisableRefund = "op_weight_msg_disable_refund"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDisableRefund int = 100

	opWeightMsgMockCreateBucket = "op_weight_msg_mock_create_bucket"
	// TODO: Determine the simulation weight value
	defaultWeightMsgMockCreateBucket int = 100

	opWeightMsgCreateMockBucketMeta = "op_weight_msg_mock_bucket_meta"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateMockBucketMeta int = 100

	opWeightMsgUpdateMockBucketMeta = "op_weight_msg_mock_bucket_meta"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateMockBucketMeta int = 100

	opWeightMsgDeleteMockBucketMeta = "op_weight_msg_mock_bucket_meta"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteMockBucketMeta int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	paymentGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&paymentGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {

	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgCreatePaymentAccount int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreatePaymentAccount, &weightMsgCreatePaymentAccount, nil,
		func(_ *rand.Rand) {
			weightMsgCreatePaymentAccount = defaultWeightMsgCreatePaymentAccount
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreatePaymentAccount,
		paymentsimulation.SimulateMsgCreatePaymentAccount(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeposit int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) {
			weightMsgDeposit = defaultWeightMsgDeposit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeposit,
		paymentsimulation.SimulateMsgDeposit(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgWithdraw int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgWithdraw, &weightMsgWithdraw, nil,
		func(_ *rand.Rand) {
			weightMsgWithdraw = defaultWeightMsgWithdraw
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgWithdraw,
		paymentsimulation.SimulateMsgWithdraw(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgSponse int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSponse, &weightMsgSponse, nil,
		func(_ *rand.Rand) {
			weightMsgSponse = defaultWeightMsgSponse
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSponse,
		paymentsimulation.SimulateMsgSponse(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDisableRefund int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDisableRefund, &weightMsgDisableRefund, nil,
		func(_ *rand.Rand) {
			weightMsgDisableRefund = defaultWeightMsgDisableRefund
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDisableRefund,
		paymentsimulation.SimulateMsgDisableRefund(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgMockCreateBucket int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgMockCreateBucket, &weightMsgMockCreateBucket, nil,
		func(_ *rand.Rand) {
			weightMsgMockCreateBucket = defaultWeightMsgMockCreateBucket
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMockCreateBucket,
		paymentsimulation.SimulateMsgMockCreateBucket(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
