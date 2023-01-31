package cli

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListBnbPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-bnb-price-price",
		Short: "list all bnb-price-price",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllBnbPriceRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.BnbPriceAll(context.Background(), params)
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

func CmdShowBnbPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-bnb-price-price [time]",
		Short: "shows a bnb-price-price",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			argTime, err := cast.ToInt64E(args[0])
			if err != nil {
				return err
			}

			params := &types.QueryGetBnbPriceRequest{
				Time: argTime,
			}

			res, err := queryClient.BnbPrice(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
