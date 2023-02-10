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

func CmdMockUpdateBucketReadPacket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-update-bucket-read-packet [bucket-name] [read-packet]",
		Short: "Broadcast message mock-update-bucket-read-packet",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argReadPacket := args[1]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMockUpdateBucketReadPacket(
				clientCtx.GetFromAddress().String(),
				argBucketName,
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
