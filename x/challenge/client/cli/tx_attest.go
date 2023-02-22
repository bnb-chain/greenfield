package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdAttest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest [challenge-id] [vote-result] [vote-validator-set] [vote-agg-signature]",
		Short: "Broadcast message attest",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			// validate that the challenge id is an uint
			argChallengeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("challenge-id %s not a valid uint, please input a valid challenge-id", args[0])
			}

			// validate that the vote result id is an uint
			argVoteResult, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("vote-result %s not a valid uint, please input a valid vote-result", args[1])
			}

			argVoteValidatorSet := make([]uint64, 0)
			splits := strings.Split(args[2], ",")
			for _, split := range splits {
				val, err := strconv.ParseUint(split, 10, 64)
				if err != nil {
					return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[2])
				}
				argVoteValidatorSet = append(argVoteValidatorSet, val)
			}
			if len(argVoteValidatorSet) == 0 {
				return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[2])
			}

			argVoteAggSignature, err := hex.DecodeString(args[3])
			if err != nil {
				return fmt.Errorf("vote-agg-signature %s not a hex encoded bytes, please input a valid vote-agg-signature", args[3])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAttest(
				clientCtx.GetFromAddress(),
				argChallengeId,
				uint32(argVoteResult),
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
