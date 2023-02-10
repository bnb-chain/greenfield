package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzcodec "github.com/cosmos/cosmos-sdk/x/authz/codec"
)

func RegisterCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgCreateStorageProvider{}, "sp/CreateStorageProvider", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "sp/Deposit", nil)
	cdc.RegisterConcrete(&MsgEditStorageProvider{}, "sp/EditStorageProvider", nil)
	// this line is used by starport scaffolding # 2
	cdc.RegisterConcrete(&DepositAuthorization{}, "sp/DepositAuthorization", nil)
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
	registry.RegisterImplementations(
		(*authz.Authorization)(nil),
		&DepositAuthorization{},
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
