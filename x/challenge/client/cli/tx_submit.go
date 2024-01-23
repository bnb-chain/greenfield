package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/types/s3util"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

var _ = strconv.Itoa(0)

func CmdSubmit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "submit [sp-operator-address] [bucket-name] [object-name] [random-index] [segment-index]",
		Short: "Broadcast message submit",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argSpOperatorAddress, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return fmt.Errorf("sp-operator-address %s not a valid address, please input a valid sp-operator-address", args[0])
			}

			argBucketName := strings.TrimSpace(args[1])
			if err := s3util.CheckValidBucketName(argBucketName); err != nil {
				return fmt.Errorf("bucket-name %s not a valid bucket name, please input a valid bucket-name", argBucketName)
			}

			argObjectName := strings.TrimSpace(args[2])
			if err := s3util.CheckValidObjectName(argObjectName); err != nil {
				return fmt.Errorf("object-name %s not a valid object name, please input a valid object-name", argObjectName)
			}

			argRandomIndex, err := strconv.ParseBool(args[3])
			if err != nil {
				return fmt.Errorf("random-index %s not a valid bool, please input a valid random-index", args[3])
			}

			argSegmentIndex := uint64(0)
			if !argRandomIndex {
				argSegmentIndex, err = strconv.ParseUint(args[4], 10, 32)
				if err != nil {
					return fmt.Errorf("segment-index %s not a valid uint, please input a valid segment-index", args[4])
				}
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
				uint32(argSegmentIndex),
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
