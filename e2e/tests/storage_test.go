package tests

import (
	"context"
	"math"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

var (
	TestBucket = "testbucket"
)

type StorageTestSuite struct {
	core.BaseSuite
}

func (s *StorageTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageTestSuite) SetupTest() {
}

func (s *StorageTestSuite) TestCreateBucket() {
	var err error
	// CreateBucket
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := core.GetRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, false, s.StorageProvider.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = s.StorageProvider.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, s.StorageProvider.OperatorKey.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.IsPublic, false)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(msgDeleteBucket, user)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
