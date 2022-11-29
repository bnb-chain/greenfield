package cli

import (
	"fmt"
	// "strings"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	// sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/bfs/x/payment/types"
)

// GetQueryCmd returns the cli query commands for this module
func GetQueryCmd(queryRoute string) *cobra.Command {
	// Group payment queries under a subcommand
	cmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      fmt.Sprintf("Querying commands for the %s module", types.ModuleName),
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(CmdQueryParams())
	cmd.AddCommand(CmdListStreamRecord())
	cmd.AddCommand(CmdShowStreamRecord())
	cmd.AddCommand(CmdListPaymentAccountCount())
	cmd.AddCommand(CmdShowPaymentAccountCount())
	cmd.AddCommand(CmdListPaymentAccount())
	cmd.AddCommand(CmdShowPaymentAccount())
	cmd.AddCommand(CmdDynamicBalance())

	cmd.AddCommand(CmdGetPaymentAccountsByOwner())

	// this line is used by starport scaffolding # 1

	return cmd
}
