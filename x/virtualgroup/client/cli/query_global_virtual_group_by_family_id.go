package cli

import (
	"strconv"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"
)

var _ = strconv.Itoa(0)

func CmdGlobalVirtualGroupByFamilyID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-virtual-group-by-family-id",
		Short: "Query GlobalVirtualGroupByFamilyID",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGlobalVirtualGroupByFamilyIDRequest{}

			res, err := queryClient.GlobalVirtualGroupByFamilyID(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
