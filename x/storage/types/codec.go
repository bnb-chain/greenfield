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
	cdc.RegisterConcrete(&MsgLeaveGroup{}, "storage/LeaveGroup", nil)
	cdc.RegisterConcrete(&MsgCopyObject{}, "storage/CopyObject", nil)
	cdc.RegisterConcrete(&MsgUpdateBucketInfo{}, "storage/UpdateBucketInfo", nil)
	cdc.RegisterConcrete(&MsgCancelCreateObject{}, "storage/CancelCreateObject", nil)
	cdc.RegisterConcrete(&MsgDeletePolicy{}, "storage/DeletePolicy", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateBucket{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteBucket{},
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
		&MsgDeleteObject{},
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
		&MsgLeaveGroup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCopyObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateBucketInfo{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelCreateObject{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeletePolicy{},
	)
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
