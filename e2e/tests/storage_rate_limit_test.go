package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *StorageTestSuite) TestSetBucketRateLimit() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 10000000)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	queryQuotaUpdateTimeResponse, err := s.Client.QueryQuotaUpdateTime(ctx, &storagetypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.CreateAt, queryQuotaUpdateTimeResponse.UpdateAt)

	fmt.Printf("User: %s\n", s.User.GetAddr().String())
	fmt.Printf("queryHeadBucketResponse.BucketInfo.Owner: %s\n", queryHeadBucketResponse.BucketInfo.Owner)
	fmt.Printf("queryHeadBucketResponse.BucketInfo.PaymentAccount: %s\n", queryHeadBucketResponse.BucketInfo.PaymentAddress)

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(s.User.GetAddr(), s.User.GetAddr(), s.User.GetAddr(), bucketName, sdkmath.NewInt(1))
	s.SendTxBlock(s.User, msgSetBucketRateLimit)
}

func (s *StorageTestSuite) TestSetBucketRateLimitToZero() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 10000000)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	queryQuotaUpdateTimeResponse, err := s.Client.QueryQuotaUpdateTime(ctx, &storagetypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.CreateAt, queryQuotaUpdateTimeResponse.UpdateAt)

	fmt.Printf("User: %s\n", s.User.GetAddr().String())
	fmt.Printf("queryHeadBucketResponse.BucketInfo.Owner: %s\n", queryHeadBucketResponse.BucketInfo.Owner)
	fmt.Printf("queryHeadBucketResponse.BucketInfo.PaymentAccount: %s\n", queryHeadBucketResponse.BucketInfo.PaymentAddress)

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(s.User.GetAddr(), s.User.GetAddr(), s.User.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(s.User, msgSetBucketRateLimit)

	// CreateObject
	objectName := storageutils.GenRandomObjectName()
	// create test buffer
	var buffer bytes.Buffer
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	payloadSize := buffer.Len()
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgCreateObject, user, "greater than the flow rate limit")
}

// TestNotOwnerSetBucketRateLimit_Object
// 1. user create a bucket with 0 read quota
// 2. the payment account set the rate limit
// 3. user create an object in the bucket
// 4. the payment account set the rate limit to 0
// 5. user create an object in the bucket and it should fail
// 6. the payment account set the rate limit to a positive number
// 7. user create an object in the bucket and it should pass
func (s *StorageTestSuite) TestNotOwnerSetBucketRateLimit_Object() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	paymentAcc := s.GenAndChargeAccounts(1, 1000000)[0]

	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()

	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		paymentAcc.GetAddr(), math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, paymentAcc.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	queryQuotaUpdateTimeResponse, err := s.Client.QueryQuotaUpdateTime(ctx, &storagetypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.CreateAt, queryQuotaUpdateTimeResponse.UpdateAt)

	fmt.Printf("User: %s\n", s.User.GetAddr().String())
	fmt.Printf("queryHeadBucketResponse.BucketInfo.Owner: %s\n", queryHeadBucketResponse.BucketInfo.Owner)
	fmt.Printf("queryHeadBucketResponse.BucketInfo.PaymentAccount: %s\n", queryHeadBucketResponse.BucketInfo.PaymentAddress)

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// CreateObject
	objectName := storageutils.GenRandomObjectName()
	// create test buffer
	var buffer bytes.Buffer
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	payloadSize := buffer.Len()
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// CreateObject
	objectName = storageutils.GenRandomObjectName()
	msgCreateObject = storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgCreateObject, user, "greater than the flow rate limit")

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	objectName = storageutils.GenRandomObjectName()
	msgCreateObject = storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)
}

// TestNotOwnerSetBucketRateLimit_Bucket
// 1. user create a bucket with 0 read quota
// 2. the payment account set the rate limit
// 3. user update the read quota to a positive number
// 4. the payment account set the rate limit to 0
// 5. user update the read quota to a positive number and it should fail
// 6. the payment account set the rate limit to a positive number
// 7. user update the read quota to a positive number and it should pass
func (s *StorageTestSuite) TestNotOwnerSetBucketRateLimit_Bucket() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	paymentAcc := s.GenAndChargeAccounts(1, 1000000)[0]

	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()

	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		paymentAcc.GetAddr(), math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, paymentAcc.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	queryQuotaUpdateTimeResponse, err := s.Client.QueryQuotaUpdateTime(ctx, &storagetypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.CreateAt, queryQuotaUpdateTimeResponse.UpdateAt)

	fmt.Printf("User: %s\n", s.User.GetAddr().String())
	fmt.Printf("queryHeadBucketResponse.BucketInfo.Owner: %s\n", queryHeadBucketResponse.BucketInfo.Owner)
	fmt.Printf("queryHeadBucketResponse.BucketInfo.PaymentAccount: %s\n", queryHeadBucketResponse.BucketInfo.PaymentAddress)

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// UpdateBucketInfo
	var readQuota uint64 = 100
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgUpdateBucketInfo)
	s.Require().NoError(err)

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// CreateObject
	readQuota = 101
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "greater than the flow rate limit")
	s.Require().NoError(err)

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	readQuota = 102
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgUpdateBucketInfo)
	s.Require().NoError(err)
}

// TestNotOwnerSetBucketRateLimit_BucketPaymentAccount
// 1. user create a bucket with positive read quota
// 2. user set the rate limit to 0
// 3. update the payment account to another payment account, it should fail
// 4. the payment account set the rate limit to 0
// 5. user update the payment account to another payment account, it should fail
// 6. the payment account set the rate limit to a positive number
// 7. user update the payment account to another payment account, it should pass
func (s *StorageTestSuite) TestNotOwnerSetBucketRateLimit_BucketPaymentAccount() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	paymentAcc := s.GenAndChargeAccounts(1, 1000000)[0]

	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()

	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		user.GetAddr(), math.MaxUint, nil, 100)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, s.User.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	queryQuotaUpdateTimeResponse, err := s.Client.QueryQuotaUpdateTime(ctx, &storagetypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.CreateAt, queryQuotaUpdateTimeResponse.UpdateAt)

	fmt.Printf("User: %s\n", s.User.GetAddr().String())
	fmt.Printf("queryHeadBucketResponse.BucketInfo.Owner: %s\n", queryHeadBucketResponse.BucketInfo.Owner)
	fmt.Printf("queryHeadBucketResponse.BucketInfo.PaymentAccount: %s\n", queryHeadBucketResponse.BucketInfo.PaymentAddress)

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(s.User.GetAddr(), s.User.GetAddr(), s.User.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(s.User, msgSetBucketRateLimit)

	// SetBucketRateLimit
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAcc.GetAddr(), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "the flow rate limit is not set")
	s.Require().NoError(err)

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// UpdateBucketInfo
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAcc.GetAddr(), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "greater than the flow rate limit")
	s.Require().NoError(err)

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// UpdateBucketInfo
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAcc.GetAddr(), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgUpdateBucketInfo)
	s.Require().NoError(err)
}
