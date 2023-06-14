package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
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

	querySPReq := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPResp, err := s.Client.StorageProviderByOperatorAddress(ctx, &querySPReq)

	genSP := &sptypes.StorageProvider{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
		FundingAddress:  sp.FundingKey.GetAddr().String(),
		SealAddress:     sp.SealKey.GetAddr().String(),
		ApprovalAddress: sp.ApprovalKey.GetAddr().String(),
		GcAddress:       sp.GcKey.GetAddr().String(),
		Description: sptypes.Description{
			Moniker:  "sp0",
			Identity: "",
			Details:  "detail_sp0",
			Website:  "http://website",
		},
		Endpoint:     "http://127.0.0.1:9033",
		TotalDeposit: querySPResp.StorageProvider.TotalDeposit,
	}

	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, genSP)
}

func TestGenStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(GenStorageProviderTestSuite))
}
