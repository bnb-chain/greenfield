package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/types/query"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var _ = strconv.Itoa(0)

func CmdGlobalVirtualGroupFamilies() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-virtual-group-families [limit]",
		Short: "query all global virtual groups families of the storage provider.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			limit, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || limit <= 0 {
				return fmt.Errorf("invalid limit %s", args[0])
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGlobalVirtualGroupFamiliesRequest{
				Pagination: &query.PageRequest{Limit: uint64(limit)},
			}

			res, err := queryClient.GlobalVirtualGroupFamilies(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
