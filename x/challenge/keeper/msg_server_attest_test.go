package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/golang/mock/gomock"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
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

	tests := []struct {
		name string
		msg  types.MsgAttest
		err  error
	}{
		{
			name: "unknown challenge",
			msg: types.MsgAttest{
				ChallengeId:       1,
				Submitter:         sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
			},
			err: types.ErrInvalidChallengeId,
		},
		{
			name: "not valid submitter",
			msg: types.MsgAttest{
				ChallengeId:       100,
				Submitter:         sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
			},
			err: types.ErrNotChallenger,
		},
		{
			name: "votes are not enough",
			msg: types.MsgAttest{
				ChallengeId:       100,
				Submitter:         validSubmitter.String(),
				SpOperatorAddress: sample.AccAddress(),
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
				SpOperatorAddress: sample.AccAddress(),
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

	existObjectName := "existobject"
	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		ObjectName:   existObjectName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(math.NewUint(10))).
		Return(existObject, true).AnyTimes()

	s.paymentKeeper.EXPECT().QueryDynamicBalance(gomock.Any(), gomock.Any()).
		Return(math.NewInt(1000000), nil).AnyTimes()
	s.paymentKeeper.EXPECT().Withdraw(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	spOperatorAcc := sample.RandAccAddress()
	attestMsg := &types.MsgAttest{
		Submitter:         validSubmitter.String(),
		ChallengeId:       challengeId,
		ObjectId:          math.NewUint(10),
		SpOperatorAddress: spOperatorAcc.String(),
		VoteResult:        types.CHALLENGE_FAILED,
		ChallengerAddress: "",
		VoteValidatorSet:  []uint64{1},
	}
	toSign := attestMsg.GetBlsSignBytes()

	voteAggSignature := blsKey.Sign(toSign[:])
	attestMsg.VoteAggSignature = voteAggSignature.Marshal()

	_, err := s.msgServer.Attest(s.ctx, attestMsg)
	require.NoError(s.T(), err)

	attestIds := s.challengeKeeper.GetAttestChallengeIds(s.ctx)
	s.Require().Contains(attestIds, challengeId)
}

func (s *TestSuite) TestAttest_Normal() {
	// prepare challenge
	challengeId := uint64(99)
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

	existObjectName := "existobject"
	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		ObjectName:   existObjectName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfoById(gomock.Any(), gomock.Eq(math.NewUint(10))).
		Return(existObject, true).AnyTimes()

	s.spKeeper.EXPECT().DepositDenomForSP(gomock.Any()).
		Return("BNB").AnyTimes()
	s.spKeeper.EXPECT().Slash(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil).AnyTimes()

	spOperatorAcc := sample.RandAccAddress()
	attestMsg := &types.MsgAttest{
		Submitter:         validSubmitter.String(),
		ChallengeId:       challengeId,
		ObjectId:          math.NewUint(10),
		SpOperatorAddress: spOperatorAcc.String(),
		VoteResult:        types.CHALLENGE_SUCCEED,
		ChallengerAddress: "",
		VoteValidatorSet:  []uint64{1},
	}
	toSign := attestMsg.GetBlsSignBytes()

	voteAggSignature := blsKey.Sign(toSign[:])
	attestMsg.VoteAggSignature = voteAggSignature.Marshal()

	_, err := s.msgServer.Attest(s.ctx, attestMsg)
	require.NoError(s.T(), err)

	attestIds := s.challengeKeeper.GetAttestChallengeIds(s.ctx)
	s.Require().Contains(attestIds, challengeId)
	s.Require().True(s.challengeKeeper.ExistsSlash(s.ctx, spOperatorAcc, attestMsg.ObjectId))
}
