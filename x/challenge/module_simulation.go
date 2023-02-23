package challenge

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
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
	_ = sample.AccAddress
	_ = challengesimulation.FindAccount
	_ = simappparams.StakePerAccount
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

	opWeightMsgHeartbeat = "op_weight_msg_heartbeat"
	// TODO: Determine the simulation weight value
	defaultWeightMsgHeartbeat int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	challengeGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&challengeGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(_ *rand.Rand) []simtypes.ParamChange {
	challengeParams := types.DefaultParams()
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyChallengeCountPerBlock), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.ChallengeCountPerBlock))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyChallengeExpirePeriod), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.ChallengeExpirePeriod))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySlashCoolingOffPeriod), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.SlashCoolingOffPeriod))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySlashAmountSizeRate), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.SlashAmountSizeRate))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySlashAmountMin), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.SlashAmountMin))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySlashAmountMax), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.SlashAmountMax))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyRewardValidatorRatio), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.RewardValidatorRatio))
		}),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyRewardChallengerRatio), func(r *rand.Rand) string {
			return string(types.Amino.MustMarshalJSON(challengeParams.RewardChallengerRatio))
		}),
	}
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

	var weightMsgHeartbeat int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgHeartbeat, &weightMsgHeartbeat, nil,
		func(_ *rand.Rand) {
			weightMsgHeartbeat = defaultWeightMsgHeartbeat
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgHeartbeat,
		challengesimulation.SimulateMsgHeartbeat(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
