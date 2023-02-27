package keeper_test

import (
	"context"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"

	keepertest "github.com/bnb-chain/greenfield/testutil/keeper"
	"github.com/bnb-chain/greenfield/x/storage/keeper"
	"github.com/bnb-chain/greenfield/x/storage/types"
)

// nolint: unused
func setupMsgServer(t testing.TB) (types.MsgServer, context.Context) {
	k, ctx := keepertest.StorageKeeper(t)
	return keeper.NewMsgServerImpl(*k), sdk.WrapSDKContext(ctx)
}

//func TestCreateBucketChargeInitialReadFee(t *testing.T) {
//	k, ctx := keepertest.StorageKeeper(t)
//	msgServer := keeper.NewMsgServerImpl(*k)
//	goCtx := sdk.WrapSDKContext(ctx)
//	// create bucket
//	msgCreateBucket := types.NewMsgCreateBucket(
//		user.GetAddr(), bucketName, false, s.StorageProvider.OperatorKey.GetAddr(),
//		nil, math.MaxUint, nil)
//	msgCreateBucket.ReadQuota = bucketReadQuota
//	msgCreateBucket.PrimarySpApproval.Sig, err = s.StorageProvider.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
//	s.Require().NoError(err)
//	s.SendTxBlock(msgCreateBucket, user)
//
//}
