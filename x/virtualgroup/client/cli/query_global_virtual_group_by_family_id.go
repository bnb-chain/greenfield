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

func CmdGlobalVirtualGroupByFamilyID() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-virtual-group-by-family-id [family-id]",
		Short: "query virtual group by family id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			familyID, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || familyID <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGlobalVirtualGroupByFamilyIDRequest{
				GlobalVirtualGroupFamilyId: uint32(familyID),
			}

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
