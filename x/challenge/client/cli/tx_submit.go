package cli

import (
	"fmt"
	"strconv"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdSubmit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [sp-operator-address] [object-id] [random-index] [index]",
		Short: "Broadcast message submit",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSpOperatorAddress, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}
			// validate that the object id is an uint
			argObjectId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("object-id %s not a valid uint, please input a valid object-id", args[1])
			}

			//TODO: parse args
			argRandomIndex := true
			// validate that the challenge id is an uint
			argIndex, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil {
				return fmt.Errorf("index %s not a valid uint, please input a valid index", args[3])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSubmit(
				clientCtx.GetFromAddress(),
				argSpOperatorAddress,
				argObjectId,
				argRandomIndex,
				uint32(argIndex),
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
