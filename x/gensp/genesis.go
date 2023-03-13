package gensp

import (
	"github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/bnb-chain/greenfield/x/gensp/types"
)

// InitGenesis initializes the module's state from a provided genesis state and deliver genesis transactions.
func InitGenesis(ctx sdk.Context, stakingKeeper types.StakingKeeper,
	deliverTx deliverTxfn, genesisState types.GenesisState,
	txEncodingConfig client.TxEncodingConfig,
) (validators []abci.ValidatorUpdate, err error) {
	// this line is used by starport scaffolding # genesis/module/init
	if len(genesisState.GenspTxs) > 0 {
		validators, err = DeliverGenTxs(ctx, genesisState.GenspTxs, stakingKeeper, deliverTx, txEncodingConfig)
	}
	return validators, err
}
