package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd() *cobra.Command {
	// Group sp queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		CmdQueryParams(),
		CmdStorageProviders(),
		CmdStorageProvider(),
		CmdStorageProviderByOperatorAddress(),
		CmdMaintenanceRecordsBySPOperatorAddress(),
		CmdStorageProviderPrice(),
		CmdStorageProviderGlobalPrice(),
	)

	// this line is used by starport scaffolding # 1
	return cmd
}

func CmdStorageProviders() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage-providers",
		Short: "Query sp info of all storage providers",
		Args:  cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryStorageProvidersRequest{}

			res, err := queryClient.StorageProviders(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdStorageProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage-provider [sp-id]",
		Short: "Query storage provider with specify sp id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqSpID := args[0]
			spID, err := strconv.ParseUint(reqSpID, 10, 32)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)
			params := &types.QueryStorageProviderRequest{
				Id: uint32(spID),
			}

			res, err := queryClient.StorageProvider(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdStorageProviderByOperatorAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage-provider-by-operator-address [operator address]",
		Short: "Query storage provider with specify operator address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqSpAddr := args[0]

			operatorAddr, err := sdk.AccAddressFromHexUnsafe(reqSpAddr)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			params := &types.QueryStorageProviderByOperatorAddressRequest{
				OperatorAddress: operatorAddr.String(),
			}

			res, err := queryClient.StorageProviderByOperatorAddress(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdMaintenanceRecordsBySPOperatorAddress() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "maintenance-records-by-operator-address [operator address]",
		Short: "Query storage provider maintenance records with specify operator address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			reqSpAddr := args[0]

			operatorAddr, err := sdk.AccAddressFromHexUnsafe(reqSpAddr)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			queryClient := types.NewQueryClient(clientCtx)
			params := &types.QueryStorageProviderMaintenanceRecordsRequest{
				OperatorAddress: operatorAddr.String(),
			}

			res, err := queryClient.StorageProviderMaintenanceRecordsByOperatorAddress(cmd.Context(), params)
			if err != nil {
				return err
			}

			return clientCtx.PrintProto(res)
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func CmdStorageProviderPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "price [sp-address]",
		Short: "Query storage provider prices, including read, store price and free quota",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			spAddr, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).
				QuerySpStoragePrice(cmd.Context(), &types.QuerySpStoragePriceRequest{
					SpAddr: spAddr.String(),
				})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}

func CmdStorageProviderGlobalPrice() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "global-price [timestamp]",
		Short: "Query the global price at a specific time(in unix), return latest price price if timestamp is 0",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientQueryContext(cmd)
			if err != nil {
				return err
			}
			ts, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil {
				return err
			}
			res, err := types.NewQueryClient(clientCtx).
				QueryGlobalSpStorePriceByTime(cmd.Context(), &types.QueryGlobalSpStorePriceByTimeRequest{
					Timestamp: ts,
				})
			if err != nil {
				return err
			}
			return clientCtx.PrintProto(res)
		},
	}
	flags.AddQueryFlagsToCmd(cmd)
	return cmd
}
