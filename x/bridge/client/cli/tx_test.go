package cli_test

import (
	"bytes"
	"fmt"
	"io"
	"testing"

	abci "github.com/cometbft/cometbft/abci/types"
	rpcclientmock "github.com/cometbft/cometbft/rpc/client/mock"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/app"
	"github.com/bnb-chain/greenfield/app/params"
	"github.com/bnb-chain/greenfield/sdk/client/test"
	"github.com/bnb-chain/greenfield/x/bridge/client/cli"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/testutil"
	"github.com/bnb-chain/greenfield/testutil/sample"
)

type CLITestSuite struct {
	suite.Suite

	kr        keyring.Keyring
	baseCtx   client.Context
	encCfg    params.EncodingConfig
	clientCtx client.Context
}

func TestCLITestSuite(t *testing.T) {
	suite.Run(t, new(CLITestSuite))
}

func (s *CLITestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	s.encCfg = app.MakeEncodingConfig()
	s.kr = keyring.NewInMemory(s.encCfg.Marshaler)
	s.baseCtx = client.Context{}.
		WithKeyring(s.kr).
		WithTxConfig(s.encCfg.TxConfig).
		WithCodec(s.encCfg.Marshaler).
		WithClient(clitestutil.MockTendermintRPC{Client: rpcclientmock.Client{}}).
		WithAccountRetriever(client.MockAccountRetriever{}).
		WithOutput(io.Discard).
		WithChainID(test.TEST_CHAIN_ID)

	accounts := testutil.CreateKeyringAccounts(s.T(), s.kr, 1)
	s.baseCtx = s.baseCtx.WithFrom(accounts[0].Address.String())
	s.baseCtx = s.baseCtx.WithFromName(accounts[0].Name)
	s.baseCtx = s.baseCtx.WithFromAddress(accounts[0].Address)

	var outBuf bytes.Buffer
	ctxGen := func() client.Context {
		bz, _ := s.encCfg.Marshaler.Marshal(&sdk.TxResponse{})
		c := clitestutil.NewMockTendermintRPC(abci.ResponseQuery{
			Value: bz,
		})

		return s.baseCtx.WithClient(c)
	}
	s.clientCtx = ctxGen().WithOutput(&outBuf)

	if testing.Short() {
		s.T().Skip("skipping test in unit-tests mode.")
	}
}

func (s *CLITestSuite) TestTxCmdTransferOut() {
	clientCtx := s.clientCtx

	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagBroadcastMode, flags.BroadcastSync),
		fmt.Sprintf("--%s=true", flags.FlagSkipConfirmation),
	}

	testCases := []struct {
		name      string
		args      []string
		expectErr bool
		expErrMsg string
	}{
		{
			"invalid to address",
			append(
				[]string{
					"invalidAddress",
					"1000000000000000000BNB",
				},
				commonFlags...,
			),
			true, "invalid address hex length",
		},
		{
			"success case",
			append(
				[]string{
					sample.RandAccAddressHex(),
					"1000000000000000000BNB",
				},
				commonFlags...,
			),
			false, "",
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.CmdTransferOut()
			out, err := clitestutil.ExecTestCLICmd(clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(clientCtx.Codec.UnmarshalJSON(out.Bytes(), &sdk.TxResponse{}), out.String())
			}
		})
	}
}
