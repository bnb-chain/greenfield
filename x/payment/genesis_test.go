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
	// this line is used by starport scaffolding # genesis/test/assert
}
