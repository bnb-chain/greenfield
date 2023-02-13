package client

import (
	_ "encoding/json"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	"github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UpgradeQueryClient = upgradetypes.QueryClient
type DistrQueryClient = distrtypes.QueryClient
type DistrMsgClient = distrtypes.MsgClient
type SlashingQueryClient = slashingtypes.QueryClient
type SlashingMsgClient = slashingtypes.MsgClient
type StakingQueryClient = stakingtypes.QueryClient
type StakingMsgClient = stakingtypes.MsgClient
type AuthQueryClient = authtypes.QueryClient
type BankQueryClient = banktypes.QueryClient
type BankMsgClient = banktypes.MsgClient
type GovQueryClient = v1beta1.QueryClient
type GovMsgClient = v1beta1.MsgClient
type AuthzQueryClient = authztypes.QueryClient
type AuthzMsgClient = authztypes.MsgClient
type FeegrantQueryClient = feegranttypes.QueryClient
type FeegrantMsgClient = feegranttypes.MsgClient
type ParamsQueryClient = paramstypes.QueryClient
type TxClient = tx.ServiceClient

type GreenfieldClient struct {
	TxClient
	UpgradeQueryClient
	DistrQueryClient
	DistrMsgClient
	SlashingQueryClient
	SlashingMsgClient
	StakingQueryClient
	StakingMsgClient
	AuthQueryClient
	BankQueryClient
	BankMsgClient
	GovQueryClient
	GovMsgClient
	AuthzQueryClient
	AuthzMsgClient
	FeegrantQueryClient
	FeegrantMsgClient
	ParamsQueryClient
	keyManager keys.KeyManager
	chainId    string
	codec      *codec.ProtoCodec
}

func grpcConn(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		panic(err)
	}
	return conn
}

func NewGreenfieldClient(grpcAddr, chainId string) GreenfieldClient {
	conn := grpcConn(grpcAddr)
	cdc := types.Cdc()
	return GreenfieldClient{
		tx.NewServiceClient(conn),
		upgradetypes.NewQueryClient(conn),
		distrtypes.NewQueryClient(conn),
		distrtypes.NewMsgClient(conn),
		slashingtypes.NewQueryClient(conn),
		slashingtypes.NewMsgClient(conn),
		stakingtypes.NewQueryClient(conn),
		stakingtypes.NewMsgClient(conn),
		authtypes.NewQueryClient(conn),
		banktypes.NewQueryClient(conn),
		banktypes.NewMsgClient(conn),
		v1beta1.NewQueryClient(conn),
		v1beta1.NewMsgClient(conn),
		authztypes.NewQueryClient(conn),
		authztypes.NewMsgClient(conn),
		feegranttypes.NewQueryClient(conn),
		feegranttypes.NewMsgClient(conn),
		paramstypes.NewQueryClient(conn),
		nil,
		chainId,
		cdc,
	}
}

func NewGreenfieldClientWithKeyManager(grpcAddr, chainId string, keyManager keys.KeyManager) GreenfieldClient {
	gnfdClient := NewGreenfieldClient(grpcAddr, chainId)
	gnfdClient.keyManager = keyManager
	return gnfdClient
}

func (c *GreenfieldClient) GetKeyManager() (keys.KeyManager, error) {
	if c.keyManager == nil {
		return nil, types.KeyManagerNotInitError
	}
	return c.keyManager, nil
}

func (c *GreenfieldClient) SetKeyManager(km keys.KeyManager) {
	c.keyManager = km
}

func (c *GreenfieldClient) SetChainId(id string) {
	c.chainId = id
}

func (c *GreenfieldClient) GetChainId() (string, error) {
	if c.chainId == "" {
		return "", types.ChainIdNotSetError
	}
	return c.chainId, nil
}
