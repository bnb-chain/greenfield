package client

import (
	"context"
	"github.com/stretchr/testify/assert"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/types"

	"github.com/bnb-chain/greenfield/sdk/client/test"
)

func TestGetABCIInfo(t *testing.T) {
	client := NewTendermintClient(test.TEST_RPC_ADDR)
	abci, err := client.TmClient.ABCIInfo(context.Background())
	assert.NoError(t, err)
	t.Log(abci.Response.LastBlockHeight)
}

func TestGetStatus(t *testing.T) {
	client := NewTendermintClient(test.TEST_RPC_ADDR)
	status, err := client.TmClient.Status(context.Background())
	assert.NoError(t, err)
	t.Log(status.ValidatorInfo)
}

func TestGetValidators(t *testing.T) {
	client := NewTendermintClient(test.TEST_RPC_ADDR)
	validators, err := client.TmClient.Validators(context.Background(), nil, nil, nil)
	assert.NoError(t, err)
	t.Log(validators.Validators)
}

func TestSubscribeEvent(t *testing.T) {
	const subscriber = "TestBlockEvents"
	client := NewTendermintClient(test.TEST_RPC_ADDR)
	err := client.TmClient.Start()
	require.NoError(t, err)
	eventCh, err := client.TmClient.Subscribe(context.Background(), subscriber, types.QueryForEvent(types.EventNewBlock).String())
	require.NoError(t, err)
	var firstBlockHeight int64
	for i := int64(0); i < 3; i++ {
		event := <-eventCh
		blockEvent, ok := event.Data.(types.EventDataNewBlock)
		require.True(t, ok)
		block := blockEvent.Block
		if firstBlockHeight == 0 {
			firstBlockHeight = block.Header.Height
		}
		require.Equal(t, firstBlockHeight+i, block.Header.Height)
	}
}
