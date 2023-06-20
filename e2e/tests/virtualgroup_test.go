package tests

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type VirtualGroupTestSuite struct {
	core.BaseSuite
}

func (s *VirtualGroupTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *VirtualGroupTestSuite) SetupTest() {
}

func TestVirtualGroupTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualGroupTestSuite))
}

func (s *VirtualGroupTestSuite) createGlobalVirtualGroup(sp core.StorageProvider, familyID uint32, secondarySPIDs []uint32, depositAmount int64) {
	// Create a GVG for each sp by default
	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(depositAmount, types.DecimalBNB),
	}
	msgCreateGVG := &virtualgroupmoduletypes.MsgCreateGlobalVirtualGroup{
		PrimarySpAddress: sp.OperatorKey.GetAddr().String(),
		SecondarySpIds:   secondarySPIDs,
		Deposit:          deposit,
		FamilyId:         familyID,
	}
	s.SendTxBlock(sp.OperatorKey, msgCreateGVG)
}

func (s *VirtualGroupTestSuite) TestBasic() {
	primarySP := s.StorageProviders[0]

	resp, err := s.Client.GlobalVirtualGroupFamilies(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupFamiliesRequest{StorageProviderId: primarySP.Info.Id})
	s.Require().NoError(err)
	s.Require().Greater(len(resp.GlobalVirtualGroupFamilies), 0)

	family := resp.GlobalVirtualGroupFamilies[0]
	s.T().Log(family.String())

	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySP.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
	}
	s.createGlobalVirtualGroup(primarySP, family.Id, secondarySPIDs, 1)

	resp2, err := s.Client.GlobalVirtualGroupByFamilyID(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupByFamilyIDRequest{
			StorageProviderId:          primarySP.Info.Id,
			GlobalVirtualGroupFamilyId: family.Id,
		})
	s.Require().NoError(err)
	s.Require().Equal(len(resp2.GlobalVirtualGroups), len(family.GlobalVirtualGroupIds)+1)
	s.T().Log(resp2.String())

	oldGVGIDs := make(map[uint32]bool)
	for _, id := range family.GlobalVirtualGroupIds {
		oldGVGIDs[id] = true
	}
	var newGVG *virtualgroupmoduletypes.GlobalVirtualGroup
	for _, gvg := range resp2.GlobalVirtualGroups {
		if !oldGVGIDs[gvg.Id] {
			newGVG = gvg
			break
		}
	}

	msgDeleteGVG := virtualgroupmoduletypes.MsgDeleteGlobalVirtualGroup{
		PrimarySpAddress:     primarySP.OperatorKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.OperatorKey, &msgDeleteGVG)

	resp3, err := s.Client.GlobalVirtualGroupByFamilyID(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupByFamilyIDRequest{
			StorageProviderId:          primarySP.Info.Id,
			GlobalVirtualGroupFamilyId: family.Id,
		})
	s.Require().Error(err)
	for _, gvg := range resp3.GlobalVirtualGroups {
		if gvg.Id == newGVG.Id {
			s.Assert().True(false)
		}
	}

	_, err = s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: newGVG.Id,
	})
	s.T().Log(err)
	s.Require().Error(err)
}
