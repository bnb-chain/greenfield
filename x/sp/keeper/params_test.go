package keeper_test

import (
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *KeeperTestSuite) TestGetParams() {
	k := s.spKeeper
	ctx := s.ctx
	params := types.DefaultParams()

	k.SetParams(ctx, params)

	require.EqualValues(s.T(), params, k.GetParams(ctx))
}
