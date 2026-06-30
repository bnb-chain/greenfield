package keeper_test

import (
	"encoding/hex"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/sp/types"
)

// TestEditStorageProviderRevokesStaleOperationalIndexes verifies that rotating an
// SP's seal/gc/approval/bls identities through MsgEditStorageProvider revokes the
// old secondary-index entries, so a leaked or retired key no longer resolves to the
// SP in the privileged x/storage authorization paths.
func (s *KeeperTestSuite) TestEditStorageProviderRevokesStaleOperationalIndexes() {
	oldSeal := sample.RandAccAddress()
	oldApproval := sample.RandAccAddress()
	oldGC := sample.RandAccAddress()
	oldBlsKey, _ := sample.RandBlsPubKeyAndBlsProof()

	sp, err := types.NewStorageProvider(
		1,
		sample.RandAccAddress(),
		sample.RandAccAddress(),
		oldSeal,
		oldApproval,
		oldGC,
		sample.RandAccAddress(),
		sdkmath.NewInt(1),
		"https://sp.example",
		types.NewDescription("sp", "", "", ""),
		oldBlsKey,
	)
	s.Require().NoError(err)

	s.spKeeper.SetStorageProvider(s.ctx, &sp)
	s.spKeeper.SetStorageProviderByOperatorAddr(s.ctx, &sp)
	s.spKeeper.SetStorageProviderByFundingAddr(s.ctx, &sp)
	s.spKeeper.SetStorageProviderBySealAddr(s.ctx, &sp)
	s.spKeeper.SetStorageProviderByApprovalAddr(s.ctx, &sp)
	s.spKeeper.SetStorageProviderByGcAddr(s.ctx, &sp)
	s.spKeeper.SetStorageProviderByBlsKey(s.ctx, &sp)

	newSeal := sample.RandAccAddress()
	newApproval := sample.RandAccAddress()
	newGC := sample.RandAccAddress()
	newBlsKey, newBlsProof := sample.RandBlsPubKeyAndBlsProof()

	_, err = s.msgServer.EditStorageProvider(sdk.WrapSDKContext(s.ctx), &types.MsgEditStorageProvider{
		SpAddress:       sp.OperatorAddress,
		SealAddress:     newSeal.String(),
		ApprovalAddress: newApproval.String(),
		GcAddress:       newGC.String(),
		BlsKey:          newBlsKey,
		BlsProof:        newBlsProof,
	})
	s.Require().NoError(err)

	// Old identities must no longer resolve after rotation.
	_, found := s.spKeeper.GetStorageProviderBySealAddr(s.ctx, oldSeal)
	s.Require().False(found, "stale seal key should be revoked after rotation")
	_, found = s.spKeeper.GetStorageProviderByApprovalAddr(s.ctx, oldApproval)
	s.Require().False(found, "stale approval key should be revoked after rotation")
	_, found = s.spKeeper.GetStorageProviderByGcAddr(s.ctx, oldGC)
	s.Require().False(found, "stale gc key should be revoked after rotation")
	oldBlsBz, _ := hex.DecodeString(oldBlsKey)
	_, found = s.spKeeper.GetStorageProviderByBlsKey(s.ctx, oldBlsBz)
	s.Require().False(found, "stale bls key should be revoked after rotation")

	// New identities must resolve to the same SP.
	got, found := s.spKeeper.GetStorageProviderBySealAddr(s.ctx, newSeal)
	s.Require().True(found)
	s.Require().Equal(sp.Id, got.Id)
	got, found = s.spKeeper.GetStorageProviderByGcAddr(s.ctx, newGC)
	s.Require().True(found)
	s.Require().Equal(sp.Id, got.Id)
	newBlsBz, _ := hex.DecodeString(newBlsKey)
	got, found = s.spKeeper.GetStorageProviderByBlsKey(s.ctx, newBlsBz)
	s.Require().True(found)
	s.Require().Equal(sp.Id, got.Id)
}

// TestEditStorageProviderRejectsIdentityOwnedByAnotherSP verifies an SP cannot
// hijack another SP's operational identity through an edit.
func (s *KeeperTestSuite) TestEditStorageProviderRejectsIdentityOwnedByAnotherSP() {
	victimSeal := sample.RandAccAddress()
	victimBls, _ := sample.RandBlsPubKeyAndBlsProof()
	victim, err := types.NewStorageProvider(
		1, sample.RandAccAddress(), sample.RandAccAddress(),
		victimSeal, sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress(),
		sdkmath.NewInt(1), "https://victim.example", types.NewDescription("victim", "", "", ""), victimBls,
	)
	s.Require().NoError(err)
	s.spKeeper.SetStorageProvider(s.ctx, &victim)
	s.spKeeper.SetStorageProviderBySealAddr(s.ctx, &victim)

	attackerBls, _ := sample.RandBlsPubKeyAndBlsProof()
	attacker, err := types.NewStorageProvider(
		2, sample.RandAccAddress(), sample.RandAccAddress(),
		sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress(), sample.RandAccAddress(),
		sdkmath.NewInt(1), "https://attacker.example", types.NewDescription("attacker", "", "", ""), attackerBls,
	)
	s.Require().NoError(err)
	s.spKeeper.SetStorageProvider(s.ctx, &attacker)
	s.spKeeper.SetStorageProviderByOperatorAddr(s.ctx, &attacker)

	_, err = s.msgServer.EditStorageProvider(sdk.WrapSDKContext(s.ctx), &types.MsgEditStorageProvider{
		SpAddress:   attacker.OperatorAddress,
		SealAddress: victimSeal.String(),
	})
	s.Require().ErrorIs(err, types.ErrStorageProviderSealAddrExists)

	// Victim's seal index must still point to the victim.
	got, found := s.spKeeper.GetStorageProviderBySealAddr(s.ctx, victimSeal)
	s.Require().True(found)
	s.Require().Equal(victim.Id, got.Id)
}
