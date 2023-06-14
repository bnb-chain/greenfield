package keeper_test

import (
	"testing"

	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"

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
	existSp := &sptypes.StorageProvider{Status: sptypes.STATUS_IN_SERVICE, Id: 100, OperatorAddress: existSpAddr.String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(existSp.Id)).
		Return(existSp, true).AnyTimes()

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
		BucketName:  existBucketName,
		PrimarySpId: existSp.Id}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(existBucketName)).
		Return(existBucket, true).AnyTimes()
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Any()).
		Return(nil, false).AnyTimes()

	s.storageKeeper.EXPECT().MaxSegmentSize(gomock.Any()).Return(uint64(10000)).AnyTimes()

	lvg := &virtualgrouptypes.LocalVirtualGroup{}
	s.virtualGroupKeeper.EXPECT().GetLVG(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(lvg, true).AnyTimes()

	gvg := &virtualgrouptypes.GlobalVirtualGroup{PrimarySpId: 100}
	s.virtualGroupKeeper.EXPECT().GetGVG(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(gvg, true).AnyTimes()

	tests := []struct {
		name string
		msg  types.MsgSubmit
		err  error
	}{
		{
			name: "not store on the sp",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: sample.AccAddress(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
			},
			err: types.ErrNotStoredOnSp,
		}, {
			name: "unknown object",
			msg: types.MsgSubmit{
				Challenger:        sample.AccAddress(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
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
