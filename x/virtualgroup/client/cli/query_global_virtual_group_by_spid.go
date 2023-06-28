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

func CmdGlobalVirtualGroupBySPID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-virtual-group-by-spid [sp-id] [limit]",
		Short: "query all global virtual groups of the storage provider.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			spID, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || spID <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}

			limit, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil || limit <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGlobalVirtualGroupBySPIDRequest{
				StorageProviderId: uint32(spID),
				Pagination:        &query.PageRequest{Limit: uint64(limit)},
			}

			res, err := queryClient.GlobalVirtualGroupBySPID(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
