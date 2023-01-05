package cli

import (
	sdkmath "cosmossdk.io/math"
	"fmt"
	"strconv"

	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdSponse() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sponse [to] [rate]",
		Short: "Broadcast message sponse",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argTo := args[0]
			argRate, ok := sdkmath.NewIntFromString(args[1])
			if !ok {
				return fmt.Errorf("invalid rate %s", args[1])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSponse(
				clientCtx.GetFromAddress().String(),
				argTo,
				argRate,
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
