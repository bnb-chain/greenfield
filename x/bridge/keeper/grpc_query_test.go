package keeper_test

import (
	gocontext "context"

	"github.com/bnb-chain/greenfield/x/bridge/types"
)

func (s *TestSuite) TestQueryParams() {
	res, err := s.queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(s.bridgeKeeper.GetParams(s.ctx), res.GetParams())
}
