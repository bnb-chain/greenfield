package challenge

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/bnb-chain/greenfield/testutil/sample"
	challengesimulation "github.com/bnb-chain/greenfield/x/challenge/simulation"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// avoid unused import issue
var (
	_ = sample.RandAccAddressHex
	_ = challengesimulation.FindAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgSubmit = "op_weight_msg_submit"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSubmit int = 100

	opWeightMsgAttest = "op_weight_msg_attest"
	// TODO: Determine the simulation weight value
	defaultWeightMsgAttest int = 100
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	challengeGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&challengeGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalMsg {
	return nil
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)

	var weightMsgSubmit int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSubmit, &weightMsgSubmit, nil,
		func(_ *rand.Rand) {
			weightMsgSubmit = defaultWeightMsgSubmit
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSubmit,
		challengesimulation.SimulateMsgSubmit(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgAttest int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgAttest, &weightMsgAttest, nil,
		func(_ *rand.Rand) {
			weightMsgAttest = defaultWeightMsgAttest
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgAttest,
		challengesimulation.SimulateMsgAttest(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	return operations
}
