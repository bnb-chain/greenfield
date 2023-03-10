package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

type GenStorageProviderTestSuite struct {
	core.BaseSuite
}

func (s *GenStorageProviderTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

// gen storage provider in genesis
func (s *GenStorageProviderTestSuite) TestGenStorageProvider() {
	ctx := context.Background()

	sp := s.StorageProviders[0]

	genSP := &sptypes.StorageProvider{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
		FundingAddress:  sp.FundingKey.GetAddr().String(),
		SealAddress:     sp.SealKey.GetAddr().String(),
		ApprovalAddress: sp.ApprovalKey.GetAddr().String(),
		Description: sptypes.Description{
			Moniker:  "sp0",
			Identity: "",
			Details:  "detail_sp0",
			Website:  "http://website",
		},
		Endpoint:     "http://127.0.0.1:9033",
		TotalDeposit: types.NewIntFromInt64WithDecimal(10000000, types.DecimalBNB),
	}
	querySPReq := sptypes.QueryStorageProviderRequest{
		SpAddress: sp.OperatorKey.GetAddr().String(),
	}
	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, genSP)
}

func TestGenStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(GenStorageProviderTestSuite))
}
