package cli

import (
	"context"

	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListAutoSettleRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-auto-settle-record",
		Short: "list all auto-settle-record",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllAutoSettleRecordRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.AutoSettleRecordAll(context.Background(), params)
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

func CmdShowAutoSettleRecord() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-auto-settle-record [timestamp] [addr]",
		Short: "shows a auto-settle-record",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argTimestamp, err := cast.ToInt64E(args[0])
			if err != nil {
				return err
			}
			argAddr := args[1]

			params := &types.QueryGetAutoSettleRecordRequest{
				Timestamp: argTimestamp,
				Addr:      argAddr,
			}

			res, err := queryClient.AutoSettleRecord(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
