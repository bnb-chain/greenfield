package client

import (
	_ "encoding/json"
	"strings"

	sdkclient "github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/tx"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	crosschaintypes "github.com/cosmos/cosmos-sdk/x/crosschain/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	grpc1 "github.com/cosmos/gogoproto/grpc"
	"google.golang.org/grpc"

	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	challengetypes "github.com/bnb-chain/greenfield/x/challenge/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
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

// SlashingQueryClient is a type to define the slashing types Query Client
type SlashingQueryClient = slashingtypes.QueryClient

// StakingQueryClient is a type to define the staking types Query Client
type StakingQueryClient = stakingtypes.QueryClient

// TxClient is a type to define the tx Service Client
type TxClient = tx.ServiceClient

// UpgradeQueryClient is a type to define the upgrade types Query Client
type UpgradeQueryClient = upgradetypes.QueryClient

// VirtualGroupQueryClient is a type to define the virtual group types Query Client
type VirtualGroupQueryClient = virtualgroupmoduletypes.QueryClient

// TmClient is a type to define the tendermint service client
type TmClient = tmservice.ServiceClient

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
	// SlashingQueryClient holds the slashing query client.
	SlashingQueryClient
	// StakingQueryClient holds the staking query client.
	StakingQueryClient
	// UpgradeQueryClient holds the upgrade query client.
	UpgradeQueryClient
	// VirtualGroupQueryClient holds the virtual group query client
	VirtualGroupQueryClient
	// TxClient holds the tx service client.
	TxClient
	// TmService holds the tendermint service client
	TmClient

	// keyManager is the manager used for generating and managing keys.
	keyManager keys.KeyManager
	// chainId is the id of the chain.
	chainId string
	// codec is the ProtoCodec used for encoding and decoding messages.
	codec *codec.ProtoCodec
	// grpcConn is for client initialization using grpc connection
	grpcConn *grpc.ClientConn
}

// NewGreenfieldClient is used to create a new GreenfieldClient structure.
func NewGreenfieldClient(rpcAddr, chainId string, opts ...GreenfieldClientOption) (*GreenfieldClient, error) {
	cdc := types.Codec()
	client := &GreenfieldClient{
		chainId: chainId,
		codec:   cdc,
	}
	for _, opt := range opts {
		opt.Apply(client)
	}
	if client.grpcConn != nil {
		setClientsConn(client, client.grpcConn)
		return client, nil
	}
	if len(strings.TrimSpace(rpcAddr)) == 0 {
		return nil, types.RpcAddressNotProvidedError
	}
	rpcClient, err := sdkclient.NewClientFromNode(rpcAddr)
	if err != nil {
		return nil, err
	}
	txConfig := authtx.NewTxConfig(cdc, []signing.SignMode{signing.SignMode_SIGN_MODE_EIP_712})
	clientCtx := sdkclient.Context{}.
		WithCodec(cdc).
		WithInterfaceRegistry(cdc.InterfaceRegistry()).
		WithTxConfig(txConfig).
		WithClient(rpcClient)

	setClientsConn(client, clientCtx)
	return client, nil
}

func setClientsConn(c *GreenfieldClient, conn grpc1.ClientConn) {
	c.AuthQueryClient = authtypes.NewQueryClient(conn)
	c.AuthQueryClient = authtypes.NewQueryClient(conn)
	c.AuthzQueryClient = authztypes.NewQueryClient(conn)
	c.BankQueryClient = banktypes.NewQueryClient(conn)
	c.ChallengeQueryClient = challengetypes.NewQueryClient(conn)
	c.CrosschainQueryClient = crosschaintypes.NewQueryClient(conn)
	c.DistrQueryClient = distrtypes.NewQueryClient(conn)
	c.FeegrantQueryClient = feegranttypes.NewQueryClient(conn)
	c.GashubQueryClient = gashubtypes.NewQueryClient(conn)
	c.PaymentQueryClient = paymenttypes.NewQueryClient(conn)
	c.SpQueryClient = sptypes.NewQueryClient(conn)
	c.BridgeQueryClient = bridgetypes.NewQueryClient(conn)
	c.StorageQueryClient = storagetypes.NewQueryClient(conn)
	c.GovQueryClientV1 = govv1.NewQueryClient(conn)
	c.OracleQueryClient = oracletypes.NewQueryClient(conn)
	c.SlashingQueryClient = slashingtypes.NewQueryClient(conn)
	c.StakingQueryClient = stakingtypes.NewQueryClient(conn)
	c.UpgradeQueryClient = upgradetypes.NewQueryClient(conn)
	c.VirtualGroupQueryClient = virtualgroupmoduletypes.NewQueryClient(conn)
	c.TmClient = tmservice.NewServiceClient(conn)
	c.TxClient = tx.NewServiceClient(conn)
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
