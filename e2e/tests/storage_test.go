package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/bnb-chain/greenfield/e2e/core"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

var TestBucket = "testbucket"

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
	bucketName := core.GenRandomBucketName()
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

func (s *StorageTestSuite) TestCreateObject() {
	var err error
	// CreateBucket
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := core.GenRandomBucketName()
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

	// CreateObject
	objectName := core.GenRandomObjectName()
	// create test buffer
	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	payloadSize := buffer.Len()
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), false, expectChecksum, contextType, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = s.StorageProvider.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.PayloadSize, uint64(payloadSize))
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.IsPublic, false)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_INIT)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)

	// SealObject
	secondarySPs := []sdk.AccAddress{
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.OperatorKey.GetAddr(),
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.OperatorKey.GetAddr(),
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.OperatorKey.GetAddr(),
	}
	msgSealObject := storagetypes.NewMsgSealObject(s.StorageProvider.SealKey.GetAddr(), bucketName, objectName, secondarySPs, nil)
	sr := storagetypes.NewSecondarySpSignDoc(s.StorageProvider.OperatorKey.GetAddr(), checksum)
	secondarySig, err := s.StorageProvider.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
	s.Require().NoError(err)
	err = storagetypes.VerifySignature(s.StorageProvider.ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()), secondarySig)
	s.Require().NoError(err)

	secondarySigs := [][]byte{secondarySig, secondarySig, secondarySig, secondarySig, secondarySig, secondarySig}
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(msgSealObject, s.StorageProvider.SealKey)

	// DeleteObject
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(msgDeleteObject, user)
	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(msgDeleteBucket, user)
}

func (s *StorageTestSuite) TestDeleteBucket() {
	var err error
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	// 1. CreateBucket1
	bucketName1 := core.GenRandomBucketName()
	msgCreateBucket1 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName1, false, s.StorageProvider.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket1.PrimarySpApproval.Sig, err = s.StorageProvider.ApprovalKey.GetPrivKey().Sign(msgCreateBucket1.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket1, user)

	// 2. CreateBucket1
	bucketName2 := core.GenRandomBucketName()
	msgCreateBucket2 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName2, false, s.StorageProvider.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket2.PrimarySpApproval.Sig, err = s.StorageProvider.ApprovalKey.GetPrivKey().Sign(msgCreateBucket2.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket2, user)

	// 3. Create object into bucket1
	// CreateObject
	objectName := core.GenRandomObjectName()
	// create test buffer
	var buffer bytes.Buffer
	line := `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	payloadSize := buffer.Len()
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName1, objectName, uint64(payloadSize),
		false, expectChecksum, contextType, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = s.StorageProvider.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user)

	// SealObject
	secondarySPs := []sdk.AccAddress{
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.OperatorKey.GetAddr(),
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.OperatorKey.GetAddr(),
		s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.OperatorKey.GetAddr(),
	}
	msgSealObject := storagetypes.NewMsgSealObject(s.StorageProvider.SealKey.GetAddr(), bucketName1, objectName,
		secondarySPs, nil)
	sr := storagetypes.NewSecondarySpSignDoc(s.StorageProvider.OperatorKey.GetAddr(), checksum)
	secondarySig, err := s.StorageProvider.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
	s.Require().NoError(err)
	err = storagetypes.VerifySignature(s.StorageProvider.ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()), secondarySig)
	s.Require().NoError(err)

	secondarySigs := [][]byte{secondarySig, secondarySig, secondarySig, secondarySig, secondarySig, secondarySig}
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(msgSealObject, s.StorageProvider.SealKey)

	// 4. Delete bucket2
	msgDeleteBucket2 := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName2)
	s.SendTxBlock(msgDeleteBucket2, user)

	// 5. Delete object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName1, objectName)
	s.SendTxBlock(msgDeleteObject, user)

	// 6. delete bucket1
	msgDeleteBucket1 := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName1)
	s.SendTxBlock(msgDeleteBucket1, user)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
