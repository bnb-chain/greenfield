package core

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
)

type SPKeyManagers struct {
	OperatorKey keys.KeyManager
	SealKey     keys.KeyManager
	FundingKey  keys.KeyManager
	ApprovalKey keys.KeyManager
}

type BaseSuite struct {
	suite.Suite
	Config          *Config
	Client          *client.GreenfieldClient
	Validator       keys.KeyManager
	StorageProvider SPKeyManagers
}

func (s *BaseSuite) SetupSuite() {
	s.Config = InitConfig()
	s.Client = client.NewGreenfieldClient(s.Config.GrpcAddr, s.Config.ChainId,
		client.WithGrpcDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	var err error
	s.Validator, err = keys.NewMnemonicKeyManager(s.Config.ValidatorMnemonic)
	s.Require().NoError(err)
	s.StorageProvider.OperatorKey, err = keys.NewMnemonicKeyManager(s.Config.SPMnemonics.OperatorMnemonic)
	s.Require().NoError(err)
	s.StorageProvider.SealKey, err = keys.NewMnemonicKeyManager(s.Config.SPMnemonics.SealMnemonic)
	s.Require().NoError(err)
	s.StorageProvider.FundingKey, err = keys.NewMnemonicKeyManager(s.Config.SPMnemonics.FundingMnemonic)
	s.Require().NoError(err)
	s.StorageProvider.ApprovalKey, err = keys.NewMnemonicKeyManager(s.Config.SPMnemonics.ApprovalMnemonic)
	s.Require().NoError(err)
}

func (s *BaseSuite) SendTxBlock(msg sdk.Msg, from keys.KeyManager) (txRes *sdk.TxResponse) {
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &types.TxOption{
		Mode:      &mode,
		GasLimit:  1000000,
		Memo:      "",
		FeeAmount: sdk.Coins{{Denom: s.Config.Denom, Amount: types.NewIntFromInt64WithDecimal(1e6, types.DecimalGwei)}},
	}
	s.Client.SetKeyManager(from)
	response, err := s.Client.BroadcastTx([]sdk.Msg{msg}, txOpt)
	s.Require().NoError(err)
	s.T().Logf("tx_hash: %s", response.TxResponse.TxHash)
	if response.TxResponse.Code != uint32(0) {
		s.T().Logf("tx failed, err: %s", response.TxResponse.String())
		s.Require().Equal(response.TxResponse.Code, uint32(0))
	}
	return response.TxResponse
}

func (s *BaseSuite) GenAndChargeAccounts(n int, balance int64) (accounts []keys.KeyManager) {
	var outputs []banktypes.Output
	denom := s.Config.Denom
	for i := 0; i < n; i++ {
		km := GenRandomKeyManager()
		accounts = append(accounts, km)
		outputs = append(outputs, banktypes.Output{
			Address: km.GetAddr().String(),
			Coins:   []sdk.Coin{{Denom: denom, Amount: types.NewIntFromInt64WithDecimal(balance, types.DecimalBNB)}},
		})
	}
	if balance == 0 {
		return
	}
	in := banktypes.Input{
		Address: s.Validator.GetAddr().String(),
		Coins:   []sdk.Coin{{Denom: denom, Amount: types.NewIntFromInt64WithDecimal(balance*int64(n), types.DecimalBNB)}},
	}
	msg := banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{in},
		Outputs: outputs,
	}
	_ = s.SendTxBlock(&msg, s.Validator)
	return accounts
}
