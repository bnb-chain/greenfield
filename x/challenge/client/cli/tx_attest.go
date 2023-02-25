package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

var _ = strconv.Itoa(0)

func CmdAttest() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "attest [challenge-id] [object-id] [sp-operator-address] [vote-result] [vote-validator-set] [vote-agg-signature]",
		Short: "Broadcast message attest",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChallengeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("challenge-id %s not a valid uint, please input a valid challenge-id", args[0])
			}

			argObjectId, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return fmt.Errorf("object-id %s not a valid uint, please input a valid object-id", args[1])
			}

			argSpOperatorAddress, err := sdk.AccAddressFromHexUnsafe(args[2])
			if err != nil {
				return fmt.Errorf("sp-operator-address %s not a valid address, please input a valid sp-operator-address", args[2])
			}

			argVoteResult, err := strconv.ParseUint(args[3], 10, 64)
			if err != nil {
				return fmt.Errorf("vote-result %s not a valid uint, please input a valid vote-result", args[3])
			}

			argVoteValidatorSet := make([]uint64, 0)
			splits := strings.Split(args[4], ",")
			for _, split := range splits {
				val, err := strconv.ParseUint(split, 10, 64)
				if err != nil {
					return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[4])
				}
				argVoteValidatorSet = append(argVoteValidatorSet, val)
			}
			if len(argVoteValidatorSet) == 0 {
				return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[4])
			}

			argVoteAggSignature, err := hex.DecodeString(args[5])
			if err != nil {
				return fmt.Errorf("vote-agg-signature %s not a hex encoded bytes, please input a valid vote-agg-signature", args[5])
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgAttest(
				clientCtx.GetFromAddress(),
				argChallengeId,
				argObjectId,
				argSpOperatorAddress,
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
