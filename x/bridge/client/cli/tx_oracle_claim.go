package cli

import (
	"bytes"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/cosmos/cosmos-sdk/x/oracle/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdOracleClaim() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim [from]",
		Short: "Broadcast message transfer-out",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgClaim(
				clientCtx.GetFromAddress().String(),
				1,
				1,
				1,
				1,
				[]byte{1, 2, 3, 4},
				[]uint64{1, 2, 3, 4},
				bytes.Repeat([]byte{1}, 96),
			)
			//if err := msg.ValidateBasic(); err != nil {
			//	return err
			//}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
