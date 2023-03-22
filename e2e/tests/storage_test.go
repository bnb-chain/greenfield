package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type StorageTestSuite struct {
	core.BaseSuite
	User keys.KeyManager
}

type StreamRecords struct {
	User paymenttypes.StreamRecord
	SPs  []paymenttypes.StreamRecord
}

func (s *StorageTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageTestSuite) SetupTest() {
	s.User = s.GenAndChargeAccounts(1, 1000000)[0]
}

func (s *StorageTestSuite) TestCreateBucket() {
	var err error
	sp := s.StorageProviders[0]
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// UpdateBucketInfo
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, math.MaxUint64, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().NoError(err)
	s.SendTxBlock(msgUpdateBucketInfo, user)
	s.Require().NoError(err)

	// verify modified bucketinfo
	queryHeadBucketResponseAfterUpdateBucket, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponseAfterUpdateBucket.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(msgDeleteBucket, user)
}

func (s *StorageTestSuite) TestCreateObject() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// CreateObject
	objectName := storageutils.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)

	// SealObject
	secondarySPs := []sdk.AccAddress{
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
	}
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, secondarySPs, nil)
	sr := storagetypes.NewSecondarySpSignDoc(sp.OperatorKey.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id, checksum)
	secondarySig, err := sp.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
	s.Require().NoError(err)
	err = storagetypes.VerifySignature(s.StorageProviders[0].ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()),
		secondarySig)
	s.Require().NoError(err)

	secondarySigs := [][]byte{secondarySig, secondarySig, secondarySig, secondarySig, secondarySig, secondarySig}
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(msgSealObject, sp.SealKey)

	// ListBuckets
	queryListBucketsRequest := storagetypes.QueryListBucketsRequest{}
	queryListBucketResponse, err := s.Client.ListBuckets(ctx, &queryListBucketsRequest)
	s.Require().NoError(err)
	s.Require().Greater(len(queryListBucketResponse.BucketInfos), 0)

	// ListObject
	queryListObjectsRequest := storagetypes.QueryListObjectsRequest{
		BucketName: bucketName,
	}
	queryListObjectsResponse, err := s.Client.ListObjects(ctx, &queryListObjectsRequest)
	s.Require().NoError(err)
	s.Require().Equal(len(queryListObjectsResponse.ObjectInfos), 1)
	s.Require().Equal(queryListObjectsResponse.ObjectInfos[0].ObjectName, objectName)

	// DeleteObject
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(msgDeleteObject, user)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(msgDeleteBucket, user)
}

func (s *StorageTestSuite) TestCreateGroup() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()})
	s.SendTxBlock(msgCreateGroup, owner)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())

	// 3. HeadGroupMember
	queryHeadGroupMemberReq := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	queryHeadGroupMemberResp, err := s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupMemberResp.GroupMember.GroupId, queryHeadGroupResp.GroupInfo.Id)

	// 4. UpdateGroupMember
	member2 := s.GenAndChargeAccounts(1, 1000000)[0]
	membersToAdd := []sdk.AccAddress{member2.GetAddr()}
	membersToDelete := []sdk.AccAddress{member.GetAddr()}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(msgUpdateGroupMember, owner)

	// 5. HeadGroupMember (delete)
	queryHeadGroupMemberReqDelete := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	_, err = s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReqDelete)
	s.Require().True(strings.Contains(err.Error(), storagetypes.ErrNoSuchGroupMember.Error()))
	// 5. HeadGroupMember (add)
	queryHeadGroupMemberReqAdd := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member2.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	queryHeadGroupMemberRespAdd, err := s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReqAdd)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupMemberRespAdd.GroupMember.GroupId, queryHeadGroupResp.GroupInfo.Id)

	// 6. Create a group with the same name
	msgCreateGroup = storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()})
	s.SendTxBlockWithExpectErrorString(msgCreateGroup, owner, "exists")
}

func (s *StorageTestSuite) TestDeleteBucket() {
	var err error
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	sp := s.StorageProviders[0]
	// 1. CreateBucket1
	bucketName1 := storageutils.GenRandomBucketName()
	msgCreateBucket1 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName1, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket1.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket1.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket1, user)

	// 2. CreateBucket2
	bucketName2 := storageutils.GenRandomBucketName()
	msgCreateBucket2 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName2, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket2.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket2.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket2, user)

	// 3. Create object into bucket1
	// CreateObject
	objectName := storageutils.GenRandomObjectName()
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
		storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user)

	// head object
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName1,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.T().Logf("queryHeadObjectResponse %s, err: %v", queryHeadObjectResponse, err)
	s.Require().NoError(err)

	// SealObject
	secondarySPs := []sdk.AccAddress{
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
	}

	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName1, objectName,
		secondarySPs, nil)
	sr := storagetypes.NewSecondarySpSignDoc(sp.OperatorKey.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id, checksum)
	secondarySig, err := sp.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
	s.Require().NoError(err)
	err = storagetypes.VerifySignature(sp.ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()), secondarySig)
	s.Require().NoError(err)

	secondarySigs := [][]byte{secondarySig, secondarySig, secondarySig, secondarySig, secondarySig, secondarySig}
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(msgSealObject, sp.SealKey)

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

func (s *StorageTestSuite) GetStreamRecord(addr string) (sr paymenttypes.StreamRecord) {
	ctx := context.Background()
	streamRecordResp, err := s.Client.StreamRecord(ctx, &paymenttypes.QueryGetStreamRecordRequest{
		Account: addr,
	})
	if streamRecordResp != nil {
		s.Require().NoError(err)
		sr = streamRecordResp.StreamRecord
	} else {
		s.Require().ErrorContainsf(err, "not found", "account: %s", addr)
		sr.StaticBalance = sdk.ZeroInt()
		sr.BufferBalance = sdk.ZeroInt()
		sr.LockBalance = sdk.ZeroInt()
		sr.NetflowRate = sdk.ZeroInt()
	}
	return sr
}

func (s *StorageTestSuite) GetStreamRecords() (streamRecords StreamRecords) {
	streamRecords.User = s.GetStreamRecord(s.User.GetAddr().String())
	for _, sp := range s.StorageProviders {
		sr := s.GetStreamRecord(sp.OperatorKey.GetAddr().String())
		streamRecords.SPs = append(streamRecords.SPs, sr)
	}
	return streamRecords
}

func (s *StorageTestSuite) TestPayment_Smoke() {
	ctx := context.Background()
	sp := s.StorageProviders[0]
	user := s.User
	var err error

	streamRecordsBeforeCreateBucket := s.GetStreamRecords()
	s.T().Logf("streamRecordsBeforeCreateBucket: %s", core.YamlString(streamRecordsBeforeCreateBucket))

	// create bucket
	bucketName := storageutils.GenRandomBucketName()
	bucketReadQuota := uint64(1000)
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.ReadQuota = bucketReadQuota
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user)

	// check bill after creating bucket
	userBankAccount, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: user.GetAddr().String(),
		Denom:   s.Config.Denom,
	})
	s.Require().NoError(err)
	s.T().Logf("user bank account %s", userBankAccount)
	streamRecordsAfterCreateBucket := s.GetStreamRecords()
	usr := streamRecordsAfterCreateBucket.User
	ssr0 := streamRecordsBeforeCreateBucket.SPs[0]
	ssr1 := streamRecordsAfterCreateBucket.SPs[0]
	s.Require().Equal(usr.StaticBalance, sdkmath.ZeroInt())
	s.Require().Len(usr.OutFlows, 1)
	s.Require().Equal(usr.OutFlows[0].Rate, usr.NetflowRate.Neg())
	s.Require().Equal(usr.OutFlows[0].ToAddress, ssr1.Account)
	s.Require().Equal(usr.NetflowRate, ssr0.NetflowRate.Sub(ssr1.NetflowRate))
	// check price and rate calculation
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: queryHeadBucketResponse.BucketInfo.BillingInfo.PriceTime,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	readChargeRate := readPrice.MulInt(sdk.NewIntFromUint64(queryHeadBucketResponse.BucketInfo.ReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	s.Require().Equal(usr.NetflowRate.Abs(), readChargeRate)

	// CreateObject
	objectName := storageutils.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user)

	// check lock balance
	queryHeadBucketResponseAfterCreateObj, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.T().Logf("queryHeadBucketResponseAfterCreateObj %s, err: %v", queryHeadBucketResponseAfterCreateObj, err)
	s.Require().NoError(err)
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.T().Logf("queryHeadObjectResponse %s, err: %v", queryHeadObjectResponse, err)
	s.Require().NoError(err)
	streamRecordsAfterCreateObject := s.GetStreamRecords()
	s.T().Logf("streamRecordsAfterCreateObject %s", core.YamlString(streamRecordsAfterCreateObject))
	userStreamAccountAfterCreateObj := streamRecordsAfterCreateObject.User
	queryGetSecondarySpStorePriceByTime, err := s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: queryHeadBucketResponse.BucketInfo.BillingInfo.PriceTime,
	})
	s.T().Logf("queryGetSecondarySpStorePriceByTime %s, err: %v", queryGetSecondarySpStorePriceByTime, err)
	s.Require().NoError(err)
	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)
	primaryStorePrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice
	secondaryStorePrice := queryGetSecondarySpStorePriceByTime.SecondarySpStorePrice.StorePrice
	chargeSize := s.GetChargeSize(queryHeadObjectResponse.ObjectInfo.PayloadSize)
	expectedChargeRate := primaryStorePrice.Add(secondaryStorePrice.MulInt64(6)).MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt()
	expectedLockedBalance := expectedChargeRate.Mul(sdkmath.NewIntFromUint64(paymentParams.Params.ReserveTime))
	s.Require().Equal(expectedLockedBalance.String(), userStreamAccountAfterCreateObj.LockBalance.String())

	// seal object
	secondaryStorageProviders := s.StorageProviders[1:7]
	secondarySPs := lo.Map(secondaryStorageProviders, func(sp core.SPKeyManagers, i int) sdk.AccAddress {
		return sp.OperatorKey.GetAddr()
	})
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, secondarySPs, nil)
	secondarySigs := lo.Map(secondaryStorageProviders, func(sp core.SPKeyManagers, i int) []byte {
		sr := storagetypes.NewSecondarySpSignDoc(sp.OperatorKey.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id, checksum)
		secondarySig, err := sp.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
		s.Require().NoError(err)
		err = storagetypes.VerifySignature(sp.ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()), secondarySig)
		s.Require().NoError(err)
		return secondarySig
	})
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(msgSealObject, sp.SealKey)

	// check bill after seal
	streamRecordsAfterSeal := s.GetStreamRecords()
	s.T().Logf("streamRecordsAfterSeal %s", core.YamlString(streamRecordsAfterSeal))
	s.Require().Equal(sdkmath.ZeroInt(), streamRecordsAfterSeal.User.LockBalance)
	userRateDiff := streamRecordsAfterSeal.User.NetflowRate.Sub(streamRecordsAfterCreateObject.User.NetflowRate)
	spRateDiffs := lo.Map(streamRecordsAfterSeal.SPs, func(sp paymenttypes.StreamRecord, i int) sdkmath.Int {
		return sp.NetflowRate.Sub(streamRecordsAfterCreateObject.SPs[i].NetflowRate)
	})
	spRateDiffsSum := lo.Reduce(spRateDiffs, func(sum sdkmath.Int, rate sdkmath.Int, i int) sdkmath.Int {
		return sum.Add(rate)
	}, sdkmath.ZeroInt())
	s.Require().Equal(userRateDiff, spRateDiffsSum.Neg())
	spRateDiffMap := lo.Reduce(spRateDiffs, func(m map[string]sdkmath.Int, rate sdkmath.Int, i int) map[string]sdkmath.Int {
		m[streamRecordsAfterSeal.SPs[i].Account] = rate
		return m
	}, make(map[string]sdkmath.Int))
	userOutflowMap := lo.Reduce(streamRecordsAfterSeal.User.OutFlows, func(m map[string]sdkmath.Int, outflow paymenttypes.OutFlow, i int) map[string]sdkmath.Int {
		m[outflow.ToAddress] = outflow.Rate
		return m
	}, make(map[string]sdkmath.Int))
	primarySpAddr := s.StorageProviders[0].OperatorKey.GetAddr().String()
	s.Require().Equal(
		userOutflowMap[primarySpAddr].Sub(readChargeRate).String(),
		spRateDiffMap[primarySpAddr].String())
	diff := spRateDiffMap[primarySpAddr].Sub(primaryStorePrice.MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt())
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	s.T().Logf("diff %s", diff.String())
	s.Require().Equal(diff.String(), sdkmath.ZeroInt().String())
	s.Require().Equal(spRateDiffMap[primarySpAddr].String(), primaryStorePrice.MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt().String())
	for i, sp := range secondarySPs {
		secondarySpAddr := sp.String()
		s.Require().Equal(userOutflowMap[secondarySpAddr].String(), spRateDiffMap[secondarySpAddr].String(), "sp %d", i+1)
		s.Require().Equal(userOutflowMap[secondarySpAddr].String(), secondaryStorePrice.MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt().String())
	}

	// change read quota

	// delete object

	// delete bucket
}

func (s *StorageTestSuite) TestPayment_AutoSettle() {
	ctx := context.Background()
	sp := s.StorageProviders[0]
	user := s.User
	userAddr := user.GetAddr().String()
	var err error

	bucketReadQuota := uint64(1000)
	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)
	reserveTime := paymentParams.Params.ReserveTime
	forcedSettleTime := paymentParams.Params.ForcedSettleTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	paymentAccountBNBNeeded := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketReadQuota * reserveTime)).TruncateInt()
	expectedRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketReadQuota)).TruncateInt()

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: userAddr,
	}
	_ = s.SendTxBlock(msgCreatePaymentAccount, user)
	paymentAccountsReq := &paymenttypes.QueryGetPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.GetPaymentAccountsByOwner(ctx, paymentAccountsReq)
	s.Require().NoError(err)
	s.T().Logf("paymentAccounts %s", core.YamlString(paymentAccounts))
	paymentAddr := paymentAccounts.PaymentAccounts[0]
	s.Require().Lenf(paymentAccounts.PaymentAccounts, 1, "paymentAccounts %s", core.YamlString(paymentAccounts))
	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAddr,
		Amount:  paymentAccountBNBNeeded,
	}
	_ = s.SendTxBlock(msgDeposit, user)

	// create bucket from payment account
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil)
	msgCreateBucket.ReadQuota = bucketReadQuota
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user)
	// check payment account stream record
	paymentAccountStreamRecord := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), sdkmath.ZeroInt().String())

	// create bucket from user
	msgCreateBucket.BucketName = storageutils.GenRandomBucketName()
	msgCreateBucket.PaymentAddress = ""
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	res := s.SendTxBlock(msgCreateBucket, user)
	s.T().Logf("res %s", core.YamlString(res))
	// check user stream record
	userStreamRecord := s.GetStreamRecord(userAddr)
	s.T().Logf("userStreamRecord %s", core.YamlString(userStreamRecord))
	s.Require().Equal(userStreamRecord.SettleTimestamp, userStreamRecord.CrudTimestamp+int64(reserveTime-forcedSettleTime))
	spStreamRecord := s.GetStreamRecord(sp.OperatorKey.GetAddr().String())
	s.T().Logf("spStreamRecord %s", core.YamlString(spStreamRecord))
	govStreamRecord := s.GetStreamRecord(paymenttypes.GovernanceAddress.String())
	s.T().Logf("govStreamRecord %s", core.YamlString(govStreamRecord))

	// wait until settle time
	retryCount := 0
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)
		currentTimestamp := latestBlock.Block.Time.Unix()
		s.T().Logf("currentTimestamp %d, userStreamRecord.SettleTimestamp %d", currentTimestamp, userStreamRecord.SettleTimestamp)
		if currentTimestamp > userStreamRecord.SettleTimestamp {
			break
		}
		time.Sleep(time.Second)
		retryCount++
		if retryCount > 31 {
			s.T().Fatalf("wait for settle time timeout")
		}
	}
	// check auto settle
	userStreamRecordAfterAutoSettle := s.GetStreamRecord(userAddr)
	s.T().Logf("userStreamRecordAfterAutoSettle %s", core.YamlString(userStreamRecordAfterAutoSettle))
	spStreamRecordAfterAutoSettle := s.GetStreamRecord(sp.OperatorKey.GetAddr().String())
	s.T().Logf("spStreamRecordAfterAutoSettle %s", core.YamlString(spStreamRecordAfterAutoSettle))
	paymentAccountStreamRecordAfterAutoSettle := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterAutoSettle %s", core.YamlString(paymentAccountStreamRecordAfterAutoSettle))
	// payment account become frozen
	s.Require().NotEqual(paymentAccountStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(spStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(userStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	// user settle time become refreshed
	s.Require().NotEqual(userStreamRecordAfterAutoSettle.SettleTimestamp, userStreamRecord.SettleTimestamp)
	s.Require().Equal(userStreamRecordAfterAutoSettle.SettleTimestamp, userStreamRecordAfterAutoSettle.CrudTimestamp+int64(reserveTime-forcedSettleTime))
	// gov stream record balance increase
	govStreamRecordAfterSettle := s.GetStreamRecord(paymenttypes.GovernanceAddress.String())
	s.T().Logf("govStreamRecordAfterSettle %s", core.YamlString(govStreamRecordAfterSettle))
	s.Require().NotEqual(govStreamRecordAfterSettle.StaticBalance.String(), govStreamRecord.StaticBalance.String())
	govStreamRecordStaticBalanceDelta := govStreamRecordAfterSettle.StaticBalance.Sub(govStreamRecord.StaticBalance)
	expectedGovBalanceDelta := userStreamRecord.NetflowRate.Neg().MulRaw(userStreamRecordAfterAutoSettle.CrudTimestamp - userStreamRecord.CrudTimestamp)
	s.Require().Equal(expectedGovBalanceDelta.String(), govStreamRecordStaticBalanceDelta.String())

	// deposit, balance not enough to resume
	depositAmount1 := sdk.NewInt(1)
	msgDeposit1 := &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  depositAmount1,
	}
	_ = s.SendTxBlock(msgDeposit1, user)
	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit1 := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit1 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit1))
	s.Require().NotEqual(paymentAccountStreamRecordAfterDeposit1.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit1.StaticBalance.String(), paymentAccountStreamRecordAfterAutoSettle.StaticBalance.Add(depositAmount1).String())

	// deposit and resume
	depositAmount2 := sdk.NewInt(1e10)
	msgDeposit2 := &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  depositAmount2,
	}
	s.SendTxBlock(msgDeposit2, user)
	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit2 := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit2 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit2))
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.StaticBalance.Add(paymentAccountStreamRecordAfterDeposit2.BufferBalance).String(), paymentAccountStreamRecordAfterDeposit1.StaticBalance.Add(depositAmount2).String())
}

func (s *StorageTestSuite) GetChargeSize(payloadSize uint64) uint64 {
	ctx := context.Background()
	storageParams, err := s.Client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	s.Require().NoError(err)
	s.T().Logf("storageParams %s", storageParams)
	minChargeSize := storageParams.Params.MinChargeSize
	if payloadSize < minChargeSize {
		return minChargeSize
	} else {
		return payloadSize
	}
}

func (s *StorageTestSuite) TestMirrorBucket() {
	var err error
	sp := s.StorageProviders[0]
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// Mirror bucket
	msgMirrorBucket := storagetypes.NewMsgMirrorBucket(user.GetAddr(), queryHeadBucketResponse.BucketInfo.Id)
	s.SendTxBlock(msgMirrorBucket, user)
}

func (s *StorageTestSuite) TestMirrorObject() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// CreateObject
	objectName := storageutils.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)

	// SealObject
	secondarySPs := []sdk.AccAddress{
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
	}
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, secondarySPs, nil)
	sr := storagetypes.NewSecondarySpSignDoc(sp.OperatorKey.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id, checksum)
	secondarySig, err := sp.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
	s.Require().NoError(err)
	err = storagetypes.VerifySignature(s.StorageProviders[0].ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()),
		secondarySig)
	s.Require().NoError(err)

	secondarySigs := [][]byte{secondarySig, secondarySig, secondarySig, secondarySig, secondarySig, secondarySig}
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(msgSealObject, sp.SealKey)

	// ListBuckets
	queryListBucketsRequest := storagetypes.QueryListBucketsRequest{}
	queryListBucketResponse, err := s.Client.ListBuckets(ctx, &queryListBucketsRequest)
	s.Require().NoError(err)
	s.Require().Greater(len(queryListBucketResponse.BucketInfos), 0)

	// ListObject
	queryListObjectsRequest := storagetypes.QueryListObjectsRequest{
		BucketName: bucketName,
	}
	queryListObjectsResponse, err := s.Client.ListObjects(ctx, &queryListObjectsRequest)
	s.Require().NoError(err)
	s.Require().Equal(len(queryListObjectsResponse.ObjectInfos), 1)
	s.Require().Equal(queryListObjectsResponse.ObjectInfos[0].ObjectName, objectName)

	// Mirror object
	msgMirrorObject := storagetypes.NewMsgMirrorObject(user.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id)
	s.SendTxBlock(msgMirrorObject, user)
}

func (s *StorageTestSuite) TestMirrorGroup() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()})
	s.SendTxBlock(msgCreateGroup, owner)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())

	// Mirror group
	msgMirrorGroup := storagetypes.NewMsgMirrorGroup(owner.GetAddr(), queryHeadGroupResp.GroupInfo.Id)
	s.SendTxBlock(msgMirrorGroup, owner)
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}
