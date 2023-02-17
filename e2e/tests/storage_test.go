package tests

import (
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	"github.com/stretchr/testify/suite"
	"math"
	"testing"

	"github.com/bnb-chain/greenfield/e2e/core"
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

func (s *PaymentTestSuite) TestCreateBucket() {
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
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
