package cli

import (
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

	sdkmath "cosmossdk.io/math"
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
		Use:   "attest [challenge-id] [object-id] [sp-operator-address] [vote-result] [challenger-address] [vote-validator-set] [vote-agg-signature]",
		Short: "Broadcast message attest",
		Args:  cobra.ExactArgs(6),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argChallengeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("challenge-id %s not a valid uint, please input a valid challenge-id", args[0])
			}

			argObjectId := sdkmath.NewUintFromString(args[1])

			argSpOperatorAddress := args[2]
			_, err = sdk.AccAddressFromHexUnsafe(argSpOperatorAddress)
			if err != nil {
				return fmt.Errorf("sp-operator-address %s not a valid address, please input a valid sp-operator-address", args[2])
			}

			argVoteResult, err := strconv.ParseUint(args[3], 10, 32)
			if err != nil || argVoteResult != uint64(types.CHALLENGE_SUCCEED) {
				return fmt.Errorf("vote-result %s not a valid uint, please input a valid vote-result", args[3])
			}

			argChallengerAddress := args[4]
			if argChallengerAddress != "" {
				_, err = sdk.AccAddressFromHexUnsafe(argChallengerAddress)
				if err != nil {
					return fmt.Errorf("challenger-address %s not a valid address, please input a valid challenger-address", args[4])
				}
			}

			argVoteValidatorSet := make([]uint64, 0)
			splits := strings.Split(args[5], ",")
			for _, split := range splits {
				val, err := strconv.ParseUint(split, 10, 64)
				if err != nil {
					return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[5])
				}
				argVoteValidatorSet = append(argVoteValidatorSet, val)
			}
			if len(argVoteValidatorSet) == 0 {
				return fmt.Errorf("vote-validator-set %s not a valid comma seperated uint array, please input a valid vote-validator-set", args[5])
			}

			argVoteAggSignature, err := hex.DecodeString(args[6])
			if err != nil {
				return fmt.Errorf("vote-agg-signature %s not a hex encoded bytes, please input a valid vote-agg-signature", args[6])
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
				types.VoteResult(argVoteResult),
				argChallengerAddress,
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
