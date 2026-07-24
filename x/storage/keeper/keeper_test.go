package keeper_test

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/golang/mock/gomock"

	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	"github.com/bnb-chain/greenfield/x/storage/types"
	vgtypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

// TestDiscontinueObjectRejectsPendingSwapInSuccessor is a regression test for SRC-947:
// from the Sahel upgrade on, a pending swap-in reservation must NOT grant discontinue
// authority. Only the current on-chain family primary SP may discontinue objects.
func (s *TestSuite) TestDiscontinueObjectRejectsPendingSwapInSuccessor() {
	// Sahel-gated fix; Hulunbeier gates the vulnerable reservation-as-authorization branch.
	ctx := sdk.NewContext(s.ctx.MultiStore(), s.ctx.BlockHeader(), false, func(_ sdk.Context, name string) bool {
		return name == upgradetypes.Hulunbeier || name == upgradetypes.Sahel
	}, s.ctx.Logger())

	// Attacker SP: in-service, resolves from the gc address, but is NOT the family primary.
	gcAddr := sample.RandAccAddress()
	attackerSP := &sptypes.StorageProvider{Id: 200, Status: sptypes.STATUS_IN_SERVICE, GcAddress: gcAddr.String()}
	s.spKeeper.EXPECT().GetStorageProviderByGcAddr(gomock.Any(), gomock.Any()).Return(attackerSP, true).AnyTimes()

	bucketInfo := &types.BucketInfo{
		Owner:                      sample.RandAccAddress().String(),
		BucketName:                 "src947-bucket",
		Id:                         sdkmath.NewUint(1),
		BucketStatus:               types.BUCKET_STATUS_CREATED,
		GlobalVirtualGroupFamilyId: 5,
	}
	s.storageKeeper.StoreBucketInfo(s.ctx, bucketInfo)

	// The family's real primary SP is a different SP (Id 100).
	s.virtualGroupKeeper.EXPECT().GetGVGFamily(gomock.Any(), bucketInfo.GlobalVirtualGroupFamilyId).
		Return(&vgtypes.GlobalVirtualGroupFamily{Id: 5, PrimarySpId: 100}, true).AnyTimes()
	s.spKeeper.EXPECT().MustGetStorageProvider(gomock.Any(), uint32(100)).
		Return(&sptypes.StorageProvider{Id: 100, Status: sptypes.STATUS_IN_SERVICE}).AnyTimes()

	// Even with a matching pending swap-in reservation, the non-primary successor must be denied.
	s.virtualGroupKeeper.EXPECT().GetSwapInInfo(gomock.Any(), bucketInfo.GlobalVirtualGroupFamilyId, vgtypes.NoSpecifiedGVGId).
		Return(&vgtypes.SwapInInfo{TargetSpId: 100, SuccessorSpId: attackerSP.Id}, true).AnyTimes()

	err := s.storageKeeper.DiscontinueObject(ctx, gcAddr, bucketInfo.BucketName, []sdkmath.Uint{sdkmath.NewUint(1)}, "attack")
	s.Require().ErrorIs(err, types.ErrAccessDenied)
}

func (s *TestSuite) TestClearDiscontinueBucketCount() {
	acc1 := sample.RandAccAddress()
	s.storageKeeper.SetDiscontinueBucketCount(s.ctx, acc1, 1)

	count := s.storageKeeper.GetDiscontinueBucketCount(s.ctx, acc1)
	s.Require().Equal(uint64(1), count)

	s.storageKeeper.ClearDiscontinueBucketCount(s.ctx)

	count = s.storageKeeper.GetDiscontinueBucketCount(s.ctx, acc1)
	s.Require().Equal(uint64(0), count)
}

func (s *TestSuite) TestClearDiscontinueObjectCount() {
	acc1 := sample.RandAccAddress()
	s.storageKeeper.SetDiscontinueObjectCount(s.ctx, acc1, 1)

	count := s.storageKeeper.GetDiscontinueObjectCount(s.ctx, acc1)
	s.Require().Equal(uint64(1), count)

	s.storageKeeper.ClearDiscontinueObjectCount(s.ctx)

	count = s.storageKeeper.GetDiscontinueObjectCount(s.ctx, acc1)
	s.Require().Equal(uint64(0), count)
}
