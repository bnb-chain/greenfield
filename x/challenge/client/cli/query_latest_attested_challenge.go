package cli

import (
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/challenge/types"
)

var _ = strconv.Itoa(0)

func CmdLatestAttestedChallenge() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "latest-attested-challenge",
		Short: "Query latestAttestedChallenge",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryLatestAttestedChallengeRequest{}

			res, err := queryClient.LatestAttestedChallenge(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
