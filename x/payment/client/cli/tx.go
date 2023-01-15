package cli

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	// "github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/bnb-chain/bfs/x/payment/types"
)

var (
	DefaultRelativePacketTimeoutTimestamp = uint64((time.Duration(10) * time.Minute).Nanoseconds())
)

const (
	flagPacketTimeoutTimestamp = "packet-timeout-timestamp"
	listSeparator              = ","
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

	cmd.AddCommand(CmdCreatePaymentAccount())
	cmd.AddCommand(CmdDeposit())
	cmd.AddCommand(CmdWithdraw())
	cmd.AddCommand(CmdSponse())
	cmd.AddCommand(CmdDisableRefund())
	cmd.AddCommand(CmdMockCreateBucket())
	cmd.AddCommand(CmdMockPutObject())
cmd.AddCommand(CmdMockSealObject())
cmd.AddCommand(CmdMockDeleteObject())
cmd.AddCommand(CmdMockSetBucketPaymentAccount())
// this line is used by starport scaffolding # 1

	return cmd
}
