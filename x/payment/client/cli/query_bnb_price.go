package cli

import (
    "context"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
    "github.com/bnb-chain/bfs/x/payment/types"
)

func CmdShowBnbPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-bnb-price",
		Short: "shows bnb-price",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx := client.GetClientContextFromCmd(cmd)

            queryClient := types.NewQueryClient(clientCtx)

            params := &types.QueryGetBnbPriceRequest{}

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
