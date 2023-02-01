package sp

import (
	"math/rand"

	"github.com/bnb-chain/greenfield/testutil/sample"
	spsimulation "github.com/bnb-chain/greenfield/x/sp/simulation"
	"github.com/bnb-chain/greenfield/x/sp/types"
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
	_ = spsimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgCreateStorageProvider = "op_weight_msg_create_storage_provider"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateStorageProvider int = 100

	opWeightMsgDeposit = "op_weight_msg_deposit"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeposit int = 100

	opWeightMsgEditStorageProvider = "op_weight_msg_edit_storage_provider"
	// TODO: Determine the simulation weight value
	defaultWeightMsgEditStorageProvider int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	spGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&spGenesis)
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

	var weightMsgCreateStorageProvider int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateStorageProvider, &weightMsgCreateStorageProvider, nil,
		func(_ *rand.Rand) {
			weightMsgCreateStorageProvider = defaultWeightMsgCreateStorageProvider
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateStorageProvider,
		spsimulation.SimulateMsgCreateStorageProvider(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeposit int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeposit, &weightMsgDeposit, nil,
		func(_ *rand.Rand) {
			weightMsgDeposit = defaultWeightMsgDeposit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeposit,
		spsimulation.SimulateMsgDeposit(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgEditStorageProvider int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgEditStorageProvider, &weightMsgEditStorageProvider, nil,
		func(_ *rand.Rand) {
			weightMsgEditStorageProvider = defaultWeightMsgEditStorageProvider
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgEditStorageProvider,
		spsimulation.SimulateMsgEditStorageProvider(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
