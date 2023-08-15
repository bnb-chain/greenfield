package cli_test

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client/flags"
	clitestutil "github.com/cosmos/cosmos-sdk/testutil/cli"
	"github.com/cosmos/gogoproto/proto"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/storage/client/cli"
	"github.com/bnb-chain/greenfield/x/storage/types"
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
			"query head-bucket",
			append(
				[]string{
					"head-bucket",
					"bucketName",
				},
				commonFlags...,
			),
			false, "", &types.QueryHeadBucketResponse{},
		},
		{
			"query head-group",
			append(
				[]string{
					"head-group",
					sample.RandAccAddressHex(),
					"groupName",
				},
				commonFlags...,
			),
			false, "", &types.QueryHeadGroupResponse{},
		},
		{
			"query head-group-member",
			append(
				[]string{
					"head-group-member",
					sample.RandAccAddressHex(),
					"groupName",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryHeadGroupMemberResponse{},
		},
		{
			"query head-object",
			append(
				[]string{
					"head-object",
					"bucketName",
					"objectName",
				},
				commonFlags...,
			),
			false, "", &types.QueryHeadObjectResponse{},
		},
		{
			"query list-buckets",
			append(
				[]string{
					"list-buckets",
				},
				commonFlags...,
			),
			false, "", &types.QueryListBucketsResponse{},
		},
		{
			"query list-groups",
			append(
				[]string{
					"list-groups",
					sample.RandAccAddressHex(),
				},
				commonFlags...,
			),
			false, "", &types.QueryListGroupsResponse{},
		},
		{
			"query list-objects",
			append(
				[]string{
					"list-objects",
					"bucketName",
				},
				commonFlags...,
			),
			false, "", &types.QueryListObjectsResponse{},
		},
		{
			"query verify-permission",
			append(
				[]string{
					"verify-permission",
					sample.RandAccAddressHex(),
					"bucketName",
					"objectName",
					"ACTION_TYPE_ALL",
				},
				commonFlags...,
			),
			false, "", &types.QueryVerifyPermissionResponse{},
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
