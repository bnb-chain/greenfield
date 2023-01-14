package payment_test

import (
	"testing"

	keepertest "github.com/bnb-chain/bfs/testutil/keeper"
	"github.com/bnb-chain/bfs/testutil/nullify"
	"github.com/bnb-chain/bfs/x/payment"
	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/stretchr/testify/require"
)

func TestGenesis(t *testing.T) {
	genesisState := types.GenesisState{
		Params: types.DefaultParams(),

		StreamRecordList: []types.StreamRecord{
			{
				Account: "0",
			},
			{
				Account: "1",
			},
		},
		PaymentAccountCountList: []types.PaymentAccountCount{
			{
				Owner: "0",
			},
			{
				Owner: "1",
			},
		},
		PaymentAccountList: []types.PaymentAccount{
			{
				Addr: "0",
			},
			{
				Addr: "1",
			},
		},
		MockBucketMetaList: []types.MockBucketMeta{
		{
			BucketName: "0",
},
		{
			BucketName: "1",
},
	},
	MockBucketMetaList: []types.MockBucketMeta{
		{
			BucketName: "0",
},
		{
			BucketName: "1",
},
	},
	FlowList: []types.Flow{
		{
			From: "0",
To: "0",
},
		{
			From: "1",
To: "1",
},
	},
	BnbPrice: &types.BnbPrice{
		Time: 70,
Price: 63,
},
		AutoSettleQueueList: []types.AutoSettleQueue{
		{
			Timestamp: 0,
User: "0",
},
		{
			Timestamp: 1,
User: "1",
},
	},
	// this line is used by starport scaffolding # genesis/test/state
	}

	k, ctx := keepertest.PaymentKeeper(t)
	payment.InitGenesis(ctx, *k, genesisState)
	got := payment.ExportGenesis(ctx, *k)
	require.NotNil(t, got)

	nullify.Fill(&genesisState)
	nullify.Fill(got)

	require.ElementsMatch(t, genesisState.StreamRecordList, got.StreamRecordList)
	require.ElementsMatch(t, genesisState.PaymentAccountCountList, got.PaymentAccountCountList)
	require.ElementsMatch(t, genesisState.PaymentAccountList, got.PaymentAccountList)
	require.ElementsMatch(t, genesisState.MockBucketMetaList, got.MockBucketMetaList)
require.ElementsMatch(t, genesisState.MockBucketMetaList, got.MockBucketMetaList)
require.ElementsMatch(t, genesisState.FlowList, got.FlowList)
require.Equal(t, genesisState.BnbPrice, got.BnbPrice)
require.ElementsMatch(t, genesisState.AutoSettleQueueList, got.AutoSettleQueueList)
// this line is used by starport scaffolding # genesis/test/assert
}
