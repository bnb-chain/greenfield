package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateBucket{}, "storage/CreateBucket", nil)
	cdc.RegisterConcrete(&MsgDeleteBucket{}, "storage/DeleteBucket", nil)
	cdc.RegisterConcrete(&MsgCreateObject{}, "storage/CreateObject", nil)
	cdc.RegisterConcrete(&MsgSealObject{}, "storage/SealObject", nil)
	cdc.RegisterConcrete(&MsgRejectSealObject{}, "storage/RejectSealObject", nil)
	cdc.RegisterConcrete(&MsgDeleteObject{}, "storage/DeleteObject", nil)
	cdc.RegisterConcrete(&MsgCreateGroup{}, "storage/CreateGroup", nil)
	cdc.RegisterConcrete(&MsgDeleteGroup{}, "storage/DeleteGroup", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupMember{}, "storage/UpdateGroupMember", nil)
	cdc.RegisterConcrete(&MsgUpdateGroupExtra{}, "storage/UpdateGroupExtra", nil)
	cdc.RegisterConcrete(&MsgLeaveGroup{}, "storage/LeaveGroup", nil)
	cdc.RegisterConcrete(&MsgCopyObject{}, "storage/CopyObject", nil)
	cdc.RegisterConcrete(&MsgUpdateBucketInfo{}, "storage/UpdateBucketInfo", nil)
	cdc.RegisterConcrete(&MsgCancelCreateObject{}, "storage/CancelCreateObject", nil)
	cdc.RegisterConcrete(&MsgDeletePolicy{}, "storage/DeletePolicy", nil)
	cdc.RegisterConcrete(&MsgMigrateBucket{}, "storage/MigrateBucket", nil)
	cdc.RegisterConcrete(&MsgCompleteMigrateBucket{}, "storage/CompleteMigrateBucket", nil)
	cdc.RegisterConcrete(&MsgCancelMigrateBucket{}, "storage/CancelMigrateBucket", nil)
	cdc.RegisterConcrete(&MsgRejectMigrateBucket{}, "storage/RejectMigrateBucket", nil)
	cdc.RegisterConcrete(&MsgSetBucketFlowRateLimit{}, "storage/SetBucketFlowRateLimit", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateBucketInfo{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMirrorBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDiscontinueBucket{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSealObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRejectSealObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCopyObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelCreateObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMirrorObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDiscontinueObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateObjectInfo{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGroup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteGroup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateGroupMember{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateGroupExtra{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgLeaveGroup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMirrorGroup{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgPutPolicy{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeletePolicy{},
	)

	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMigrateBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCompleteMigrateBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelMigrateBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgRejectMigrateBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateObjectContent{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelUpdateObjectContent{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDelegateCreateObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgToggleSPAsDelegatedAgent{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSealObjectV2{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDelegateUpdateObjectContent{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSetBucketFlowRateLimit{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
