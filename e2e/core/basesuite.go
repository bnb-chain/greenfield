package core

import (
	"context"
	"strings"

	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
)

type SPKeyManagers struct {
	OperatorKey keys.KeyManager
	SealKey     keys.KeyManager
	FundingKey  keys.KeyManager
	ApprovalKey keys.KeyManager
	GcKey       keys.KeyManager
}

type BaseSuite struct {
	suite.Suite
	Config           *Config
	Client           *client.GreenfieldClient
	TmClient         *client.TendermintClient
	Validator        keys.KeyManager
	ValidatorBLS     keys.KeyManager
	Relayer          keys.KeyManager
	Challenger       keys.KeyManager
	StorageProviders []SPKeyManagers
}

func (s *BaseSuite) SetupSuite() {
	s.Config = InitConfig()
	s.Client, _ = client.NewGreenfieldClient(s.Config.TendermintAddr, s.Config.ChainId)
	tmClient := client.NewTendermintClient(s.Config.TendermintAddr)
	s.TmClient = &tmClient
	var err error
	s.Validator, err = keys.NewMnemonicKeyManager(s.Config.ValidatorMnemonic)
	s.Require().NoError(err)
	s.ValidatorBLS, err = keys.NewBlsMnemonicKeyManager(s.Config.ValidatorBlsMnemonic)
	s.Require().NoError(err)
	s.Relayer, err = keys.NewMnemonicKeyManager(s.Config.RelayerMnemonic)
	s.Require().NoError(err)
	s.Challenger, err = keys.NewMnemonicKeyManager(s.Config.ChallengerMnemonic)
	s.Require().NoError(err)
	for _, spMnemonics := range s.Config.SPMnemonics {
		sPKeyManagers := SPKeyManagers{}
		sPKeyManagers.OperatorKey, err = keys.NewMnemonicKeyManager(spMnemonics.OperatorMnemonic)
		s.Require().NoError(err)
		sPKeyManagers.SealKey, err = keys.NewMnemonicKeyManager(spMnemonics.SealMnemonic)
		s.Require().NoError(err)
		sPKeyManagers.FundingKey, err = keys.NewMnemonicKeyManager(spMnemonics.FundingMnemonic)
		s.Require().NoError(err)
		sPKeyManagers.ApprovalKey, err = keys.NewMnemonicKeyManager(spMnemonics.ApprovalMnemonic)
		s.Require().NoError(err)
		sPKeyManagers.GcKey, err = keys.NewMnemonicKeyManager(spMnemonics.GcMnemonic)
		s.Require().NoError(err)
		s.StorageProviders = append(s.StorageProviders, sPKeyManagers)
	}
}

func (s *BaseSuite) SendTxBlock(msg sdk.Msg, from keys.KeyManager) (txRes *sdk.TxResponse) {
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	response, err := s.Client.BroadcastTx(context.Background(), []sdk.Msg{msg}, txOpt)
	s.Require().NoError(err)
	s.T().Logf("block_height: %d, tx_hash: 0x%s", response.TxResponse.Height, response.TxResponse.TxHash)
	s.Require().Equal(response.TxResponse.Code, uint32(0), "tx failed, err: %s", response.TxResponse.String())
	return response.TxResponse
}

func (s *BaseSuite) SendTxBlockWithoutCheck(msg sdk.Msg, from keys.KeyManager) (*tx.BroadcastTxResponse, error) {
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	return s.Client.BroadcastTx(context.Background(), []sdk.Msg{msg}, txOpt)
}

func (s *BaseSuite) SendTxBlockWithExpectErrorString(msg sdk.Msg, from keys.KeyManager, expectErrorString string) {
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	_, err := s.Client.BroadcastTx(context.Background(), []sdk.Msg{msg}, txOpt)
	s.T().Logf("tx failed, err: %s, expect error string: %s", err, expectErrorString)
	s.Require().Error(err)
	s.Require().True(strings.Contains(err.Error(), expectErrorString))
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
