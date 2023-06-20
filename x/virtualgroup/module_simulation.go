package virtualgroup

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/bnb-chain/greenfield/testutil/sample"
	virtualgroupsimulation "github.com/bnb-chain/greenfield/x/virtualgroup/simulation"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = virtualgroupsimulation.FindAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
	_ = rand.Rand{}
)

const (
	opWeightMsgStorageProviderExit = "op_weight_msg_storage_provider_exit"
	// TODO: Determine the simulation weight value
	defaultWeightMsgStorageProviderExit int = 100

	opWeightMsgCompleteStorageProviderExit = "op_weight_msg_complete_storage_provider_exit"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCompleteStorageProviderExit int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	virtualgroupGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&virtualgroupGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// ProposalContents doesn't return any content functions for governance proposals.
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgStorageProviderExit int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgStorageProviderExit, &weightMsgStorageProviderExit, nil,
		func(_ *rand.Rand) {
			weightMsgStorageProviderExit = defaultWeightMsgStorageProviderExit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgStorageProviderExit,
		virtualgroupsimulation.SimulateMsgStorageProviderExit(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgCompleteStorageProviderExit int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCompleteStorageProviderExit, &weightMsgCompleteStorageProviderExit, nil,
		func(_ *rand.Rand) {
			weightMsgCompleteStorageProviderExit = defaultWeightMsgCompleteStorageProviderExit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCompleteStorageProviderExit,
		virtualgroupsimulation.SimulateMsgCompleteStorageProviderExit(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{
		simulation.NewWeightedProposalMsg(
			opWeightMsgStorageProviderExit,
			defaultWeightMsgStorageProviderExit,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				virtualgroupsimulation.SimulateMsgStorageProviderExit(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		simulation.NewWeightedProposalMsg(
			opWeightMsgCompleteStorageProviderExit,
			defaultWeightMsgCompleteStorageProviderExit,
			func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
				virtualgroupsimulation.SimulateMsgCompleteStorageProviderExit(am.accountKeeper, am.bankKeeper, am.keeper)
				return nil
			},
		),
		// this line is used by starport scaffolding # simapp/module/OpMsg
	}
}
