package main

import (
	"os"

	"cosmossdk.io/log"
	svrcmd "github.com/cosmos/cosmos-sdk/server/cmd"

	"github.com/bnb-chain/greenfield/app"
	"github.com/bnb-chain/greenfield/cmd/gnfd/cmd"
)

func main() {
	rootCmd, _ := cmd.NewRootCmd()
	if err := svrcmd.Execute(rootCmd, "", app.DefaultNodeHome); err != nil {
		log.NewLogger(rootCmd.OutOrStderr()).Error("failure when running app", "err", err)
		os.Exit(1)
	}
}
