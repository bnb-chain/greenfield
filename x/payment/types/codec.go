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
	cdc.RegisterConcrete(&MsgCreatePaymentAccount{}, "payment/CreatePaymentAccount", nil)
	cdc.RegisterConcrete(&MsgDeposit{}, "payment/Deposit", nil)
	cdc.RegisterConcrete(&MsgWithdraw{}, "payment/Withdraw", nil)
	cdc.RegisterConcrete(&MsgSponse{}, "payment/Sponse", nil)
	cdc.RegisterConcrete(&MsgDisableRefund{}, "payment/DisableRefund", nil)
	cdc.RegisterConcrete(&MsgMockCreateBucket{}, "payment/MockCreateBucket", nil)
	// this line is used by starport scaffolding # 2
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgCreatePaymentAccount{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDeposit{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgWithdraw{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgSponse{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgDisableRefund{},
	)
	registry.RegisterImplementations((*sdk.Msg)(nil),
		&MsgMockCreateBucket{},
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
