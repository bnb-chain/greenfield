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
		Use:   "submit [sp-operator-address] [bucket-name] [object-name] [random-index] [index]",
		Short: "Broadcast message submit",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSpOperatorAddress, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}

			argBucketName := args[1]
			argObjectName := args[2]

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
				argBucketName,
				argObjectName,
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
