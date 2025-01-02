package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (s *TestSuite) TestAttest_Invalid() {
	// prepare challenge
	s.challengeKeeper.SaveChallenge(s.ctx, types.Challenge{
		Id: 100,
	})

	validSubmitter := sample.RandAccAddress()

	blsKey, _ := bls.RandKey()
	historicalInfo := stakingtypes.HistoricalInfo{
		Header: tmproto.Header{},
		Valset: []stakingtypes.Validator{{
			BlsKey:            blsKey.PublicKey().Marshal(),
			ChallengerAddress: validSubmitter.String(),
		}},
	}
	s.stakingKeeper.EXPECT().GetHistoricalInfo(gomock.Any(), gomock.Any()).
		Return(historicalInfo, true).AnyTimes()

	existObjectName := "existobject"
	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		ObjectName:   existObjectName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(math.NewUint(10))).
		Return(existObject, true).AnyTimes()

	spOperatorAcc := sample.RandAccAddress()
	sp := &sptypes.StorageProvider{Id: 10, OperatorAddress: spOperatorAcc.String()}
	s.spKeeper.EXPECT().GetStorageProviderByOperatorAddr(gomock.Any(), gomock.Any()).
		Return(sp, true).AnyTimes()

	tests := []struct {
		name string
		msg  types.MsgAttest
		err  error
	}{
		{
			name: "unknown challenge",
			msg: types.MsgAttest{
				ChallengeId:       1,
				Submitter:         sample.RandAccAddressHex(),
				SpOperatorAddress: sample.RandAccAddressHex(),
			},
			err: types.ErrInvalidChallengeId,
		},
		{
			name: "not valid submitter",
			msg: types.MsgAttest{
				ChallengeId:       100,
				Submitter:         sample.RandAccAddressHex(),
				SpOperatorAddress: sample.RandAccAddressHex(),
			},
			err: types.ErrNotChallenger,
		},
		{
			name: "votes are not enough",
			msg: types.MsgAttest{
				ChallengeId:       100,
				Submitter:         validSubmitter.String(),
				SpOperatorAddress: sample.RandAccAddressHex(),
				ObjectId:          math.NewUint(10),
				VoteValidatorSet:  []uint64{},
				VoteAggSignature:  []byte{},
			},
			err: types.ErrNotEnoughVotes,
		},
		{
			name: "invalid signature",
			msg: types.MsgAttest{
				ChallengeId:       100,
				Submitter:         validSubmitter.String(),
				SpOperatorAddress: sample.RandAccAddressHex(),
				ObjectId:          math.NewUint(10),
				VoteValidatorSet:  []uint64{1},
				VoteAggSignature:  []byte{},
			},
			err: types.ErrInvalidVoteAggSignature,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			_, err := s.msgServer.Attest(s.ctx, &tt.msg)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func (s *TestSuite) TestAttest_Heartbeat() {
	// prepare challenge
	challengeId := s.challengeKeeper.GetParams(s.ctx).HeartbeatInterval
	s.challengeKeeper.SaveChallenge(s.ctx, types.Challenge{
		Id: challengeId,
	})

	validSubmitter := sample.RandAccAddress()

	blsKey, _ := bls.RandKey()
	historicalInfo := stakingtypes.HistoricalInfo{
		Header: tmproto.Header{},
		Valset: []stakingtypes.Validator{{
			BlsKey:            blsKey.PublicKey().Marshal(),
			ChallengerAddress: validSubmitter.String(),
		}},
	}
	s.stakingKeeper.EXPECT().GetHistoricalInfo(gomock.Any(), gomock.Any()).
		Return(historicalInfo, true).AnyTimes()

	existBucket := &storagetypes.BucketInfo{
		Id:                         math.NewUint(10),
		GlobalVirtualGroupFamilyId: 10,
		BucketName:                 "existbucket",
	}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(existBucket.BucketName)).
		Return(existBucket, true).AnyTimes()

	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		ObjectName:   "existobject",
		BucketName:   existBucket.BucketName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(math.NewUint(10))).
		Return(existObject, true).AnyTimes()

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1000000), nil).AnyTimes()
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	spOperatorAcc := sample.RandAccAddress()
	sp := &sptypes.StorageProvider{Id: 10, OperatorAddress: spOperatorAcc.String()}

	s.spKeeper.EXPECT().GetStorageProviderByOperatorAddr(gomock.Any(), gomock.Any()).
		Return(sp, true).AnyTimes()

	s.storageKeeper.EXPECT().MustGetPrimarySPForBucket(gomock.Any(), gomock.Any()).Return(sp).AnyTimes()

	gvg := &virtualgrouptypes.GlobalVirtualGroup{
		SecondarySpIds: []uint32{10},
	}
	s.storageKeeper.EXPECT().GetObjectGVG(gomock.Any(), gomock.Eq(existBucket.Id), gomock.Any()).
		Return(gvg, true).AnyTimes()

	attestMsg := &types.MsgAttest{
		Submitter:         validSubmitter.String(),
		ChallengeId:       challengeId,
		ObjectId:          math.NewUint(10),
		SpOperatorAddress: sp.OperatorAddress,
		VoteResult:        types.CHALLENGE_FAILED,
		ChallengerAddress: "",
		VoteValidatorSet:  []uint64{1},
	}
	toSign := attestMsg.GetBlsSignBytes(s.ctx.ChainID())

	voteAggSignature := blsKey.Sign(toSign[:])
	attestMsg.VoteAggSignature = voteAggSignature.Marshal()

	_, err := s.msgServer.Attest(s.ctx, attestMsg)
	require.NoError(s.T(), err)

	attestedChallenges := s.challengeKeeper.GetAttestedChallenges(s.ctx)
	found := false
	for _, c := range attestedChallenges {
		if c.Id == challengeId {
			found = true
		}
	}
	s.Require().True(found)
}

func (s *TestSuite) TestAttest_Normal() {
	// prepare challenge
	challenge1Id := uint64(99)
	s.challengeKeeper.SaveChallenge(s.ctx, types.Challenge{
		Id: challenge1Id,
	})
	challenge2Id := uint64(100)
	s.challengeKeeper.SaveChallenge(s.ctx, types.Challenge{
		Id: challenge2Id,
	})

	validSubmitter := sample.RandAccAddress()

	blsKey, _ := bls.RandKey()
	historicalInfo := stakingtypes.HistoricalInfo{
		Header: tmproto.Header{},
		Valset: []stakingtypes.Validator{{
			BlsKey:            blsKey.PublicKey().Marshal(),
			ChallengerAddress: validSubmitter.String(),
		}},
	}
	s.stakingKeeper.EXPECT().GetHistoricalInfo(gomock.Any(), gomock.Any()).
		Return(historicalInfo, true).AnyTimes()

	existBucket := &storagetypes.BucketInfo{
		Id:         math.NewUint(10),
		BucketName: "existbucket",
	}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(existBucket.BucketName)).
		Return(existBucket, true).AnyTimes()

	existObject1 := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		ObjectName:   "existobject1",
		BucketName:   existBucket.BucketName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(math.NewUint(10))).
		Return(existObject1, true).AnyTimes()

	existObject2 := &storagetypes.ObjectInfo{
		Id:           math.NewUint(100),
		ObjectName:   "existobject2",
		BucketName:   existBucket.BucketName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(math.NewUint(100))).
		Return(existObject2, true).AnyTimes()

	spOperatorAcc := sample.RandAccAddress()
	sp := &sptypes.StorageProvider{Id: 1, OperatorAddress: spOperatorAcc.String()}
	s.spKeeper.EXPECT().DepositDenomForSP(gomock.Any()).
		Return("BNB").AnyTimes()
	s.spKeeper.EXPECT().Slash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()
	s.spKeeper.EXPECT().GetStorageProviderByOperatorAddr(gomock.Any(), gomock.Any()).
		Return(sp, true).AnyTimes()
	s.storageKeeper.EXPECT().MustGetPrimarySPForBucket(gomock.Any(), gomock.Any()).Return(sp).AnyTimes()

	// success attestation
	attestMsg1 := &types.MsgAttest{
		Submitter:         validSubmitter.String(),
		ChallengeId:       challenge1Id,
		ObjectId:          math.NewUint(10),
		SpOperatorAddress: spOperatorAcc.String(),
		VoteResult:        types.CHALLENGE_SUCCEED,
		ChallengerAddress: "",
		VoteValidatorSet:  []uint64{1},
	}
	toSign1 := attestMsg1.GetBlsSignBytes(s.ctx.ChainID())
	voteAggSignature1 := blsKey.Sign(toSign1[:])
	attestMsg1.VoteAggSignature = voteAggSignature1.Marshal()
	_, err := s.msgServer.Attest(s.ctx, attestMsg1)
	require.NoError(s.T(), err)

	attestedChallenges := s.challengeKeeper.GetAttestedChallenges(s.ctx)
	attest1Found := false
	for _, c := range attestedChallenges {
		if c.Id == challenge1Id {
			attest1Found = true
		}
	}
	s.Require().True(attest1Found)
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, sp.Id, attestMsg1.ObjectId))

	// success attestation even exceed the max slash amount
	params := s.challengeKeeper.GetParams(s.ctx)
	params.SpSlashMaxAmount = math.NewInt(1)
	_ = s.challengeKeeper.SetParams(s.ctx, params)

	attestMsg2 := &types.MsgAttest{
		Submitter:         validSubmitter.String(),
		ChallengeId:       challenge2Id,
		ObjectId:          math.NewUint(100),
		SpOperatorAddress: spOperatorAcc.String(),
		VoteResult:        types.CHALLENGE_SUCCEED,
		ChallengerAddress: sample.RandAccAddress().String(),
		VoteValidatorSet:  []uint64{1},
	}
	toSign2 := attestMsg2.GetBlsSignBytes(s.ctx.ChainID())
	voteAggSignature2 := blsKey.Sign(toSign2[:])
	attestMsg2.VoteAggSignature = voteAggSignature2.Marshal()
	_, err = s.msgServer.Attest(s.ctx, attestMsg2)
	require.NoError(s.T(), err)

	attestedChallenges = s.challengeKeeper.GetAttestedChallenges(s.ctx)
	attest2Found := false
	for _, c := range attestedChallenges {
		if c.Id == challenge1Id {
			attest2Found = true
		}
	}
	s.Require().True(attest1Found)
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, sp.Id, attestMsg1.ObjectId))
	s.Require().True(attest2Found)
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, sp.Id, attestMsg2.ObjectId))

	// the sp and the object had been slashed
	attestMsg3 := &types.MsgAttest{
		Submitter:         validSubmitter.String(),
		ChallengeId:       challenge2Id,
		ObjectId:          math.NewUint(100),
		SpOperatorAddress: spOperatorAcc.String(),
		VoteResult:        types.CHALLENGE_SUCCEED,
		ChallengerAddress: sample.RandAccAddress().String(),
		VoteValidatorSet:  []uint64{1},
	}
	toSign3 := attestMsg3.GetBlsSignBytes(s.ctx.ChainID())
	voteAggSignature3 := blsKey.Sign(toSign3[:])
	attestMsg3.VoteAggSignature = voteAggSignature3.Marshal()
	_, err = s.msgServer.Attest(s.ctx, attestMsg3)
	require.Error(s.T(), err)
}
