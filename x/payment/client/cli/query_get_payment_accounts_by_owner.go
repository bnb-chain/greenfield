package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

var _ = strconv.Itoa(0)

func CmdGetPaymentAccountsByOwner() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-payment-accounts-by-owner [owner]",
		Short: "Query get-payment-accounts-by-owner",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqOwner := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryPaymentAccountsByOwnerRequest{

				Owner: reqOwner,
			}

			res, err := queryClient.PaymentAccountsByOwner(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
