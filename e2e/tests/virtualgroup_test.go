package tests

import (
	"context"
	"testing"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/stretchr/testify/suite"
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

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroup(gvgID uint32) *virtualgroupmoduletypes.GlobalVirtualGroup {
	resp, err := s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroup
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroupByFamily(spID, familyID uint32) []*virtualgroupmoduletypes.GlobalVirtualGroup {
	resp, err := s.Client.GlobalVirtualGroupByFamilyID(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupByFamilyIDRequest{
			StorageProviderId:          spID,
			GlobalVirtualGroupFamilyId: familyID,
		})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroups
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroupFamilies(spID uint32) []*virtualgroupmoduletypes.GlobalVirtualGroupFamily {
	resp, err := s.Client.GlobalVirtualGroupFamilies(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupFamiliesRequest{StorageProviderId: spID})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroupFamilies
}

func (s *VirtualGroupTestSuite) TestBasic() {
	primarySP := s.StorageProviders[0]

	gvgFamilies := s.queryGlobalVirtualGroupFamilies(primarySP.Info.Id)
	s.Require().Greater(len(gvgFamilies), 0)

	family := gvgFamilies[0]
	s.T().Log(family.String())

	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySP.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
	}
	s.createGlobalVirtualGroup(primarySP, family.Id, secondarySPIDs, 1)

	gvgs := s.queryGlobalVirtualGroupByFamily(primarySP.Info.Id, family.Id)
	s.Require().Equal(len(gvgs), len(family.GlobalVirtualGroupIds)+1)

	oldGVGIDs := make(map[uint32]bool)
	for _, id := range family.GlobalVirtualGroupIds {
		oldGVGIDs[id] = true
	}
	var newGVG *virtualgroupmoduletypes.GlobalVirtualGroup

	for _, gvg := range gvgs {
		if !oldGVGIDs[gvg.Id] {
			newGVG = gvg
			break
		}
	}

	s.Require().Equal(newGVG.TotalDeposit.Int64(), int64(1000000000000000000))

	// test deposit
	msgDeposit := virtualgroupmoduletypes.MsgDeposit{
		FundingAddress:       primarySP.FundingKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
		Deposit:              sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(1, types.DecimalBNB)),
	}
	s.SendTxBlock(primarySP.FundingKey, &msgDeposit)

	gvgAfterDeposit := s.queryGlobalVirtualGroup(newGVG.Id)
	s.Require().Equal(gvgAfterDeposit.TotalDeposit.Int64(), int64(2000000000000000000))

	// test withdraw
	balance, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySP.FundingKey.GetAddr().String()})
	s.Require().NoError(err)

	msgWithdraw := virtualgroupmoduletypes.MsgWithdraw{
		FundingAddress:       primarySP.FundingKey.GetAddr().String(),
		Withdraw:             sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(1, types.DecimalBNB)),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.FundingKey, &msgWithdraw)
	balanceAfterWithdraw, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySP.FundingKey.GetAddr().String()})
	s.Require().NoError(err)

	s.T().Logf("balance: %s, after: %s", balance.String(), balanceAfterWithdraw.String())
	s.Require().Equal(balanceAfterWithdraw.Balance.Amount.Sub(balance.Balance.Amount).Int64(), int64(999994000000000000))

	// test delete gvg
	msgDeleteGVG := virtualgroupmoduletypes.MsgDeleteGlobalVirtualGroup{
		PrimarySpAddress:     primarySP.OperatorKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.OperatorKey, &msgDeleteGVG)

	newGVGs := s.queryGlobalVirtualGroupByFamily(primarySP.Info.Id, family.Id)

	for _, gvg := range newGVGs {
		if gvg.Id == newGVG.Id {
			s.Assert().True(false)
		}
	}
	_, err = s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: newGVG.Id})
	s.Require().Error(err)
}
