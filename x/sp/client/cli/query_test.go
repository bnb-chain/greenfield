package cli_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/sp/client/cli"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *CLITestSuite) TestQueryCmd() {
	commonFlags := []string{
		fmt.Sprintf("--%s=%s", flags.FlagOutput, "json"),
	}

	testCases := []struct {
		name         string
		args         []string
		expectErr    bool
		expectErrMsg string
		respType     proto.Message
	}{
		{
			"query params",
			append(
				[]string{
					"params",
				},
				commonFlags...,
			),
			false, "", &types.QueryParamsResponse{},
		},
		{
			"query storage-provider",
			append(
				[]string{
					"storage-provider",
					"1",
				},
				commonFlags...,
			),
			false, "", &types.QueryStorageProviderResponse{},
		},
		{
			"query storage-provider-by-operator-address",
			append(
				[]string{
					"storage-provider-by-operator-address",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryStorageProviderByOperatorAddressResponse{},
		},
		{
			"query storage-providers",
			append(
				[]string{
					"storage-providers",
				},
				commonFlags...,
			),
			false, "", &types.QueryStorageProvidersResponse{},
		},
	}

	for _, tc := range testCases {
		tc := tc

		s.Run(tc.name, func() {
			cmd := cli.GetQueryCmd()
			out, err := clitestutil.ExecTestCLICmd(s.clientCtx, cmd, tc.args)

			if tc.expectErr {
				s.Require().Error(err)
				s.Require().Contains(err.Error(), tc.expectErrMsg)
			} else {
				s.Require().NoError(err)
				s.Require().NoError(s.clientCtx.Codec.UnmarshalJSON(out.Bytes(), tc.respType), out.String())
			}
		})
	}
}
