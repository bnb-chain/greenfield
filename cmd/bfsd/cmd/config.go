package cmd

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/bfs/app"
)

func initSDKConfig() {
	// Set and seal config
	config := sdk.GetConfig()
	config.SetCoinType(app.CoinType)
	config.Seal()
}
