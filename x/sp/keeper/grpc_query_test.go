package keeper_test

import (
	gocontext "context"

	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *KeeperTestSuite) TestQueryParams() {
	res, err := s.queryClient.Params(gocontext.Background(), &types.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().Equal(s.spKeeper.GetParams(s.ctx), res.GetParams())
}
