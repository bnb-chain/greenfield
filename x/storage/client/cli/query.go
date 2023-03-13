package cli

import (
	"context"
	"fmt"

	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

// TODO: Support List bucket/object/group with pagination.
// TODO: Support HeadGroup

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group storage queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdHeadBucket())
	cmd.AddCommand(CmdHeadObject())
	cmd.AddCommand(CmdListBuckets())
	cmd.AddCommand(CmdListObjects())

	cmd.AddCommand(CmdGetPolicy())

	cmd.AddCommand(CmdVerifyPermission())

	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdHeadBucket() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "head-bucket [bucket-name]",
		Short: "Query bucket by bucket name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqBucketName := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryHeadBucketRequest{
				BucketName: reqBucketName,
			}

			res, err := queryClient.HeadBucket(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdHeadObject() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "head-object [bucket-name] [object-name]",
		Short: "Query object by bucket-name and object-name",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqBucketName := args[0]
			reqObjectName := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryHeadObjectRequest{
				BucketName: reqBucketName,
				ObjectName: reqObjectName,
			}

			res, err := queryClient.HeadObject(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListBuckets() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-buckets",
		Short: "Query listBuckets",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx := client.GetClientContextFromCmd(cmd)
			pageReq, err := client.ReadPageRequest(cmd.Flags())
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryListBucketsRequest{
				Pagination: pageReq,
			}

			res, err := queryClient.ListBuckets(context.Background(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListObjects() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-objects [bucket-name]",
		Short: "Query listObjects",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqBucketName := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryListObjectsRequest{

				BucketName: reqBucketName,
			}

			res, err := queryClient.ListObjects(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdGetPolicy() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bucket-policy [policy-id]",
		Short: "Query bucket-policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqPolicyId := args[0]

			ID, err := math.ParseUint(reqPolicyId)
			if err != nil {
				return err
			}
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryGetPolicyRequest{
				PolicyId: ID.String(),
			}

			res, err := queryClient.GetPolicy(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdVerifyPermission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-permission",
		Short: "Query verify-permission",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryVerifyPermissionRequest{}

			res, err := queryClient.VerifyPermission(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
