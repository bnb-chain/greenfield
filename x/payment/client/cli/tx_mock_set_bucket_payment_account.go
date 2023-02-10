package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/payment/types"
)

var _ = strconv.Itoa(0)

func CmdMockSetBucketPaymentAccount() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-set-bucket-payment-account [bucket-name] [read-payment-account] [store-payment-account]",
		Short: "Broadcast message mock-set-bucket-payment-account",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argReadPaymentAccount := args[1]
			argStorePaymentAccount := args[2]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMockSetBucketPaymentAccount(
				clientCtx.GetFromAddress().String(),
				argBucketName,
				argReadPaymentAccount,
				argStorePaymentAccount,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
