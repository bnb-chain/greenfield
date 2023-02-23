package cli

import (
	"fmt"
	"time"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

const (
	FlagSpendLimit  = "spend-limit"
	FlagSpAddress   = "SPAddress"
	FlagExpiration  = "expiration"
	DefaultEndpoint = "sp0.greenfield.io"
)

// GetTxCmd returns the transaction commands for this module
func GetTxCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("%s transactions subcommands", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdDeposit())
	cmd.AddCommand(CmdEditStorageProvider())
	cmd.AddCommand(CmdGrantDepositAuthorization())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdEditStorageProvider() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "edit-storage-provider [sp-address]",
		Short: "Broadcast message editStorageProvider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {

			endpoint, _ := cmd.Flags().GetString(FlagEndpoint)

			argSpAddress := args[0]

			spAddress, err := sdk.AccAddressFromHexUnsafe(argSpAddress)
			if err != nil {
				return err
			}

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgEditStorageProvider(
				spAddress,
				endpoint,
				types.Description{},
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdDeposit() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deposit [sp-address] [value]",
		Short: "Broadcast message deposit",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			argValue := args[0]

			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			// TODO:impl later
			coin := sdk.Coin{Denom: argValue}
			msg := types.NewMsgDeposit(
				clientCtx.GetFromAddress(),
				clientCtx.GetFromAddress(),
				coin,
			)
			if err := msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)

	return cmd
}

func CmdGrantDepositAuthorization() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "grant <grantee> --from <granter>",
		Short: "Broadcast message deposit authorization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			grantee, err := sdk.AccAddressFromHexUnsafe(args[0])
			if err != nil {
				return err
			}
			var authorization authz.Authorization
			var depositLimit *sdk.Coin

			spAddressString, err := cmd.Flags().GetString(FlagSpAddress)
			if err != nil {
				return err
			}
			spAddress, err := sdk.AccAddressFromHexUnsafe(spAddressString)
			if err != nil {
				return err
			}

			limit, err := cmd.Flags().GetString(FlagSpendLimit)
			if err != nil {
				return err
			}
			spendLimit, err := sdk.ParseCoinNormalized(limit)
			if err != nil {
				return err
			}

			queryClient := types.NewQueryClient(clientCtx)

			res, err := queryClient.Params(cmd.Context(), &types.QueryParamsRequest{})
			if err != nil {
				return err
			}

			if spendLimit.Denom != res.Params.DepositDenom {
				return fmt.Errorf("invalid denom %s; coin denom should match the current bond denom %s", spendLimit.Denom, res.Params.DepositDenom)
			}

			if !spendLimit.IsPositive() {
				return fmt.Errorf("spend-limit should be greater than zero")
			}
			depositLimit = &spendLimit

			authorization, err = types.NewDepositAuthorization(spAddress, depositLimit)
			if err != nil {
				return err
			}

			expire, err := getExpireTime(cmd)
			if err != nil {
				return err
			}

			msg, err := authz.NewMsgGrant(clientCtx.GetFromAddress(), grantee, authorization, expire)
			if err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	cmd.Flags().String(FlagSpendLimit, "", "SpendLimit for deposit Authorization, an array of Coins allowed spend")
	cmd.Flags().String(FlagSpAddress, "", "The account address of storage provider")
	cmd.Flags().Int64(FlagExpiration, 0, "Expire time as Unix timestamp. Set zero (0) for no expiry. Default is 0.")
	return cmd
}

func getExpireTime(cmd *cobra.Command) (*time.Time, error) {
	exp, err := cmd.Flags().GetInt64(FlagExpiration)
	if err != nil {
		return nil, err
	}
	if exp == 0 {
		return nil, nil
	}
	e := time.Unix(exp, 0)
	return &e, nil
}
