package cli_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/payment/client/cli"
	"github.com/bnb-chain/greenfield/x/payment/types"
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
			"query dynamic-balance",
			append(
				[]string{
					"dynamic-balance",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryDynamicBalanceResponse{},
		},
		{
			"query get-payment-accounts-by-owner",
			append(
				[]string{
					"get-payment-accounts-by-owner",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryPaymentAccountsByOwnerResponse{},
		},
		{
			"query list-auto-settle-record",
			append(
				[]string{
					"list-auto-settle-record",
				},
				commonFlags...,
			),
			false, "", &types.QueryAutoSettleRecordsResponse{},
		},
		{
			"query list-payment-account",
			append(
				[]string{
					"list-payment-account",
				},
				commonFlags...,
			),
			false, "", &types.QueryPaymentAccountsResponse{},
		},
		{
			"query list-payment-account-count",
			append(
				[]string{
					"list-payment-account-count",
				},
				commonFlags...,
			),
			false, "", &types.QueryPaymentAccountCountsResponse{},
		},
		{
			"query list-stream-record",
			append(
				[]string{
					"list-stream-record",
				},
				commonFlags...,
			),
			false, "", &types.QueryStreamRecordsResponse{},
		},
		{
			"query show-payment-account",
			append(
				[]string{
					"show-payment-account",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryPaymentAccountResponse{},
		},
		{
			"query show-payment-account-count",
			append(
				[]string{
					"show-payment-account-count",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryPaymentAccountCountResponse{},
		},
		{
			"query show-stream-record",
			append(
				[]string{
					"show-stream-record",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryGetStreamRecordResponse{},
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
