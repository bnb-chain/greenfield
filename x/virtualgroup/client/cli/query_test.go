package cli_test

import (
	"fmt"

	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/x/virtualgroup/client/cli"
	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cosmos/cosmos-sdk/client/flags"
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
			"query global-virtual-group",
			append(
				[]string{
					"global-virtual-group",
					"1",
				},
				commonFlags...,
			),
			false, "", &types.QueryGlobalVirtualGroupResponse{},
		},
		{
			"query global-virtual-group-by-family-id",
			append(
				[]string{
					"global-virtual-group-by-family-id",
					"1",
				},
				commonFlags...,
			),
			false, "", &types.QueryGlobalVirtualGroupByFamilyIDResponse{},
		},
		{
			"query global-virtual-group-families",
			append(
				[]string{
					"global-virtual-group-families",
					"100",
				},
				commonFlags...,
			),
			false, "", &types.QueryGlobalVirtualGroupFamiliesResponse{},
		},
		{
			"query global-virtual-group-family",
			append(
				[]string{
					"global-virtual-group-family",
					"1",
				},
				commonFlags...,
			),
			false, "", &types.QueryGlobalVirtualGroupFamilyResponse{},
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
