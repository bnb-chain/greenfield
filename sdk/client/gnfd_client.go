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
	"google.golang.org/grpc/credentials/insecure"
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
		authtypes.NewQueryClient(conn),
		authztypes.NewQueryClient(conn),
		banktypes.NewQueryClient(conn),
		crosschaintypes.NewQueryClient(conn),
		distrtypes.NewQueryClient(conn),
		feegranttypes.NewQueryClient(conn),
		gashubtypes.NewQueryClient(conn),
		paymenttypes.NewQueryClient(conn),
		sptypes.NewQueryClient(conn),
		bridgetypes.NewQueryClient(conn),
		storagetypes.NewQueryClient(conn),
		govv1.NewQueryClient(conn),
		oracletypes.NewQueryClient(conn),
		paramstypes.NewQueryClient(conn),
		slashingtypes.NewQueryClient(conn),
		stakingtypes.NewQueryClient(conn),
		tx.NewServiceClient(conn),
		upgradetypes.NewQueryClient(conn),
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

func (c *GreenfieldClient) SetKeyManager(keyManager keys.KeyManager) {
	c.keyManager = keyManager
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
