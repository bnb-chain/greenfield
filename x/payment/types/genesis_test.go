package types_test

import (
	"testing"

	"github.com/bnb-chain/bfs/x/payment/types"
	"github.com/stretchr/testify/require"
)

func TestGenesisState_Validate(t *testing.T) {
	for _, tc := range []struct {
		desc     string
		genState *types.GenesisState
		valid    bool
	}{
		{
			desc:     "default is valid",
			genState: types.DefaultGenesis(),
			valid:    true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{

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
						To:   "0",
					},
					{
						From: "1",
						To:   "1",
					},
				},
				BnbPrice: &types.BnbPrice{
					Time:  87,
					Price: 30,
				},
				AutoSettleQueueList: []types.AutoSettleQueue{
					{
						Timestamp: 0,
						User:      "0",
					},
					{
						Timestamp: 1,
						User:      "1",
					},
				},
				MockObjectInfoList: []types.MockObjectInfo{
					{
						BucketName: "0",
						ObjectName: "0",
					},
					{
						BucketName: "1",
						ObjectName: "1",
					},
				},
				// this line is used by starport scaffolding # types/genesis/validField
			},
			valid: true,
		},
		{
			desc: "duplicated streamRecord",
			genState: &types.GenesisState{
				StreamRecordList: []types.StreamRecord{
					{
						Account: "0",
					},
					{
						Account: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated paymentAccountCount",
			genState: &types.GenesisState{
				PaymentAccountCountList: []types.PaymentAccountCount{
					{
						Owner: "0",
					},
					{
						Owner: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated paymentAccount",
			genState: &types.GenesisState{
				PaymentAccountList: []types.PaymentAccount{
					{
						Addr: "0",
					},
					{
						Addr: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated mockBucketMeta",
			genState: &types.GenesisState{
				MockBucketMetaList: []types.MockBucketMeta{
					{
						BucketName: "0",
					},
					{
						BucketName: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated mockBucketMeta",
			genState: &types.GenesisState{
				MockBucketMetaList: []types.MockBucketMeta{
					{
						BucketName: "0",
					},
					{
						BucketName: "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated flow",
			genState: &types.GenesisState{
				FlowList: []types.Flow{
					{
						From: "0",
						To:   "0",
					},
					{
						From: "0",
						To:   "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated autoSettleQueue",
			genState: &types.GenesisState{
				AutoSettleQueueList: []types.AutoSettleQueue{
					{
						Timestamp: 0,
						User:      "0",
					},
					{
						Timestamp: 0,
						User:      "0",
					},
				},
			},
			valid: false,
		},
		{
			desc: "duplicated mockObjectInfo",
			genState: &types.GenesisState{
				MockObjectInfoList: []types.MockObjectInfo{
					{
						BucketName: "0",
						ObjectName: "0",
					},
					{
						BucketName: "0",
						ObjectName: "0",
					},
				},
			},
			valid: false,
		},
		// this line is used by starport scaffolding # types/genesis/testcase
	} {
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
