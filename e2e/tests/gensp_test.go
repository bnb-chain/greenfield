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

	sp := s.BaseSuite.PickStorageProvider()

	querySPReq := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPResp, err := s.Client.StorageProviderByOperatorAddress(ctx, &querySPReq)

	genSP := &sptypes.StorageProvider{
		OperatorAddress:    sp.OperatorKey.GetAddr().String(),
		FundingAddress:     sp.FundingKey.GetAddr().String(),
		SealAddress:        sp.SealKey.GetAddr().String(),
		ApprovalAddress:    sp.ApprovalKey.GetAddr().String(),
		GcAddress:          sp.GcKey.GetAddr().String(),
		MaintenanceAddress: sp.TestKey.GetAddr().String(),
		BlsKey:             sp.BlsKey.PubKey().Bytes(),
		Description: sptypes.Description{
			Moniker:  sp.Info.Description.Moniker,
			Identity: sp.Info.Description.Identity,
			Details:  sp.Info.Description.Details,
			Website:  sp.Info.Description.Website,
		},
		Endpoint:     sp.Info.Endpoint,
		TotalDeposit: querySPResp.StorageProvider.TotalDeposit,
	}

	s.Require().NoError(err)
	genSP.Id = querySPResp.StorageProvider.Id
	s.Require().Equal(querySPResp.StorageProvider, genSP)
}

func TestGenStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(GenStorageProviderTestSuite))
}
