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
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

func (s *TestSuite) TestSubmit() {
	existSpAddr := sample.RandAccAddress()
	existSp := &sptypes.StorageProvider{Status: sptypes.STATUS_IN_SERVICE, Id: 100, OperatorAddress: existSpAddr.String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(existSp.Id)).
		Return(existSp, true).AnyTimes()

	jailedSpAddr := sample.RandAccAddress()
	jailedSp := &sptypes.StorageProvider{Status: sptypes.STATUS_IN_JAILED, Id: 200, OperatorAddress: jailedSpAddr.String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(jailedSp.Id)).
		Return(jailedSp, true).AnyTimes()

	existBucketName, existObjectName := "existbucket", "existobject"
	existObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		BucketName:   existBucketName,
		ObjectName:   existObjectName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfo(gomock.Any(), gomock.Eq(existBucketName), gomock.Eq(existObjectName)).
		Return(existObject, true).AnyTimes()

	existBucket := &storagetypes.BucketInfo{
		BucketName: existBucketName,
	}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(existBucketName)).
		Return(existBucket, true).AnyTimes()
	s.storageKeeper.EXPECT().MustGetPrimarySPForBucket(gomock.Any(), gomock.Eq(existBucket)).Return(existSp).AnyTimes()

	jailedBucketName, jailedObjectName := "jailedbucket", "jailedobject"
	jailedObject := &storagetypes.ObjectInfo{
		Id:           math.NewUint(10),
		BucketName:   jailedBucketName,
		ObjectName:   jailedObjectName,
		ObjectStatus: storagetypes.OBJECT_STATUS_SEALED,
		PayloadSize:  500}
	s.storageKeeper.EXPECT().GetObjectInfo(gomock.Any(), gomock.Eq(jailedBucketName), gomock.Eq(jailedObjectName)).
		Return(jailedObject, true).AnyTimes()

	jailedBucket := &storagetypes.BucketInfo{
		BucketName: jailedBucketName,
	}
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Eq(jailedBucketName)).
		Return(jailedBucket, true).AnyTimes()
	s.storageKeeper.EXPECT().MustGetPrimarySPForBucket(gomock.Any(), gomock.Eq(jailedBucket)).Return(jailedSp).AnyTimes()

	s.storageKeeper.EXPECT().GetObjectInfo(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, false).AnyTimes()
	s.storageKeeper.EXPECT().GetBucketInfo(gomock.Any(), gomock.Any()).
		Return(nil, false).AnyTimes()

	s.storageKeeper.EXPECT().MaxSegmentSize(gomock.Any()).Return(uint64(10000)).AnyTimes()

	gvg := &virtualgrouptypes.GlobalVirtualGroup{PrimarySpId: 100, SecondarySpIds: []uint32{
		1,
	}}
	s.storageKeeper.EXPECT().GetObjectGVG(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(gvg, true).AnyTimes()

	secondarySpAddr := sample.RandAccAddress()
	secondarySp := &sptypes.StorageProvider{Status: sptypes.STATUS_IN_SERVICE, Id: 1, OperatorAddress: secondarySpAddr.String()}
	s.spKeeper.EXPECT().GetStorageProvider(gomock.Any(), gomock.Eq(secondarySp.Id)).
		Return(secondarySp, true).AnyTimes()

	tests := []struct {
		name string
		msg  types.MsgSubmit
		err  error
	}{
		{
			name: "incorrect sp status",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: sample.RandAccAddressHex(),
				BucketName:        jailedBucketName,
				ObjectName:        jailedObjectName,
			},
			err: types.ErrInvalidSpStatus,
		},
		{
			name: "not store on the sp",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: sample.RandAccAddressHex(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
			},
			err: types.ErrNotStoredOnSp,
		},
		{
			name: "unknown bucket",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        "unknownbucket",
				ObjectName:        "nonexistobject",
			},
			err: types.ErrUnknownBucketObject,
		},
		{
			name: "unknown object",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        "nonexistobject",
			},
			err: types.ErrUnknownBucketObject,
		},
		{
			name: "invalid segment index",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
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
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
				SegmentIndex:      0,
			},
		}, {
			name: "success with random index",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: existSpAddr.String(),
				BucketName:        existBucketName,
				ObjectName:        existObjectName,
				RandomIndex:       true,
			},
		}, {
			name: "success with secondary sp",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
				SpOperatorAddress: secondarySpAddr.String(),
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
	s.Require().Equal(uint64(3), s.challengeKeeper.GetChallengeCountCurrentBlock(s.ctx))
	s.Require().Equal(uint64(3), s.challengeKeeper.GetChallengeId(s.ctx))

	// create slash
	s.challengeKeeper.SaveSlash(s.ctx, types.Slash{
		SpId:     existSp.Id,
		ObjectId: existObject.Id,
		Height:   100,
	})

	tests = []struct {
		name string
		msg  types.MsgSubmit
		err  error
	}{
		{
			name: "failed due to recent slash",
			msg: types.MsgSubmit{
				Challenger:        sample.RandAccAddressHex(),
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
