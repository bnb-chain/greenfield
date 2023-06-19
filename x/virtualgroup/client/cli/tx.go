package cli

import (
	"fmt"
	"strconv"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
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

	cmd.AddCommand(CmdStorageProviderExit())
	cmd.AddCommand(CmdCompleteStorageProviderExit())
	cmd.AddCommand(CmdWithdrawFromGVGFamily())
	cmd.AddCommand(CmdWithdrawFromGVG())
	// this line is used by starport scaffolding # 1

	return cmd
}

func CmdWithdrawFromGVGFamily() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-from-gvg-family",
		Short: "Broadcast message WithdrawFromGVGFamily",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			id, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || id <= 0 {
				return fmt.Errorf("invalid GVG family id %s", args[1])
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawFromGVGFamily(
				clientCtx.GetFromAddress(),
				uint32(id),
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

func CmdWithdrawFromGVG() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "withdraw-from-gvg",
		Short: "Broadcast message WithdrawFromGVG",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			id, err := strconv.ParseInt(args[0], 10, 32)
			if err != nil || id <= 0 {
				return fmt.Errorf("invalid GVG id %s", args[1])
			}
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			msg := types.NewMsgWithdrawFromGVG(
				clientCtx.GetFromAddress(),
				uint32(id),
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
