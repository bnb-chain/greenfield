package cli

import (
	"context"
	"fmt"
	"strconv"

	"github.com/bnb-chain/greenfield/x/challenge/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

func CmdListOngoingChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-ongoing-challenge",
		Short: "list all ongoing-challenge",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)

			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryAllChallengeRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.ChallengeAll(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddPaginationFlagsToCmd(cmd, cmd.Use)
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdShowOngoingChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-ongoing-challenge [challenge-id]",
		Short: "shows a ongoing-challenge",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)

			queryClient := types.NewQueryClient(clientCtx)

			// validate that the challenge id is an uint
			argChallengeId, err := strconv.ParseUint(args[0], 10, 64)
			if err != nil {
				return fmt.Errorf("challenge-id %s not a valid uint, please input a valid challenge-id", args[0])
			}

			params := &types.QueryGetChallengeRequest{
				ChallengeId: argChallengeId,
			}

			res, err := queryClient.Challenge(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
