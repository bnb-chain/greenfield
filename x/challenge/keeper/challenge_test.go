package keeper_test

import (
	"strconv"
	"testing"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	moduletestutil "github.com/cosmos/cosmos-sdk/types/module/testutil"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/mint"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/challenge/keeper"
	"github.com/bnb-chain/greenfield/x/challenge/types"
)

// Prevent strconv unused error
var _ = strconv.IntSize

func TestGetChallengeId(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	keeper.SaveChallenge(ctx, types.Challenge{
		Id:            100,
		ExpiredHeight: 1000,
	})
	require.True(t, keeper.GetChallengeId(ctx) == 100)
}

func TestAttestedChallenges(t *testing.T) {
	keeper, ctx := makeKeeper(t)
	params := types.DefaultParams()
	params.AttestationKeptCount = 5
	err := keeper.SetParams(ctx, params)
	require.NoError(t, err)

	c1 := &types.AttestedChallenge{Id: 1, Result: types.CHALLENGE_FAILED}
	c2 := &types.AttestedChallenge{Id: 2, Result: types.CHALLENGE_SUCCEED}
	c3 := &types.AttestedChallenge{Id: 3, Result: types.CHALLENGE_FAILED}

	keeper.AppendAttestedChallenge(ctx, c1)
	keeper.AppendAttestedChallenge(ctx, c2)
	keeper.AppendAttestedChallenge(ctx, c3)
	require.Equal(t, []*types.AttestedChallenge{c1, c2, c3}, keeper.GetAttestedChallenges(ctx))

	c4 := &types.AttestedChallenge{Id: 4, Result: types.CHALLENGE_FAILED}
	c5 := &types.AttestedChallenge{Id: 5, Result: types.CHALLENGE_FAILED}
	c6 := &types.AttestedChallenge{Id: 6, Result: types.CHALLENGE_SUCCEED}

	keeper.AppendAttestedChallenge(ctx, c4)
	keeper.AppendAttestedChallenge(ctx, c5)
	keeper.AppendAttestedChallenge(ctx, c6)
	require.Equal(t, []*types.AttestedChallenge{c2, c3, c4, c5, c6}, keeper.GetAttestedChallenges(ctx))

	params.AttestationKeptCount = 8
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)
	c7 := &types.AttestedChallenge{Id: 7, Result: types.CHALLENGE_FAILED}
	c8 := &types.AttestedChallenge{Id: 8, Result: types.CHALLENGE_SUCCEED}
	keeper.AppendAttestedChallenge(ctx, c7)
	keeper.AppendAttestedChallenge(ctx, c8)
	require.Equal(t, []*types.AttestedChallenge{c2, c3, c4, c5, c6, c7, c8}, keeper.GetAttestedChallenges(ctx))

	params.AttestationKeptCount = 3
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)
	c9 := &types.AttestedChallenge{Id: 9, Result: types.CHALLENGE_SUCCEED}
	keeper.AppendAttestedChallenge(ctx, c9)
	require.Equal(t, []*types.AttestedChallenge{c7, c8, c9}, keeper.GetAttestedChallenges(ctx))

	params.AttestationKeptCount = 5
	err = keeper.SetParams(ctx, params)
	require.NoError(t, err)
	c10 := &types.AttestedChallenge{Id: 10, Result: types.CHALLENGE_SUCCEED}
	keeper.AppendAttestedChallenge(ctx, c10)
	require.Equal(t, []*types.AttestedChallenge{c7, c8, c9, c10}, keeper.GetAttestedChallenges(ctx))
}

func makeKeeper(t *testing.T) (*keeper.Keeper, sdk.Context) {
	encCfg := moduletestutil.MakeTestEncodingConfig(mint.AppModuleBasic{})
	key := storetypes.NewKVStoreKey(types.StoreKey)
	testCtx := testutil.DefaultContextWithDB(t, key, storetypes.NewTransientStoreKey("transient_test"))

	k := keeper.NewKeeper(
		encCfg.Codec,
		key,
		key,
		&types.MockBankKeeper{},
		&types.MockStorageKeeper{},
		&types.MockSpKeeper{},
		&types.MockStakingKeeper{},
		&types.MockPaymentKeeper{},
		&types.MockVirtualGroupKeeper{},
		authtypes.NewModuleAddress(types.ModuleName).String(),
	)

	return k, testCtx.Ctx
}
