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

func CmdGlobalVirtualGroupFamily() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-virtual-group-family [sp-id] [family-id]",
		Short: "query global virtual group family by its id.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			spID, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || spID <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}

			familyID, err := strconv.ParseInt(args[1], 10, 32)
			if err != nil || familyID <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGlobalVirtualGroupFamilyRequest{
				StorageProviderId: uint32(spID),
				FamilyId:          uint32(familyID),
			}

			res, err := queryClient.GlobalVirtualGroupFamily(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
