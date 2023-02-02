package cli

import (
	"strconv"

	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdMockCreateBucket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-create-bucket [bucket-name] [read-payment-account] [store-payment-account] [sp-address] [read-packet]",
		Short: "Broadcast message mock-create-bucket",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argReadPaymentAccount := args[1]
			argStorePaymentAccount := args[2]
			argSpAddress := args[3]
			argReadPacket := args[4]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMockCreateBucket(
				clientCtx.GetFromAddress().String(),
				argBucketName,
				argReadPaymentAccount,
				argStorePaymentAccount,
				argSpAddress,
				argReadPacket,
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
