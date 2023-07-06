package cli

import (
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/spf13/cobra"

	"github.com/bnb-chain/greenfield/x/virtualgroup/types"
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

	cmd.AddCommand(CmdGlobalVirtualGroup())
	cmd.AddCommand(CmdGlobalVirtualGroupByFamilyID())
	cmd.AddCommand(CmdGlobalVirtualGroupFamilies())
	cmd.AddCommand(CmdGlobalVirtualGroupFamily())

	// this line is used by starport scaffolding # 1

	return cmd
}
