package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	"github.com/bnb-chain/greenfield/types/common"
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
	Tax  paymenttypes.StreamRecord
}

func (s *StorageTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageTestSuite) SetupTest() {
	s.User = s.GenAndChargeAccounts(1, 1000000)[0]
}

var (
	line = `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`
)

func (s *StorageTestSuite) TestCreateBucket() {
	var err error
	sp := s.StorageProviders[0]
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
		user.GetAddr(), bucketName, nil, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
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
		nil, math.MaxUint, nil, 0)
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

	// 3. ListGroup
	queryListGroupReq := storagetypes.QueryListGroupRequest{GroupOwner: owner.GetAddr().String()}
	queryListGroupResp, err := s.Client.ListGroup(ctx, &queryListGroupReq)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(queryListGroupResp.GroupInfos), 1)

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
		nil, math.MaxUint, nil, 0)
	msgCreateBucket1.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket1.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket1, user)

	// 2. CreateBucket2
	bucketName2 := storageutils.GenRandomBucketName()
	msgCreateBucket2 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName2, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket2.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket2.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket2, user)

	// 3. Create object into bucket1
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
	streamRecords.Tax = s.GetStreamRecord(paymenttypes.ValidatorTaxPoolAddress.String())
	return streamRecords
}

func (s *StorageTestSuite) CheckStreamRecordsBeforeAndAfter(streamRecordsBefore StreamRecords, streamRecordsAfter StreamRecords, readPrice sdk.Dec,
	readChargeRate sdkmath.Int, primaryStorePrice sdk.Dec, secondaryStorePrice sdk.Dec, chargeSize uint64, secondarySPs []sdk.AccAddress, payloadSize uint64) {
	userRateDiff := streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate)
	taxRateDiff := streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate)
	spRateDiffs := lo.Map(streamRecordsAfter.SPs, func(sp paymenttypes.StreamRecord, i int) sdkmath.Int {
		return sp.NetflowRate.Sub(streamRecordsBefore.SPs[i].NetflowRate)
	})
	spRateDiffsSum := lo.Reduce(spRateDiffs, func(sum sdkmath.Int, rate sdkmath.Int, i int) sdkmath.Int {
		return sum.Add(rate)
	}, sdkmath.ZeroInt())
	s.Require().Equal(userRateDiff, spRateDiffsSum.Add(taxRateDiff).Neg())
	spRateDiffMap := lo.Reduce(spRateDiffs, func(m map[string]sdkmath.Int, rate sdkmath.Int, i int) map[string]sdkmath.Int {
		m[streamRecordsAfter.SPs[i].Account] = rate
		return m
	}, make(map[string]sdkmath.Int))
	userOutflowMap := lo.Reduce(streamRecordsAfter.User.OutFlows, func(m map[string]sdkmath.Int, outflow paymenttypes.OutFlow, i int) map[string]sdkmath.Int {
		m[outflow.ToAddress] = outflow.Rate
		return m
	}, make(map[string]sdkmath.Int))
	if payloadSize != 0 {
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
	}

}

func (s *StorageTestSuite) TestPayment_Smoke() {
	ctx := context.Background()
	sp := s.StorageProviders[0]
	user := s.User
	var err error

	streamRecordsBeforeCreateBucket := s.GetStreamRecords()
	s.T().Logf("streamRecordsBeforeCreateBucket: %s", core.YamlString(streamRecordsBeforeCreateBucket))
	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	// create bucket
	bucketName := storageutils.GenRandomBucketName()
	bucketChargedReadQuota := uint64(1000)
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.ChargedReadQuota = bucketChargedReadQuota
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
	ssr1 := streamRecordsAfterCreateBucket.SPs[0]
	s.Require().Equal(usr.StaticBalance, sdkmath.ZeroInt())
	s.Require().Len(usr.OutFlows, 2)
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
	readChargeRate := readPrice.MulInt(sdk.NewIntFromUint64(queryHeadBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	userTaxRate := paymentParams.Params.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate := readChargeRate.Add(userTaxRate)
	s.Require().Equal(usr.NetflowRate.Abs(), userTotalRate)
	expectedOutFlows := []paymenttypes.OutFlow{
		{ToAddress: ssr1.Account, Rate: readChargeRate},
		{ToAddress: paymenttypes.ValidatorTaxPoolAddress.String(), Rate: userTaxRate},
	}
	sort.Slice(usr.OutFlows, func(i, j int) bool {
		return usr.OutFlows[i].ToAddress < usr.OutFlows[j].ToAddress
	})
	sort.Slice(expectedOutFlows, func(i, j int) bool {
		return expectedOutFlows[i].ToAddress < expectedOutFlows[j].ToAddress
	})
	s.Require().Equal(usr.OutFlows, expectedOutFlows)

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
	s.CheckStreamRecordsBeforeAndAfter(streamRecordsAfterCreateObject, streamRecordsAfterSeal, readPrice, readChargeRate, primaryStorePrice, secondaryStorePrice, chargeSize, secondarySPs, uint64(payloadSize))

	// create empty object
	streamRecordsBeforeCreateEmptyObject := s.GetStreamRecords()
	s.T().Logf("streamRecordsBeforeCreateEmptyObject %s", core.YamlString(streamRecordsBeforeCreateEmptyObject))

	emptyObjectName := "sub_directory/"
	// create empty test buffer
	var emptyBuffer bytes.Buffer
	emptyPayloadSize := emptyBuffer.Len()
	emptyChecksum := sdk.Keccak256(emptyBuffer.Bytes())
	emptyExpectChecksum := [][]byte{emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum}
	msgCreateEmptyObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, emptyObjectName, uint64(emptyPayloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, emptyExpectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateEmptyObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateEmptyObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateEmptyObject, user)

	streamRecordsAfterCreateEmptyObject := s.GetStreamRecords()
	s.T().Logf("streamRecordsAfterCreateEmptyObject %s", core.YamlString(streamRecordsAfterCreateEmptyObject))
	chargeSize = s.GetChargeSize(uint64(emptyPayloadSize))

	s.CheckStreamRecordsBeforeAndAfter(streamRecordsBeforeCreateEmptyObject, streamRecordsAfterCreateEmptyObject, readPrice, readChargeRate, primaryStorePrice, secondaryStorePrice, chargeSize, secondarySPs, uint64(emptyPayloadSize))

	// test query auto settle records
	queryAllAutoSettleRecordRequest := paymenttypes.QueryAllAutoSettleRecordRequest{}
	queryAllAutoSettleRecordResponse, err := s.Client.AutoSettleRecordAll(ctx, &queryAllAutoSettleRecordRequest)
	s.Require().NoError(err)
	s.T().Logf("queryAllAutoSettleRecordResponse %s", core.YamlString(queryAllAutoSettleRecordResponse))
	s.Require().True(len(queryAllAutoSettleRecordResponse.AutoSettleRecord) >= 1)

	// change read quota

	// delete object

	// delete bucket
}

func (s *StorageTestSuite) TestPayment_DeleteBucketWithReadQuota() {
	var err error
	sp := s.StorageProviders[0]
	user := s.User
	// CreateBucket
	chargedReadQuota := uint64(100)
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, chargedReadQuota)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user)

	streamRecordsBeforeDelete := s.GetStreamRecords()
	s.T().Logf("streamRecordsBeforeDelete: %s", core.YamlString(streamRecordsBeforeDelete))
	s.Require().NotEqual(streamRecordsBeforeDelete.User.NetflowRate.String(), "0")

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(msgDeleteBucket, user)

	// check the billing change
	streamRecordsAfterDelete := s.GetStreamRecords()
	s.T().Logf("streamRecordsBeforeDelete: %s", core.YamlString(streamRecordsAfterDelete))
	s.Require().Equal(streamRecordsAfterDelete.User.NetflowRate.String(), "0")
}

func (s *StorageTestSuite) TestPayment_AutoSettle() {
	ctx := context.Background()
	sp := s.StorageProviders[0]
	user := s.User
	userAddr := user.GetAddr().String()
	var err error

	bucketChargedReadQuota := uint64(1000)
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
	totalUserRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	taxRateParam := paymentParams.Params.ValidatorTaxRate
	taxStreamRate := taxRateParam.MulInt(totalUserRate).TruncateInt()
	expectedRate := totalUserRate.Add(taxStreamRate)
	paymentAccountBNBNeeded := expectedRate.Mul(sdkmath.NewIntFromUint64(reserveTime))

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
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil, bucketChargedReadQuota)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user)
	// check payment account stream record
	paymentAccountStreamRecord := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), sdkmath.ZeroInt().String())

	// increase bucket charged read quota is not allowed since the balance is not enough
	msgUpdateBucketInfo := &storagetypes.MsgUpdateBucketInfo{
		Operator:         user.GetAddr().String(),
		BucketName:       bucketName,
		ChargedReadQuota: &common.UInt64Value{Value: bucketChargedReadQuota + 1},
		Visibility:       storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
	}
	_, err = s.SendTxBlockWithoutCheck(msgUpdateBucketInfo, user)
	s.Require().ErrorContains(err, "balance not enough, lack of")

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
		if retryCount > 60 {
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
		nil, math.MaxUint, nil, 0)
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
		nil, math.MaxUint, nil, 0)
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

func (s *StorageTestSuite) TestDiscontinueObject_Normal() {
	sp1, _, bucketName1, _, _, objectId1 := s.createObject()
	sp2, _, bucketName2, _, _, objectId2 := s.createObject()

	// DiscontinueObject
	msgDiscontinueObject := storagetypes.NewMsgDiscontinueObject(sp1.GcKey.GetAddr(), bucketName1, []sdkmath.Uint{objectId1}, "test")
	txRes := s.SendTxBlock(msgDiscontinueObject, sp1.GcKey)
	deleteAt := int64(filterDiscontinueObjectEventFromTx(txRes).DeleteAt)

	time.Sleep(5 * time.Second)
	msgDiscontinueObject2 := storagetypes.NewMsgDiscontinueObject(sp2.GcKey.GetAddr(), bucketName2, []sdkmath.Uint{objectId2}, "test")
	s.SendTxBlock(msgDiscontinueObject2, sp2.GcKey)

	// Wait after the delete timestamp
	heightBefore := txRes.Height
	heightAfter := int64(0)
	for {
		time.Sleep(500 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime >= deleteAt {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	events := make([]storagetypes.EventDeleteObject, 0)
	for heightBefore <= heightAfter {
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &heightBefore)
		s.Require().NoError(err)
		events = append(events, filterDeleteObjectEventFromBlock(blockRes)...)
		heightBefore++
	}

	object1Found, object2Found := false, false
	for _, event := range events {
		if event.ObjectId.Equal(objectId1) {
			object1Found = true
		}
		if event.ObjectId.Equal(objectId2) {
			object2Found = true
		}
	}
	s.Require().True(object1Found)
	s.Require().True(!object2Found)
}

func (s *StorageTestSuite) TestDiscontinueObject_UserDeleted() {
	sp, user, bucketName, _, objectName, objectId := s.createObject()

	// DiscontinueObject
	msgDiscontinueObject := storagetypes.NewMsgDiscontinueObject(sp.GcKey.GetAddr(), bucketName, []sdkmath.Uint{objectId}, "test")
	txRes := s.SendTxBlock(msgDiscontinueObject, sp.GcKey)
	deleteAt := filterDiscontinueObjectEventFromTx(txRes).DeleteAt

	// DeleteObject before discontinue confirm window
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	txRes = s.SendTxBlock(msgDeleteObject, user)
	event := filterDeleteObjectEventFromTx(txRes)
	s.Require().Equal(event.ObjectId, objectId)

	// Wait after the delete timestamp
	heightBefore := txRes.Height
	heightAfter := int64(0)
	for {
		time.Sleep(500 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime >= deleteAt {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	events := make([]storagetypes.EventDeleteObject, 0)
	for heightBefore <= heightAfter {
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &heightBefore)
		s.Require().NoError(err)
		events = append(events, filterDeleteObjectEventFromBlock(blockRes)...)
		heightBefore++
	}

	// Already deleted by user
	found := false
	for _, event := range events {
		if event.ObjectId.Equal(objectId) {
			found = true
		}
	}
	s.Require().True(!found)

	time.Sleep(500 * time.Millisecond)
	statusRes, err := s.TmClient.TmClient.Status(context.Background())
	s.Require().NoError(err)
	s.Require().True(statusRes.SyncInfo.LatestBlockHeight > heightAfter)
}

func (s *StorageTestSuite) TestDiscontinueBucket_Normal() {
	sp1, _, bucketName1, bucketId1, _, _ := s.createObject()
	sp2, _, bucketName2, bucketId2, _, _ := s.createObject()

	// DiscontinueBucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp1.GcKey.GetAddr(), bucketName1, "test")
	txRes := s.SendTxBlock(msgDiscontinueBucket, sp1.GcKey)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	time.Sleep(3 * time.Second)
	msgDiscontinueBucket2 := storagetypes.NewMsgDiscontinueBucket(sp1.GcKey.GetAddr(), bucketName2, "test")
	s.SendTxBlock(msgDiscontinueBucket2, sp2.GcKey)

	// Wait after the delete timestamp
	heightBefore := txRes.Height
	heightAfter := int64(0)
	for {
		time.Sleep(500 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime >= deleteAt {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	events := make([]storagetypes.EventDeleteBucket, 0)
	for heightBefore <= heightAfter {
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &heightBefore)
		s.Require().NoError(err)
		events = append(events, filterDeleteBucketEventFromBlock(blockRes)...)
		heightBefore++
	}

	bucket1Found, bucket2Found := false, false
	for _, event := range events {
		if event.BucketId.Equal(bucketId1) {
			bucket1Found = true
		}
		if event.BucketId.Equal(bucketId2) {
			bucket2Found = true
		}
	}
	s.Require().True(bucket1Found)
	s.Require().True(!bucket2Found)
}

func (s *StorageTestSuite) TestDiscontinueBucket_UserDeleted() {
	sp, user, bucketName, bucketId, objectName, _ := s.createObject()

	// DiscontinueBucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(msgDiscontinueBucket, sp.GcKey)
	deleteAt := int64(filterDiscontinueBucketEventFromTx(txRes).DeleteAt)

	// DeleteBucket before discontinue confirm window
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(msgDeleteObject, user)
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	txRes = s.SendTxBlock(msgDeleteBucket, user)
	event := filterDeleteBucketEventFromTx(txRes)
	s.Require().Equal(event.BucketId, bucketId)

	// Wait after the delete timestamp
	heightBefore := txRes.Height
	heightAfter := int64(0)
	for {
		time.Sleep(500 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime >= deleteAt {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	events := make([]storagetypes.EventDeleteBucket, 0)
	for heightBefore <= heightAfter {
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &heightBefore)
		s.Require().NoError(err)
		events = append(events, filterDeleteBucketEventFromBlock(blockRes)...)
		heightBefore++
	}

	// Already deleted by user
	found := false
	for _, event := range events {
		if event.BucketId.Equal(bucketId) {
			found = true
		}
	}
	s.Require().True(!found)

	time.Sleep(500 * time.Millisecond)
	statusRes, err := s.TmClient.TmClient.Status(context.Background())
	s.Require().NoError(err)
	s.Require().True(statusRes.SyncInfo.LatestBlockHeight > heightAfter)
}

func (s *StorageTestSuite) createObject() (core.SPKeyManagers, keys.KeyManager, string, storagetypes.Uint, string, storagetypes.Uint) {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	return sp, user, bucketName, queryHeadBucketResponse.BucketInfo.Id, objectName, queryListObjectsResponse.ObjectInfos[0].Id
}

func TestStorageTestSuite(t *testing.T) {
	suite.Run(t, new(StorageTestSuite))
}

func filterDiscontinueObjectEventFromTx(txRes *sdk.TxResponse) storagetypes.EventDiscontinueObject {
	deleteAtStr := ""
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "bnbchain.greenfield.storage.EventDiscontinueObject" {
			for _, attr := range event.Attributes {
				if attr.Key == "delete_at" {
					deleteAtStr = strings.Trim(attr.Value, `"`)
				}
			}
		}
	}
	deleteAt, _ := strconv.ParseInt(deleteAtStr, 10, 64)
	return storagetypes.EventDiscontinueObject{
		DeleteAt: deleteAt,
	}
}

func filterDeleteObjectEventFromBlock(blockRes *ctypes.ResultBlockResults) []storagetypes.EventDeleteObject {
	events := make([]storagetypes.EventDeleteObject, 0)

	for _, event := range blockRes.EndBlockEvents {
		if event.Type == "bnbchain.greenfield.storage.EventDeleteObject" {
			objectIdStr := ""
			for _, attr := range event.Attributes {
				if string(attr.Key) == "object_id" {
					objectIdStr = strings.Trim(string(attr.Value), `"`)
				}
			}
			objectId := sdkmath.NewUintFromString(objectIdStr)
			events = append(events, storagetypes.EventDeleteObject{
				ObjectId: objectId,
			})
		}
	}
	return events
}

func filterDeleteObjectEventFromTx(txRes *sdk.TxResponse) storagetypes.EventDeleteObject {
	objectIdStr := ""
	for _, event := range txRes.Events {
		if event.Type == "bnbchain.greenfield.storage.EventDeleteObject" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "object_id" {
					objectIdStr = strings.Trim(string(attr.Value), `"`)
				}
			}
		}
	}
	objectId := sdkmath.NewUintFromString(objectIdStr)
	return storagetypes.EventDeleteObject{
		ObjectId: objectId,
	}
}

func filterDiscontinueBucketEventFromTx(txRes *sdk.TxResponse) storagetypes.EventDiscontinueBucket {
	deleteAtStr := ""
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "bnbchain.greenfield.storage.EventDiscontinueBucket" {
			for _, attr := range event.Attributes {
				if attr.Key == "delete_at" {
					deleteAtStr = strings.Trim(attr.Value, `"`)
				}
			}
		}
	}
	deleteAt, _ := strconv.ParseInt(deleteAtStr, 10, 64)
	return storagetypes.EventDiscontinueBucket{
		DeleteAt: deleteAt,
	}
}

func filterDeleteBucketEventFromBlock(blockRes *ctypes.ResultBlockResults) []storagetypes.EventDeleteBucket {
	events := make([]storagetypes.EventDeleteBucket, 0)

	for _, event := range blockRes.EndBlockEvents {
		if event.Type == "bnbchain.greenfield.storage.EventDeleteBucket" {
			bucketIdStr := ""
			for _, attr := range event.Attributes {
				if string(attr.Key) == "bucket_id" {
					bucketIdStr = strings.Trim(string(attr.Value), `"`)
				}
			}
			bucketId := sdkmath.NewUintFromString(bucketIdStr)
			events = append(events, storagetypes.EventDeleteBucket{
				BucketId: bucketId,
			})
		}
	}
	return events
}

func filterDeleteBucketEventFromTx(txRes *sdk.TxResponse) storagetypes.EventDeleteBucket {
	bucketIdStr := ""
	for _, event := range txRes.Events {
		if event.Type == "bnbchain.greenfield.storage.EventDeleteBucket" {
			for _, attr := range event.Attributes {
				if string(attr.Key) == "bucket_id" {
					bucketIdStr = strings.Trim(string(attr.Value), `"`)
				}
			}
		}
	}
	bucketId := sdkmath.NewUintFromString(bucketIdStr)
	return storagetypes.EventDeleteBucket{
		BucketId: bucketId,
	}
}

func (s *StorageTestSuite) TestCancelCreateObject() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(len(queryHeadObjectResponse.ObjectInfo.SecondarySpAddresses), 0)
	// CancelCreateObject
	msgCancelCreateObject := storagetypes.NewMsgCancelCreateObject(user.GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlock(msgCancelCreateObject, user)
}

func (s *StorageTestSuite) TestCreateObjectWithCommonPrefix() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	objectName := "sub_directory/"
	// create empty test buffer
	var buffer bytes.Buffer

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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)
	s.Require().Equal(len(queryHeadObjectResponse.ObjectInfo.SecondarySpAddresses), 0)

	// CopyObject
	dstBucketName := bucketName
	dstObjectName := "new_directory/"
	msgCopyObject := storagetypes.NewMsgCopyObject(user.GetAddr(), bucketName, dstBucketName, objectName, dstObjectName, math.MaxUint, nil)
	msgCopyObject.DstPrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCopyObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCopyObject, user)

	// HeadObject
	queryCopyObjectHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: dstBucketName,
		ObjectName: dstObjectName,
	}
	queryCopyObjectHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryCopyObjectHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.ObjectName, dstObjectName)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.BucketName, dstBucketName)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.PayloadSize, uint64(payloadSize))
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryCopyObjectHeadObjectResponse.ObjectInfo.ContentType, contextType)
	s.Require().Equal(len(queryCopyObjectHeadObjectResponse.ObjectInfo.SecondarySpAddresses), 0)
}
