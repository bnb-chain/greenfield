package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateStorageProvider{}, "sp/CreateStorageProvider", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "sp/Deposit", nil)
	cdc.RegisterConcrete(&MsgEditStorageProvider{}, "sp/EditStorageProvider", nil)
	cdc.RegisterConcrete(&MsgUpdateSpStoragePrice{}, "sp/UpdateSpStoragePrice", nil)
	cdc.RegisterConcrete(&DepositAuthorization{}, "sp/DepositAuthorization", nil)
	cdc.RegisterConcrete(&MsgUpdateStorageProviderStatus{}, "sp/UpdateSpStatus", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreateStorageProvider{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeposit{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgEditStorageProvider{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateSpStoragePrice{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateStorageProviderStatus{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgUpdateParams{},
	)

	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&DepositAuthorization{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}

var (
	ModuleCdc = codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
)
