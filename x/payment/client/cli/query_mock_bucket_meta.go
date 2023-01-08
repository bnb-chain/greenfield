package cli

import (
    "context"
	
    "github.com/spf13/cobra"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
    "github.com/bnb-chain/bfs/x/payment/types"
)

func CmdListMockBucketMeta() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-mock-bucket-meta",
		Short: "list all mock-bucket-meta",
		RunE: func(cmd *cobra.Command, args []string) error {
            clientCtx := client.GetClientContextFromCmd(cmd)

            pageReq, err := client.ReadPageRequest(cmd.Flags())
            if err != nil {
                return err
            }

            queryClient := types.NewQueryClient(clientCtx)

            params := &types.QueryAllMockBucketMetaRequest{
                Pagination: pageReq,
            }

            res, err := queryClient.MockBucketMetaAll(context.Background(), params)
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

func CmdShowMockBucketMeta() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-mock-bucket-meta [bucket-name]",
		Short: "shows a mock-bucket-meta",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
            clientCtx := client.GetClientContextFromCmd(cmd)

            queryClient := types.NewQueryClient(clientCtx)

             argBucketName := args[0]
            
            params := &types.QueryGetMockBucketMetaRequest{
                BucketName: argBucketName,
                
            }

            res, err := queryClient.MockBucketMeta(context.Background(), params)
            if err != nil {
                return err
            }

            return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

    return cmd
}
