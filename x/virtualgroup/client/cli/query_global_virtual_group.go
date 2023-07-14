package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

var _ = strconv.Itoa(0)

func CmdGlobalVirtualGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-virtual-group [id]",
		Short: "query global virtual group by its id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			gvgID, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || gvgID <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGlobalVirtualGroupRequest{
				GlobalVirtualGroupId: uint32(gvgID),
			}

			res, err := queryClient.GlobalVirtualGroup(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
