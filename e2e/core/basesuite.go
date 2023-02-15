package core

import (
	"context"
	client "github.com/bnb-chain/greenfield/sdk/client/chain"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
	"time"
)

type BaseSuite struct {
	suite.Suite
	config      *Config
	Client      client.GreenfieldClient
	TestAccount keys.KeyManager
}

func (s *BaseSuite) SetupSuite() {
	s.config = InitConfig()
	s.Client = client.NewGreenfieldClient(s.config.GrpcAddr, s.config.ChainId)
	var err error
	s.TestAccount, err = keys.NewMnemonicKeyManager(s.config.Mnemonic)
	s.Require().NoError(err)
}

func (s *BaseSuite) SendTxBlock(msg sdk.Msg, from keys.KeyManager) (txRes *tx.GetTxResponse) {
	mode := tx.BroadcastMode_BROADCAST_MODE_BLOCK
	txOpt := &types.TxOption{
		Mode:      &mode,
		GasLimit:  1000000,
		Memo:      "",
		FeeAmount: sdk.Coins{{Denom: s.config.Denom, Amount: sdk.NewInt(1)}},
	}
	s.Client.SetKeyManager(from)
	response, err := s.Client.BroadcastTx([]sdk.Msg{msg}, txOpt)
	s.Require().NoError(err)
	s.T().Logf("tx_hash: %s", response.TxResponse.TxHash)
	//s.T().Log(response)
	s.Require().Equal(response.TxResponse.Code, uint32(0))
	retry := 10
	for {
		getTxRequest := &tx.GetTxRequest{
			Hash: response.TxResponse.TxHash,
		}
		txRes, err = s.Client.GetTx(context.Background(), getTxRequest)
		if err == nil {
			return
		}
		s.Require().ErrorContains(err, "tx not found")
		retry--
		if retry < 0 {
			s.Require().Fail("reach max retry")
		}
		time.Sleep(time.Second)
	}
}

func (s *BaseSuite) GenAndChargeAccounts(n int, balance int64) (accounts []keys.KeyManager) {
	var outputs []banktypes.Output
	denom := s.config.Denom
	for i := 0; i < n; i++ {
		km := GenRandomKeyManager()
		accounts = append(accounts, km)
		outputs = append(outputs, banktypes.Output{
			Address: km.GetAddr().String(),
			Coins:   []sdk.Coin{{Denom: denom, Amount: sdk.NewInt(balance)}},
		})
	}
	if balance == 0 {
		return
	}
	in := banktypes.Input{
		Address: s.TestAccount.GetAddr().String(),
		Coins:   []sdk.Coin{{Denom: denom, Amount: sdk.NewInt(balance * int64(n))}},
	}
	msg := banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{in},
		Outputs: outputs,
	}
	_ = s.SendTxBlock(&msg, s.TestAccount)
	return accounts
}
