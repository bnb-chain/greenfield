package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	consensustypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	proposaltypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	challengetypes "github.com/bnb-chain/greenfield/x/challenge/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func Codec() *codec.ProtoCodec {
	interfaceRegistry := types.NewInterfaceRegistry()
	challengetypes.RegisterInterfaces(interfaceRegistry)
	cryptocodec.RegisterInterfaces(interfaceRegistry)
	authtypes.RegisterInterfaces(interfaceRegistry)
	authztypes.RegisterInterfaces(interfaceRegistry)
	banktypes.RegisterInterfaces(interfaceRegistry)
	distrtypes.RegisterInterfaces(interfaceRegistry)
	feegranttypes.RegisterInterfaces(interfaceRegistry)
	proposaltypes.RegisterInterfaces(interfaceRegistry)
	slashingtypes.RegisterInterfaces(interfaceRegistry)
	stakingtypes.RegisterInterfaces(interfaceRegistry)
	bridgetypes.RegisterInterfaces(interfaceRegistry)
	sptypes.RegisterInterfaces(interfaceRegistry)
	paymenttypes.RegisterInterfaces(interfaceRegistry)
	storagetypes.RegisterInterfaces(interfaceRegistry)
	govv1.RegisterInterfaces(interfaceRegistry)
	crosschaintypes.RegisterInterfaces(interfaceRegistry)
	consensustypes.RegisterInterfaces(interfaceRegistry)
	oracletypes.RegisterInterfaces(interfaceRegistry)
	nft.RegisterInterfaces(interfaceRegistry)
	evidencetypes.RegisterInterfaces(interfaceRegistry)
	gashubtypes.RegisterInterfaces(interfaceRegistry)
	minttypes.RegisterInterfaces(interfaceRegistry)

	return codec.NewProtoCodec(interfaceRegistry)
}
