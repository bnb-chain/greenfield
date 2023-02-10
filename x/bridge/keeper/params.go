package keeper

import (
	"fmt"
	"math/big"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/x/bridge/types"
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

// GetTransferOutRelayerFee gets the transfer out relayer fee params
func (k Keeper) GetTransferOutRelayerFee(ctx sdk.Context) (*big.Int, *big.Int) {
	var relayerFeeParam, ackRelayerFeeParam string

	k.paramstore.Get(ctx, types.KeyParamTransferOutRelayerFee, &relayerFeeParam)
	relayerFee, valid := big.NewInt(0).SetString(relayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid relayer fee: %s", relayerFeeParam))
	}

	k.paramstore.Get(ctx, types.KeyParamTransferOutAckRelayerFee, &ackRelayerFeeParam)
	ackRelayerFee, valid := big.NewInt(0).SetString(ackRelayerFeeParam, 10)
	if !valid {
		panic(fmt.Sprintf("invalid ack relayer fee: %s", ackRelayerFeeParam))
	}

	return relayerFee, ackRelayerFee
}
