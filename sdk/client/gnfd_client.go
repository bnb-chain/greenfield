package client

import (
	_ "encoding/json"

	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"google.golang.org/grpc"
)

type AuthQueryClient = authtypes.QueryClient
type AuthzQueryClient = authztypes.QueryClient
type BankQueryClient = banktypes.QueryClient
type CrosschainQueryClient = crosschaintypes.QueryClient
type DistrQueryClient = distrtypes.QueryClient
type FeegrantQueryClient = feegranttypes.QueryClient
type GashubQueryClient = gashubtypes.QueryClient
type PaymentQueryClient = paymenttypes.QueryClient
type SpQueryClient = sptypes.QueryClient
type BridgeQueryClient = bridgetypes.QueryClient
type StorageQueryClient = storagetypes.QueryClient
type GovQueryClientV1 = govv1.QueryClient
type OracleQueryClient = oracletypes.QueryClient
type ParamsQueryClient = paramstypes.QueryClient
type SlashingQueryClient = slashingtypes.QueryClient
type StakingQueryClient = stakingtypes.QueryClient
type TxClient = tx.ServiceClient
type UpgradeQueryClient = upgradetypes.QueryClient
type GreenfieldClient struct {
	AuthQueryClient
	AuthzQueryClient
	BankQueryClient
	CrosschainQueryClient
	DistrQueryClient
	FeegrantQueryClient
	GashubQueryClient
	PaymentQueryClient
	SpQueryClient
	BridgeQueryClient
	StorageQueryClient
	GovQueryClientV1
	OracleQueryClient
	ParamsQueryClient
	SlashingQueryClient
	StakingQueryClient
	TxClient
	UpgradeQueryClient
	keyManager keys.KeyManager
	chainId    string
	codec      *codec.ProtoCodec

	// option field
	grpcDialOption []grpc.DialOption
}

func grpcConn(addr string, opts ...grpc.DialOption) *grpc.ClientConn {
	conn, err := grpc.Dial(
		addr,
		opts...,
	)
	if err != nil {
		panic(err)
	}
	return conn
}

func NewGreenfieldClient(grpcAddr, chainId string, opts ...GreenfieldClientOption) *GreenfieldClient {
	client := &GreenfieldClient{
		chainId: chainId,
		codec:   types.Cdc(),
	}
	for _, opt := range opts {
		opt.Apply(client)
	}

	conn := grpcConn(grpcAddr, client.grpcDialOption...)
	client.AuthQueryClient = authtypes.NewQueryClient(conn)
	client.AuthzQueryClient = authztypes.NewQueryClient(conn)
	client.BankQueryClient = banktypes.NewQueryClient(conn)
	client.CrosschainQueryClient = crosschaintypes.NewQueryClient(conn)
	client.DistrQueryClient = distrtypes.NewQueryClient(conn)
	client.FeegrantQueryClient = feegranttypes.NewQueryClient(conn)
	client.GashubQueryClient = gashubtypes.NewQueryClient(conn)
	client.PaymentQueryClient = paymenttypes.NewQueryClient(conn)
	client.SpQueryClient = sptypes.NewQueryClient(conn)
	client.BridgeQueryClient = bridgetypes.NewQueryClient(conn)
	client.StorageQueryClient = storagetypes.NewQueryClient(conn)
	client.GovQueryClientV1 = govv1.NewQueryClient(conn)
	client.OracleQueryClient = oracletypes.NewQueryClient(conn)
	client.ParamsQueryClient = paramstypes.NewQueryClient(conn)
	client.SlashingQueryClient = slashingtypes.NewQueryClient(conn)
	client.StakingQueryClient = stakingtypes.NewQueryClient(conn)
	client.UpgradeQueryClient = upgradetypes.NewQueryClient(conn)
	client.TxClient = tx.NewServiceClient(conn)
	return client
}

func (c *GreenfieldClient) SetKeyManager(keyManager keys.KeyManager) {
	c.keyManager = keyManager
}

func (c *GreenfieldClient) GetKeyManager() (keys.KeyManager, error) {
	if c.keyManager == nil {
		return nil, types.KeyManagerNotInitError
	}
	return c.keyManager, nil
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
