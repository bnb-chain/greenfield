package cli

import (
	"fmt"
	"strconv"

	"github.com/bnb-chain/bfs/x/bridge/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdTransferOut() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transfer-out [from]",
		Short: "Broadcast message transfer-out",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			toAddr, err := sdk.ETHAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}

			coin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			expireTime, err := strconv.ParseUint(args[2], 10, 64)
			if err != nil {
				return fmt.Errorf("expire time(%s) is invalid", args[2])
			}

			msg := types.NewMsgTransferOut(
				clientCtx.GetFromAddress().String(),
				toAddr.String(),
				&coin,
				expireTime,
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
