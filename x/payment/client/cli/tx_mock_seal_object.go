package cli

import (
	"strconv"

	"github.com/bnb-chain/greenfield/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
	"strings"
)

var _ = strconv.Itoa(0)

func CmdMockSealObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mock-seal-object [bucket-name] [object-name] [secondarySPs]",
		Short: "Broadcast message mock-seal-object",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argObjectName := args[1]
			argSecondarySPs := strings.Split(args[2], listSeparator)

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgMockSealObject(
				clientCtx.GetFromAddress().String(),
				argBucketName,
				argObjectName,
				argSecondarySPs,
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
