package keeper_test

import (
	"fmt"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	types2 "github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *KeeperTestSuite) TestSetGetStorageProvider() {
	keeper := s.spKeeper
	ctx := s.ctx
	sp := &types.StorageProvider{Id: 100}
	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)
	sp.OperatorAddress = spAcc.String()

	keeper.SetStorageProvider(ctx, sp)
	_, found := keeper.GetStorageProvider(ctx, 100)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)
}

// TestStorageProviderBasics tests GetStorageProviderByOperatorAddr, GetStorageProviderByFundingAddr,
// GetStorageProviderBySealAddr, GetStorageProviderByApprovalAddr
func (s *KeeperTestSuite) TestStorageProviderBasics() {
	k := s.spKeeper
	ctx := s.ctx
	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)

	fundingAccStr := sample.AccAddress()
	fundingAcc := sdk.MustAccAddressFromHex(fundingAccStr)

	sealAccStr := sample.AccAddress()
	sealAcc := sdk.MustAccAddressFromHex(sealAccStr)

	approvalAccStr := sample.AccAddress()
	approvalAcc := sdk.MustAccAddressFromHex(approvalAccStr)

	sp := &types.StorageProvider{
		Id:              100,
		OperatorAddress: spAcc.String(),
		FundingAddress:  fundingAcc.String(),
		SealAddress:     sealAcc.String(),
		ApprovalAddress: approvalAcc.String(),
	}

	k.SetStorageProvider(ctx, sp)
	_, found := k.GetStorageProvider(ctx, 100)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)

	k.SetStorageProviderByFundingAddr(ctx, sp)
	_, found = k.GetStorageProviderByFundingAddr(ctx, fundingAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)

	k.SetStorageProviderBySealAddr(ctx, sp)
	_, found = k.GetStorageProviderBySealAddr(ctx, sealAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)

	k.SetStorageProviderByApprovalAddr(ctx, sp)
	_, found = k.GetStorageProviderByApprovalAddr(ctx, approvalAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)
}

func (s *KeeperTestSuite) TestSlashBasic() {
	// mock
	s.bankKeeper.EXPECT().SendCoinsFromModuleToAccount(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil)

	k := s.spKeeper
	ctx := s.ctx
	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)

	fundingAccStr := sample.AccAddress()
	fundingAcc := sdk.MustAccAddressFromHex(fundingAccStr)

	sealAccStr := sample.AccAddress()
	sealAcc := sdk.MustAccAddressFromHex(sealAccStr)

	approvalAccStr := sample.AccAddress()
	approvalAcc := sdk.MustAccAddressFromHex(approvalAccStr)

	sp := &types.StorageProvider{
		Id:              100,
		OperatorAddress: spAcc.String(),
		FundingAddress:  fundingAcc.String(),
		SealAddress:     sealAcc.String(),
		ApprovalAddress: approvalAcc.String(),
		TotalDeposit:    math.NewIntWithDecimal(2010, types2.DecimalBNB),
	}

	k.SetStorageProvider(ctx, sp)
	_, found := k.GetStorageProvider(ctx, 100)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)

	rewardInfo := types.RewardInfo{
		Address: sample.AccAddress(),
		Amount:  sdk.NewCoin(types2.Denom, math.NewIntWithDecimal(10, types2.DecimalBNB)),
	}

	err := k.Slash(ctx, spAcc, []types.RewardInfo{rewardInfo})
	require.NoError(s.T(), err)

	spAfterSlash, found := k.GetStorageProvider(ctx, 100)
	require.True(s.T(), found)
	s.T().Logf("%s", spAfterSlash.TotalDeposit.String())
	require.True(s.T(), spAfterSlash.TotalDeposit.Equal(math.NewIntWithDecimal(2000, types2.DecimalBNB)))
}
