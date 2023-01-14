package cli

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListAutoSettleQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-auto-settle-queue",
		Short: "list all auto-settle-queue",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllAutoSettleQueueRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.AutoSettleQueueAll(context.Background(), params)
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

func CmdShowAutoSettleQueue() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-auto-settle-queue [timestamp] [user]",
		Short: "shows a auto-settle-queue",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argTimestamp, err := cast.ToInt64E(args[0])
			if err != nil {
				return err
			}
			argUser := args[1]

			params := &types.QueryGetAutoSettleQueueRequest{
				Timestamp: argTimestamp,
				User:      argUser,
			}

			res, err := queryClient.AutoSettleQueue(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
