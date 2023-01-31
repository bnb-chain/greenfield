package cli

import (
	"context"

	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
)

func CmdListBnbPricePrice() *cobra.Command {
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

			params := &types.QueryAllBnbPricePriceRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.BnbPricePriceAll(context.Background(), params)
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

func CmdShowBnbPricePrice() *cobra.Command {
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

			params := &types.QueryGetBnbPricePriceRequest{
				Time: argTime,
			}

			res, err := queryClient.BnbPricePrice(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
