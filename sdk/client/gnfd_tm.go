package client

import (
	"context"
	"encoding/hex"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
)

// GetBlock by height, gets the latest block if height is nil
func (c *GreenfieldClient) GetBlock(ctx context.Context, height *int64) (*ctypes.ResultBlock, error) {
	return c.tendermintClient.Block(ctx, height)
}

// Tx gets a tx by detail by the tx hash
func (c *GreenfieldClient) Tx(ctx context.Context, txHash string) (*ctypes.ResultTx, error) {
	hash, err := hex.DecodeString(txHash)
	if err != nil {
		return nil, err
	}
	return c.tendermintClient.Tx(ctx, hash, true)
}

// GetBlockResults by height, gets the latest block result if height is nil
func (c *GreenfieldClient) GetBlockResults(ctx context.Context, height *int64) (*ctypes.ResultBlockResults, error) {
	return c.tendermintClient.BlockResults(ctx, height)
}

// GetValidators by height, gets the latest validators if height is nil
func (c *GreenfieldClient) GetValidators(ctx context.Context, height *int64) (*ctypes.ResultValidators, error) {
	return c.tendermintClient.Validators(ctx, height, nil, nil)
}

// GetHeader by height, gets the latest block header if height is nil
func (c *GreenfieldClient) GetHeader(ctx context.Context, height *int64) (*ctypes.ResultHeader, error) {
	return c.tendermintClient.Header(ctx, height)
}

// GetUnconfirmedTxs by height, gets the latest block header if height is nil
func (c *GreenfieldClient) GetUnconfirmedTxs(ctx context.Context, limit *int) (*ctypes.ResultUnconfirmedTxs, error) {
	return c.tendermintClient.UnconfirmedTxs(ctx, limit)
}
