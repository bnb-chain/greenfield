package cli

import (
	"context"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/storage/types"
)

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
	cmd.AddCommand(CmdVerifyPermission())
	cmd.AddCommand(CmdHeadGroup())
	cmd.AddCommand(CmdListGroup())
	cmd.AddCommand(CmdHeadGroupMember())

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

func CmdVerifyPermission() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "verify-permission",
		Short: "Query verify-permission",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqOperator := args[0]
			reqBucketName := args[1]
			reqObjectName := args[2]
			reqActionType := args[3]

			actionType, err := GetActionType(reqActionType)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryVerifyPermissionRequest{
				Operator:   reqOperator,
				BucketName: reqBucketName,
				ObjectName: reqObjectName,
				ActionType: actionType,
			}

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

func CmdHeadGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "head-group [group-owner] [group-name]",
		Short: "Query head-group",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqGroupOwner := args[0]
			reqGroupName := args[1]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryHeadGroupRequest{
				GroupOwner: reqGroupOwner,
				GroupName:  reqGroupName,
			}

			res, err := queryClient.HeadGroup(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdListGroup() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-group [group-owner]",
		Short: "Query list-group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqGroupOwner := args[0]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryListGroupRequest{
				GroupOwner: reqGroupOwner,
			}

			res, err := queryClient.ListGroup(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdHeadGroupMember() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "head-group-member [group-owner] [group-name] [group-member]",
		Short: "Query head-group-member",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqGroupOwner := args[0]
			reqGroupName := args[1]
			reqGroupMember := args[2]

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryHeadGroupMemberRequest{
				GroupOwner: reqGroupOwner,
				GroupName:  reqGroupName,
				Member:     reqGroupMember,
			}

			res, err := queryClient.HeadGroupMember(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
