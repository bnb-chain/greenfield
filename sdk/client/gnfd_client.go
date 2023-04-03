package client

import (
	_ "encoding/json"

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

	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	challengetypes "github.com/bnb-chain/greenfield/x/challenge/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

// AuthQueryClient is a type to define the auth types Query Client
type AuthQueryClient = authtypes.QueryClient

// AuthzQueryClient is a type to define the authz types Query Client
type AuthzQueryClient = authztypes.QueryClient

// BankQueryClient is a type to define the bank types Query Client
type BankQueryClient = banktypes.QueryClient

// ChallengeQueryClient is a type to define the challenge types Query Client
type ChallengeQueryClient = challengetypes.QueryClient

// CrosschainQueryClient is a type to define the crosschain types Query Client
type CrosschainQueryClient = crosschaintypes.QueryClient

// DistrQueryClient is a type to define the distribution types Query Client
type DistrQueryClient = distrtypes.QueryClient

// FeegrantQueryClient is a type to define the feegrant types Query Client
type FeegrantQueryClient = feegranttypes.QueryClient

// GashubQueryClient is a type to define the gashub types Query Client
type GashubQueryClient = gashubtypes.QueryClient

// PaymentQueryClient is a type to define the payment types Query Client
type PaymentQueryClient = paymenttypes.QueryClient

// SpQueryClient is a type to define the sp types Query Client
type SpQueryClient = sptypes.QueryClient

// BridgeQueryClient is a type to define the bridge types Query Client
type BridgeQueryClient = bridgetypes.QueryClient

// StorageQueryClient is a type to define the storage types Query Client
type StorageQueryClient = storagetypes.QueryClient

// GovQueryClientV1 is a type to define the governance types Query Client V1
type GovQueryClientV1 = govv1.QueryClient

// OracleQueryClient is a type to define the oracle types Query Client
type OracleQueryClient = oracletypes.QueryClient

// ParamsQueryClient is a type to define the parameters proposal types Query Client
type ParamsQueryClient = paramstypes.QueryClient

// SlashingQueryClient is a type to define the slashing types Query Client
type SlashingQueryClient = slashingtypes.QueryClient

// StakingQueryClient is a type to define the staking types Query Client
type StakingQueryClient = stakingtypes.QueryClient

// TxClient is a type to define the tx Service Client
type TxClient = tx.ServiceClient

// UpgradeQueryClient is a type to define the upgrade types Query Client
type UpgradeQueryClient = upgradetypes.QueryClient

// GreenfieldClient holds all necessary information for creating/querying transactions.
type GreenfieldClient struct {
	// AuthQueryClient holds the auth query client.
	AuthQueryClient
	// AuthzQueryClient holds the authz query client.
	AuthzQueryClient
	// BankQueryClient holds the bank query client.
	BankQueryClient
	// ChallengeQueryClient holds the bank query client.
	ChallengeQueryClient
	// CrosschainQueryClient holds the crosschain query client.
	CrosschainQueryClient
	// DistrQueryClient holds the distr query client.
	DistrQueryClient
	// FeegrantQueryClient holds the feegrant query client.
	FeegrantQueryClient
	// GashubQueryClient holds the gashub query client.
	GashubQueryClient
	// PaymentQueryClient holds the payment query client.
	PaymentQueryClient
	// SpQueryClient holds the sp query client.
	SpQueryClient
	// BridgeQueryClient holds the bridge query client.
	BridgeQueryClient
	// StorageQueryClient holds the storage query client.
	StorageQueryClient
	// GovQueryClientV1 holds the gov query client V1.
	GovQueryClientV1
	// OracleQueryClient holds the oracle query client.
	OracleQueryClient
	// ParamsQueryClient holds the params query client.
	ParamsQueryClient
	// SlashingQueryClient holds the slashing query client.
	SlashingQueryClient
	// StakingQueryClient holds the staking query client.
	StakingQueryClient
	// UpgradeQueryClient holds the upgrade query client.
	UpgradeQueryClient
	// TxClient holds the tx service client.
	TxClient

	// keyManager is the manager used for generating and managing keys.
	keyManager keys.KeyManager
	// chainId is the id of the chain.
	chainId string
	// codec is the ProtoCodec used for encoding and decoding messages.
	codec *codec.ProtoCodec

	// option fields
	// grpcDialOption is the list of grpc dial options.
	grpcDialOption []grpc.DialOption
}

// grpcConn is used to establish a connection with a given address and dial options.
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

// NewGreenfieldClient is used to create a new GreenfieldClient structure.
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
	client.ChallengeQueryClient = challengetypes.NewQueryClient(conn)
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

// SetKeyManager sets a key manager in the GreenfieldClient structure.
func (c *GreenfieldClient) SetKeyManager(keyManager keys.KeyManager) {
	c.keyManager = keyManager
}

// GetKeyManager returns the key manager set in the GreenfieldClient structure.
func (c *GreenfieldClient) GetKeyManager() (keys.KeyManager, error) {
	if c.keyManager == nil {
		return nil, types.KeyManagerNotInitError
	}
	return c.keyManager, nil
}

// SetChainId sets the chain ID in the GreenfieldClient structure.
func (c *GreenfieldClient) SetChainId(id string) {
	c.chainId = id
}

// GetChainId returns the chain ID set in the GreenfieldClient structure.
func (c *GreenfieldClient) GetChainId() (string, error) {
	if c.chainId == "" {
		return "", types.ChainIdNotSetError
	}
	return c.chainId, nil
}

func (c *GreenfieldClient) GetCodec() *codec.ProtoCodec {
	return c.codec
}
