package cli

import (
    "context"
	
    "github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
    "github.com/bnb-chain/bfs/x/payment/types"
)

func CmdListFlow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-flow",
		Short: "list all flow",
		RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx := client.GetClientContextFromCmd(cmd)

            pageReq, err := client.ReadPageRequest(cmd.Flags())
            if err != nil {
                return err
            }

            queryClient := types.NewQueryClient(clientCtx)

            params := &types.QueryAllFlowRequest{
                Pagination: pageReq,
            }

            res, err := queryClient.FlowAll(context.Background(), params)
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

func CmdShowFlow() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-flow [from] [to]",
		Short: "shows a flow",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
            clientCtx := client.GetClientContextFromCmd(cmd)

            queryClient := types.NewQueryClient(clientCtx)

             argFrom := args[0]
             argTo := args[1]
            
            params := &types.QueryGetFlowRequest{
                From: argFrom,
                To: argTo,
                
            }

            res, err := queryClient.Flow(context.Background(), params)
            if err != nil {
                return err
            }

            return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

    return cmd
}
