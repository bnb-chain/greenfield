package keeper_test

import (
	"testing"

	"cosmossdk.io/math"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"

	"github.com/bnb-chain/greenfield/testutil/sample"
	"github.com/bnb-chain/greenfield/x/challenge/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *TestSuite) TestSubmit() {
	existSpAddr := sample.RandAccAddress()
	existSp := &sptypes.StorageProvider{Status: sptypes.STATUS_IN_SERVICE}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(existSpAddr)).
		Return(existSp, true).AnyTimes()
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Any()).
		Return(nil, false).AnyTimes()

	existBucketName, existObjectName := "existbucket", "existobject"
	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		BucketName:   existBucketName,
		ObjectName:   existObjectName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfo(gomock.Any(), gomock.Eq(existBucketName), gomock.Eq(existObjectName)).
		Return(existObject, true).AnyTimes()
	s.storageKeeper.EXPECT().GetObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, false).AnyTimes()

	existBucket := &storagetypes.BucketInfo{
		BucketName:       existBucketName,
		PrimarySpAddress: existSpAddr.String()}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(existBucketName)).
		Return(existBucket, true).AnyTimes()

	s.storageKeeper.EXPECT().MaxSegmentSize(gomock.Any()).Return(uint64(10000)).AnyTimes()

	tests := []struct {
		name string
		msg  types.MsgSubmit
		err  error
	}{
		{
			name: "unknown sp",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
			},
			err: types.ErrUnknownSp,
		}, {
			name: "unknown object",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: existSpAddr.String(),
				ObjectName:        "nonexistobject",
			},
			err: types.ErrUnknownObject,
		},
		{
			name: "invalid segment index",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
				SegmentIndex:      10,
			},
			err: types.ErrInvalidSegmentIndex,
		},
		{
			name: "success with specific index",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
				SegmentIndex:      0,
			},
		}, {
			name: "success with random index",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
				RandomIndex:       true,
			},
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			_, err := s.msgServer.Submit(s.ctx, &tt.msg)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}

	// verify storage
	s.Require().Equal(uint64(2), s.challengeKeeper.GetChallengeCountCurrentBlock(s.ctx))
	s.Require().Equal(uint64(2), s.challengeKeeper.GetChallengeId(s.ctx))

	// create slash
	s.challengeKeeper.SaveSlash(s.ctx, types.Slash{
		SpOperatorAddress: existSpAddr,
		ObjectId:          existObject.Id,
		Height:            100,
	})

	tests = []struct {
		name string
		msg  types.MsgSubmit
		err  error
	}{
		{
			name: "failed due to recent slash",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
				RandomIndex:       true,
			},
			err: types.ErrExistsRecentSlash,
		},
	}
	for _, tt := range tests {
		s.T().Run(tt.name, func(t *testing.T) {
			_, err := s.msgServer.Submit(s.ctx, &tt.msg)
			if tt.err != nil {
				require.ErrorIs(t, err, tt.err)
				return
			}
			require.NoError(t, err)
		})
	}
}
