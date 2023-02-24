package keeper_test

import (
	"context"
	"fmt"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/sp/keeper"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

// nolint
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.SpKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

func TestKeeper(t *testing.T) {
	k, ctx := keepertest.SpKeeper(t)
	sp := types.StorageProvider{}
	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)

	sp.OperatorAddress = spAcc.String()

	k.SetStorageProvider(ctx, sp)
	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(t, found, true)
}

// This function tests GetStorageProvider, GetStorageProviderByFundingAddr,
//  GetStorageProviderBySealAddr, GetStorageProviderByApprovalAddr
func TestStorageProviderBasics(t *testing.T) {
	k, ctx := keepertest.SpKeeper(t)

	spAccStr := sample.AccAddress()
	spAcc := sdk.MustAccAddressFromHex(spAccStr)

	fundingAccStr := sample.AccAddress()
	fundingAcc := sdk.MustAccAddressFromHex(fundingAccStr)

	sealAccStr := sample.AccAddress()
	sealAcc := sdk.MustAccAddressFromHex(sealAccStr)

	approvalAccStr := sample.AccAddress()
	approvalAcc := sdk.MustAccAddressFromHex(approvalAccStr)

	sp := types.StorageProvider{
		OperatorAddress: spAcc.String(),
		FundingAddress:  fundingAcc.String(),
		SealAddress:     sealAcc.String(),
		ApprovalAddress: approvalAcc.String(),
	}

	k.SetStorageProvider(ctx, sp)
	sp, found := k.GetStorageProvider(ctx, spAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(t, found, true)

	k.SetStorageProviderByFundingAddr(ctx, sp)
	sp, found = k.GetStorageProviderByFundingAddr(ctx, fundingAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(t, found, true)

	k.SetStorageProviderBySealAddr(ctx, sp)
	sp, found = k.GetStorageProviderBySealAddr(ctx, sealAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(t, found, true)

	k.SetStorageProviderByApprovalAddr(ctx, sp)
	sp, found = k.GetStorageProviderByApprovalAddr(ctx, approvalAcc)
	if !found {
		fmt.Printf("no such sp: %s", spAcc)
	}
	require.EqualValues(t, found, true)
}
