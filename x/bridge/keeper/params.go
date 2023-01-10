package keeper

import (
	"fmt"
	"math/big"

	"github.com/bnb-chain/bfs/x/bridge/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramstore.GetParamSet(ctx, &params)
	return params
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

// GetTransferOutRelayerFee gets the transfer out relayer syn fee param
func (k Keeper) GetTransferOutRelayerFee(ctx sdk.Context) (*big.Int, *big.Int) {
	var relayerSynFeeParam, relayerAckFeeParam string

	k.paramstore.Get(ctx, types.KeyParamTransferOutRelayerSynFee, &relayerSynFeeParam)
	relayerSynFee, valid := big.NewInt(0).SetString(relayerSynFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerSynFeeParam))
	}

	k.paramstore.Get(ctx, types.KeyParamTransferOutRelayerAckFee, &relayerAckFeeParam)
	relayerAckFee, valid := big.NewInt(0).SetString(relayerAckFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerAckFeeParam))
	}

	return relayerSynFee, relayerAckFee
}
