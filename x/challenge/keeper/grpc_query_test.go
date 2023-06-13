package keeper_test

import (
	"encoding/hex"
	"testing"

	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

func TestParamsQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	params := types.DefaultParams()
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	response, err := keeper.Params(ctx, &types.QueryParamsRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryParamsResponse{Params: params}, response)
}

func TestLatestAttestedChallengesQuery(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	err := keeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)
	c100 := &types.AttestedChallenge{Id: 100, Result: types.CHALLENGE_SUCCEED}
	c200 := &types.AttestedChallenge{Id: 200, Result: types.CHALLENGE_FAILED}
	keeper.AppendAttestedChallenge(ctx, c100)
	keeper.AppendAttestedChallenge(ctx, c200)

	response, err := keeper.LatestAttestedChallenges(ctx, &types.QueryLatestAttestedChallengesRequest{})
	require.NoError(t, err)
	require.Equal(t, &types.QueryLatestAttestedChallengesResponse{Challenges: []*types.AttestedChallenge{c100, c200}}, response)
}

func TestInturnAttestationSubmitterQuery(t *testing.T) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test")).Ctx

	ctrl := gomock.NewController(t)
	stakingKeeper := types.NewMockStakingKeeper(ctrl)

	blsKey := []byte("blskey")
	historicalInfo := stakingtypes.HistoricalInfo{
		Header: tmproto.Header{},
		Valset: []stakingtypes.Validator{{BlsKey: blsKey}},
	}
	stakingKeeper.EXPECT().GetHistoricalInfo(gomock.Any(), gomock.Any()).Return(historicalInfo, true).AnyTimes()

	keeper := keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		&types.MockBankKeeper{},
		&types.MockStorageKeeper{},
		&types.MockSpKeeper{},
		stakingKeeper,
		&types.MockPaymentKeeper{},
		&types.MockVirtualGroupKeeper{},
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)

	err := keeper.SetParams(ctx, types.DefaultParams())
	require.NoError(t, err)

	response, err := keeper.InturnAttestationSubmitter(ctx, &types.QueryInturnAttestationSubmitterRequest{})
	require.NoError(t, err)
	require.Equal(t, hex.EncodeToString(blsKey), response.BlsPubKey)
}
