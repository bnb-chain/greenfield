package tests

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type Eip712TestSuite struct {
	core.BaseSuite
}

func (s *Eip712TestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *Eip712TestSuite) SetupTest() {
}

func TestEip712TestSuite(t *testing.T) {
	suite.Run(t, new(Eip712TestSuite))
}

func (s *Eip712TestSuite) TestMultiMessages() {
	var err error
	sp := s.StorageProviders[0]
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)

	// UpdateBucketInfo
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().NoError(err)

	// send two messages together without error
	s.SendTxBlock(user, msgCreateBucket, msgUpdateBucketInfo)

	// verify modified bucketinfo
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponseAfterUpdateBucket, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponseAfterUpdateBucket.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
}
