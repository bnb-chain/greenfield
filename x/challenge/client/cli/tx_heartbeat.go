package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

var _ = strconv.Itoa(0)

func CmdHeartbeat() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "heartbeat [challenge-id] [vote-validator-set] [vote-agg-signature]",
		Short: "Broadcast message heartbeat",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// validate that the challenge id is an uint
			argChallengeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("challenge-id %s not a valid uint, please input a valid challenge-id", args[0])
			}

			argVoteValidatorSet := make([]uint64, 0)
			splits := strings.Split(args[1], ",")
			for _, split := range splits {
				val, err := strconv.ParseUint(split, 10, 64)
				if err != nil {
					return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[1])
				}
				argVoteValidatorSet = append(argVoteValidatorSet, val)
			}
			if len(argVoteValidatorSet) == 0 {
				return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[1])
			}

			argVoteAggSignature, err := hex.DecodeString(args[2])
			if err != nil {
				return fmt.Errorf("vote-agg-signature %s not a hex encoded bytes, please input a valid vote-agg-signature", args[2])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgHeartbeat(
				clientCtx.GetFromAddress(),
				argChallengeId,
				argVoteValidatorSet,
				argVoteAggSignature,
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
