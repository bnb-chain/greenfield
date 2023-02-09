package cli

import (
	"context"

	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdListMockObjectInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-mock-object-info",
		Short: "list all mock-object-info",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllMockObjectInfoRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.MockObjectInfoAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowMockObjectInfo() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-mock-object-info [bucket-name] [object-name]",
		Short: "shows a mock-object-info",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argBucketName := args[0]
			argObjectName := args[1]

			params := &types.QueryGetMockObjectInfoRequest{
				BucketName: argBucketName,
				ObjectName: argObjectName,
			}

			res, err := queryClient.MockObjectInfo(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
