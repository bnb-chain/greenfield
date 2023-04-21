package keeper_test

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

func (s *KeeperTestSuite) TestSetGetStorageProvider() {
	keeper := s.spKeeper
	ctx := s.ctx
	sp := &types.StorageProvider{}
	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)

	sp.OperatorAddress = spAcc.String()

	keeper.SetStorageProvider(ctx, sp)
	_, found := keeper.GetStorageProvider(ctx, spAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(s.T(), found, true)
}

// TestStorageProviderBasics tests GetStorageProvider, GetStorageProviderByFundingAddr,
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
		OperatorAddress: spAcc.String(),
		FundingAddress:  fundingAcc.String(),
		SealAddress:     sealAcc.String(),
		ApprovalAddress: approvalAcc.String(),
	}

	k.SetStorageProvider(ctx, sp)
	_, found := k.GetStorageProvider(ctx, spAcc)
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
