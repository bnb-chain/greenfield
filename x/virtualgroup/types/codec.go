package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgStorageProviderExit{}, "virtualgroup/StorageProviderExit", nil)
	cdc.RegisterConcrete(&MsgCompleteStorageProviderExit{}, "virtualgroup/CompleteStorageProviderExit", nil)
	cdc.RegisterConcrete(&MsgCompleteSwapOut{}, "virtualgroup/CompleteSwapOut", nil)
	cdc.RegisterConcrete(&MsgCancelSwapOut{}, "virtualgroup/CancelSwapOut", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateGlobalVirtualGroup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeleteGlobalVirtualGroup{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgStorageProviderExit{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCompleteStorageProviderExit{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSwapOut{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeposit{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWithdraw{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSettle{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCompleteSwapOut{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCancelSwapOut{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	Amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
