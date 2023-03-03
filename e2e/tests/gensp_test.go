package tests

import (
	"context"
	"testing"

	"github.com/bnb-chain/greenfield/sdk/types"

	"github.com/bnb-chain/greenfield/e2e/core"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/stretchr/testify/suite"
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

	genSP := &sptypes.StorageProvider{
		OperatorAddress: s.StorageProvider.OperatorKey.GetAddr().String(),
		FundingAddress:  s.StorageProvider.FundingKey.GetAddr().String(),
		SealAddress:     s.StorageProvider.SealKey.GetAddr().String(),
		ApprovalAddress: s.StorageProvider.ApprovalKey.GetAddr().String(),
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
		SpAddress: s.StorageProvider.OperatorKey.GetAddr().String(),
	}
	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, genSP)
}

func TestGenStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(GenStorageProviderTestSuite))
}
