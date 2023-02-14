package storage

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/bnb-chain/greenfield/testutil/sample"
	storagesimulation "github.com/bnb-chain/greenfield/x/storage/simulation"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

// avoid unused import issue
var (
	_ = sample.AccAddress
	_ = storagesimulation.FindAccount
	_ = simappparams.StakePerAccount
	_ = simulation.MsgEntryKind
	_ = baseapp.Paramspace
)

const (
	opWeightMsgCreateBucket = "op_weight_msg_create_bucket"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateBucket int = 100

	opWeightMsgDeleteBucket = "op_weight_msg_delete_bucket"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteBucket int = 100

	opWeightMsgPutObject = "op_weight_msg_put_object"
	// TODO: Determine the simulation weight value
	defaultWeightMsgPutObject int = 100

	opWeightMsgSealObject = "op_weight_msg_seal_object"
	// TODO: Determine the simulation weight value
	defaultWeightMsgSealObject int = 100

	opWeightMsgRejectUnsealedObject = "op_weight_msg_reject_unsealed_object"
	// TODO: Determine the simulation weight value
	defaultWeightMsgRejectUnsealedObject int = 100

	opWeightMsgDeleteObject = "op_weight_msg_delete_object"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteObject int = 100

	opWeightMsgCreateGroup = "op_weight_msg_create_group"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCreateGroup int = 100

	opWeightMsgDeleteGroup = "op_weight_msg_delete_group"
	// TODO: Determine the simulation weight value
	defaultWeightMsgDeleteGroup int = 100

	opWeightMsgUpdateGroupMember = "op_weight_msg_update_group_member"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateGroupMember int = 100

	opWeightMsgLeaveGroup = "op_weight_msg_leave_group"
	// TODO: Determine the simulation weight value
	defaultWeightMsgLeaveGroup int = 100

	opWeightMsgCopyObject = "op_weight_msg_copy_object"
	// TODO: Determine the simulation weight value
	defaultWeightMsgCopyObject int = 100

	opWeightMsgUpdateBucketReadQuota = "op_weight_msg_update_bucket_read_quota"
	// TODO: Determine the simulation weight value
	defaultWeightMsgUpdateBucketReadQuota int = 100

	// this line is used by starport scaffolding # simapp/module/const
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	greenfieldGenesis := types.GenesisState{
		Params: types.DefaultParams(),
		// this line is used by starport scaffolding # simapp/module/genesisState
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&greenfieldGenesis)
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

	var weightMsgCreateBucket int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateBucket, &weightMsgCreateBucket, nil,
		func(_ *rand.Rand) {
			weightMsgCreateBucket = defaultWeightMsgCreateBucket
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateBucket,
		storagesimulation.SimulateMsgCreateBucket(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteBucket int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteBucket, &weightMsgDeleteBucket, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteBucket = defaultWeightMsgDeleteBucket
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteBucket,
		storagesimulation.SimulateMsgDeleteBucket(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgPutObject int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgPutObject, &weightMsgPutObject, nil,
		func(_ *rand.Rand) {
			weightMsgPutObject = defaultWeightMsgPutObject
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgPutObject,
		storagesimulation.SimulateMsgPutObject(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgSealObject int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgSealObject, &weightMsgSealObject, nil,
		func(_ *rand.Rand) {
			weightMsgSealObject = defaultWeightMsgSealObject
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgSealObject,
		storagesimulation.SimulateMsgSealObject(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgRejectUnsealedObject int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgRejectUnsealedObject, &weightMsgRejectUnsealedObject, nil,
		func(_ *rand.Rand) {
			weightMsgRejectUnsealedObject = defaultWeightMsgRejectUnsealedObject
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgRejectUnsealedObject,
		storagesimulation.SimulateMsgRejectUnsealedObject(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteObject int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteObject, &weightMsgDeleteObject, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteObject = defaultWeightMsgDeleteObject
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteObject,
		storagesimulation.SimulateMsgDeleteObject(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgCreateGroup int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCreateGroup, &weightMsgCreateGroup, nil,
		func(_ *rand.Rand) {
			weightMsgCreateGroup = defaultWeightMsgCreateGroup
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCreateGroup,
		storagesimulation.SimulateMsgCreateGroup(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgDeleteGroup int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgDeleteGroup, &weightMsgDeleteGroup, nil,
		func(_ *rand.Rand) {
			weightMsgDeleteGroup = defaultWeightMsgDeleteGroup
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDeleteGroup,
		storagesimulation.SimulateMsgDeleteGroup(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateGroupMember int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateGroupMember, &weightMsgUpdateGroupMember, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateGroupMember = defaultWeightMsgUpdateGroupMember
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateGroupMember,
		storagesimulation.SimulateMsgUpdateGroupMember(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgLeaveGroup int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgLeaveGroup, &weightMsgLeaveGroup, nil,
		func(_ *rand.Rand) {
			weightMsgLeaveGroup = defaultWeightMsgLeaveGroup
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgLeaveGroup,
		storagesimulation.SimulateMsgLeaveGroup(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgCopyObject int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgCopyObject, &weightMsgCopyObject, nil,
		func(_ *rand.Rand) {
			weightMsgCopyObject = defaultWeightMsgCopyObject
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgCopyObject,
		storagesimulation.SimulateMsgCopyObject(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	var weightMsgUpdateBucketReadQuota int
	simState.AppParams.GetOrGenerate(simState.Cdc, opWeightMsgUpdateBucketReadQuota, &weightMsgUpdateBucketReadQuota, nil,
		func(_ *rand.Rand) {
			weightMsgUpdateBucketReadQuota = defaultWeightMsgUpdateBucketReadQuota
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdateBucketReadQuota,
		storagesimulation.SimulateMsgUpdateReadQuota(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	// this line is used by starport scaffolding # simapp/module/operation

	return operations
}
