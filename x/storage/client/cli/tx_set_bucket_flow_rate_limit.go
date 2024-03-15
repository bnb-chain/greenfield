package cli

import (
	"fmt"
	"strconv"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

var _ = strconv.Itoa(0)

func CmdSetBucketFlowRateLimit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-bucket-flow-rate-limit [bucket-name] [payment-account] [bucket-owner] [flow-rate-limit]",
		Short: "set flow rate limit for a bucket",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argBucketName := args[0]
			argPaymentAcc := args[1]
			paymentAcc, err := sdk.AccAddressFromHexUnsafe(argPaymentAcc)
			if err != nil {
				return err
			}
			argBucketOwner := args[2]
			bucketOwner, err := sdk.AccAddressFromHexUnsafe(argBucketOwner)
			if err != nil {
				return err
			}
			argFlowRateLimit, ok := sdkmath.NewIntFromString(args[3])
			if !ok {
				return fmt.Errorf("invalid flow-rate-limit: %s", args[3])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgSetBucketFlowRateLimit(
				clientCtx.GetFromAddress(),
				bucketOwner,
				paymentAcc,
				argBucketName,
				argFlowRateLimit,
			)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}
