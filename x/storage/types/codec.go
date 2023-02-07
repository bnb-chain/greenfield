package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
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
	// this line is used by starport scaffolding # 3

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(Amino)
)

func init() {
	RegisterCodec(Amino)
	cryptocodec.RegisterCrypto(Amino)
	sdk.RegisterLegacyAminoCodec(Amino)

	// Register all Amino interfaces and concrete types on the authz Amino codec so that this can later be
	// used to properly serialize MsgGrant and MsgExec instances
	RegisterCodec(authzcodec.Amino)
}
