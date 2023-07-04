package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	types2 "github.com/bnb-chain/greenfield/x/virtualgroup/types"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type StorageTestSuite struct {
	core.BaseSuite
	User keys.KeyManager
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
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// UpdateBucketInfo
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgUpdateBucketInfo)
	s.Require().NoError(err)

	// verify modified bucketinfo
	queryHeadBucketResponseAfterUpdateBucket, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponseAfterUpdateBucket.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)
}

func (s *StorageTestSuite) TestCreateObject() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

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
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[i], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

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

	// UpdateObjectInfo
	updateObjectInfo := storagetypes.NewMsgUpdateObjectInfo(
		user.GetAddr(), bucketName, objectName, storagetypes.VISIBILITY_TYPE_INHERIT)
	s.Require().NoError(err)
	s.SendTxBlock(user, updateObjectInfo)
	s.Require().NoError(err)

	// verify modified objectinfo
	// head object
	queryHeadObjectAfterUpdateObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_INHERIT)

	// DeleteObject
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)
}

func (s *StorageTestSuite) TestCreateGroup() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()}, "")
	s.SendTxBlock(owner, msgCreateGroup)
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
	s.SendTxBlock(owner, msgUpdateGroupMember)

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
	msgCreateGroup = storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()}, "")
	s.SendTxBlockWithExpectErrorString(msgCreateGroup, owner, "exists")
}

func (s *StorageTestSuite) TestDeleteBucket() {
	var err error
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	s.T().Logf("Global virtual group: %s", gvg.String())
	// 1. CreateBucket1
	bucketName1 := storageutils.GenRandomBucketName()
	msgCreateBucket1 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName1, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket1.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket1.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket1.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket1)

	// 2. CreateBucket2
	bucketName2 := storageutils.GenRandomBucketName()
	msgCreateBucket2 := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName2, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket2.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket2.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket2.
		GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket2)

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
		storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

	// head object
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName1,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.T().Logf("queryHeadObjectResponse %s, err: %v", queryHeadObjectResponse, err)
	s.Require().NoError(err)

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName1, objectName,
		gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[i], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig

	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

	// 4. Delete bucket2
	msgDeleteBucket2 := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName2)
	s.SendTxBlock(user, msgDeleteBucket2)

	// 5. Delete object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName1, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// 6. delete bucket1
	msgDeleteBucket1 := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName1)
	s.SendTxBlock(user, msgDeleteBucket1)
}

func (s *StorageTestSuite) TestMirrorBucket() {
	var err error
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.User
	// CreateBucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// MirrorBucket using id
	msgMirrorBucket := storagetypes.NewMsgMirrorBucket(user.GetAddr(), queryHeadBucketResponse.BucketInfo.Id, "")
	s.SendTxBlock(user, msgMirrorBucket)

	// CreateBucket
	bucketName = storageutils.GenRandomBucketName()
	msgCreateBucket = storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// MirrorBucket using name
	msgMirrorBucket = storagetypes.NewMsgMirrorBucket(user.GetAddr(), sdk.NewUint(0), bucketName)
	s.SendTxBlock(user, msgMirrorBucket)
}

func (s *StorageTestSuite) TestMirrorObject() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

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
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvg.Id, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[i], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

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

	// MirrorObject using id
	msgMirrorObject := storagetypes.NewMsgMirrorObject(user.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id, "", "")
	s.SendTxBlock(user, msgMirrorObject)

	// CreateObject
	objectName = storageutils.GenRandomObjectName()
	msgCreateObject = storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

	queryHeadObjectRequest = storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)

	// SealObject
	gvgId := gvg.Id
	msgSealObject = storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvgId, nil)
	secondarySigs = make([][]byte, 0)
	secondarySPBlsPubKeys = make([]bls.PublicKey, 0)
	blsSignHash = storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[i], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err = core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.SendTxBlock(sp.SealKey, msgSealObject)

	// MirrorObject using names
	msgMirrorObject = storagetypes.NewMsgMirrorObject(user.GetAddr(), sdk.NewUint(0), bucketName, objectName)
	s.SendTxBlock(user, msgMirrorObject)
}

func (s *StorageTestSuite) TestMirrorGroup() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()}, "")
	s.SendTxBlock(owner, msgCreateGroup)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())

	// MirrorGroup using id
	msgMirrorGroup := storagetypes.NewMsgMirrorGroup(owner.GetAddr(), queryHeadGroupResp.GroupInfo.Id, "")
	s.SendTxBlock(owner, msgMirrorGroup)

	// CreateGroup
	groupName = storageutils.GenRandomGroupName()
	msgCreateGroup = storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, []sdk.AccAddress{member.GetAddr()}, "")
	s.SendTxBlock(owner, msgCreateGroup)

	// MirrorGroup using name
	msgMirrorGroup = storagetypes.NewMsgMirrorGroup(owner.GetAddr(), sdk.NewUint(0), groupName)
	s.SendTxBlock(owner, msgMirrorGroup)
}

func (s *StorageTestSuite) TestDiscontinueObject_Normal() {
	sp1, _, bucketName1, _, _, objectId1 := s.createObject()
	sp2, _, bucketName2, _, _, objectId2 := s.createObject()

	// DiscontinueObject
	msgDiscontinueObject := storagetypes.NewMsgDiscontinueObject(sp1.GcKey.GetAddr(), bucketName1, []sdkmath.Uint{objectId1}, "test")
	txRes1 := s.SendTxBlock(sp1.GcKey, msgDiscontinueObject)
	deleteAt1 := int64(filterDiscontinueObjectEventFromTx(txRes1).DeleteAt)

	time.Sleep(3 * time.Second)
	msgDiscontinueObject2 := storagetypes.NewMsgDiscontinueObject(sp2.GcKey.GetAddr(), bucketName2, []sdkmath.Uint{objectId2}, "test")
	txRes2 := s.SendTxBlock(sp2.GcKey, msgDiscontinueObject2)
	deleteAt2 := int64(filterDiscontinueObjectEventFromTx(txRes2).DeleteAt)

	// Wait after the delete timestamp for first discontinue request
	heightBefore := txRes1.Height
	heightAfter := int64(0)
	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt1)

		if blockTime >= deleteAt1 {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	time.Sleep(200 * time.Millisecond)
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

	// Wait after the delete timestamp for second discontinue request
	heightBefore = heightAfter
	heightAfter = int64(0)
	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt2)

		if blockTime >= deleteAt2 {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	time.Sleep(200 * time.Millisecond)
	events = make([]storagetypes.EventDeleteObject, 0)
	for heightBefore <= heightAfter {
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &heightBefore)
		s.Require().NoError(err)
		events = append(events, filterDeleteObjectEventFromBlock(blockRes)...)
		heightBefore++
	}
	for _, event := range events {
		if event.ObjectId.Equal(objectId2) {
			object2Found = true
		}
	}
	s.Require().True(object2Found)
}

func (s *StorageTestSuite) TestDiscontinueObject_UserDeleted() {
	sp, user, bucketName, _, objectName, objectId := s.createObject()

	// DiscontinueObject
	msgDiscontinueObject := storagetypes.NewMsgDiscontinueObject(sp.GcKey.GetAddr(), bucketName, []sdkmath.Uint{objectId}, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueObject)
	deleteAt := filterDiscontinueObjectEventFromTx(txRes).DeleteAt

	// DeleteObject before discontinue confirm window
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	txRes = s.SendTxBlock(user, msgDeleteObject)
	event := filterDeleteObjectEventFromTx(txRes)
	s.Require().Equal(event.ObjectId, objectId)

	// Wait after the delete timestamp
	heightBefore := txRes.Height
	heightAfter := int64(0)
	for {
		time.Sleep(200 * time.Millisecond)
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

	time.Sleep(200 * time.Millisecond)
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
	txRes1 := s.SendTxBlock(sp1.GcKey, msgDiscontinueBucket)
	deleteAt1 := filterDiscontinueBucketEventFromTx(txRes1).DeleteAt

	time.Sleep(3 * time.Second)
	msgDiscontinueBucket2 := storagetypes.NewMsgDiscontinueBucket(sp1.GcKey.GetAddr(), bucketName2, "test")
	txRes2 := s.SendTxBlock(sp2.GcKey, msgDiscontinueBucket2)
	deleteAt2 := filterDiscontinueBucketEventFromTx(txRes2).DeleteAt

	// Wait after the delete timestamp for the first discontinue request
	heightBefore := txRes1.Height
	heightAfter := int64(0)
	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt1)

		if blockTime >= deleteAt1 {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	time.Sleep(200 * time.Millisecond)
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

	// Wait after the delete timestamp for the second discontinue request
	heightBefore = heightAfter
	heightAfter = int64(0)
	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt2)

		if blockTime >= deleteAt2 {
			heightAfter = statusRes.SyncInfo.LatestBlockHeight
			break
		} else {
			heightBefore = statusRes.SyncInfo.LatestBlockHeight
		}
	}

	time.Sleep(200 * time.Millisecond)
	events = make([]storagetypes.EventDeleteBucket, 0)
	for heightBefore <= heightAfter {
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &heightBefore)
		s.Require().NoError(err)
		events = append(events, filterDeleteBucketEventFromBlock(blockRes)...)
		heightBefore++
	}

	for _, event := range events {
		if event.BucketId.Equal(bucketId2) {
			bucket2Found = true
		}
	}
	s.Require().True(bucket2Found)
}

func (s *StorageTestSuite) TestDiscontinueBucket_UserDeleted() {
	sp, user, bucketName, bucketId, objectName, _ := s.createObject()

	// DiscontinueBucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueBucket)
	deleteAt := int64(filterDiscontinueBucketEventFromTx(txRes).DeleteAt)

	// DeleteBucket before discontinue confirm window
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	txRes = s.SendTxBlock(user, msgDeleteBucket)
	event := filterDeleteBucketEventFromTx(txRes)
	s.Require().Equal(event.BucketId, bucketId)

	// Wait after the delete timestamp
	heightBefore := txRes.Height
	heightAfter := int64(0)
	for {
		time.Sleep(200 * time.Millisecond)
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

	time.Sleep(200 * time.Millisecond)
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

// createObject with default VISIBILITY_TYPE_PRIVATE
func (s *StorageTestSuite) createObject() (core.StorageProvider, keys.KeyManager, string, storagetypes.Uint, string, storagetypes.Uint) {
	return s.createObjectWithVisibility(storagetypes.VISIBILITY_TYPE_PRIVATE)
}

func (s *StorageTestSuite) createObjectWithVisibility(v storagetypes.VisibilityType) (core.StorageProvider, keys.KeyManager, string, storagetypes.Uint, string, storagetypes.Uint) {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, v, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, v)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), v, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Visibility, v)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user.GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvgId, nil)

	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[i], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig

	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

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
		if event.Type == "greenfield.storage.EventDiscontinueObject" {
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
		if event.Type == "greenfield.storage.EventDeleteObject" {
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
		if event.Type == "greenfield.storage.EventDeleteObject" {
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
		if event.Type == "greenfield.storage.EventDiscontinueBucket" {
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
		if event.Type == "greenfield.storage.EventDeleteBucket" {
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
		if event.Type == "greenfield.storage.EventDeleteBucket" {
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
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Creator, "")
	// CancelCreateObject
	msgCancelCreateObject := storagetypes.NewMsgCancelCreateObject(user.GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCancelCreateObject)
}

func (s *StorageTestSuite) TestCreateObjectWithCommonPrefix() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

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

	// CopyObject
	dstBucketName := bucketName
	dstObjectName := "new_directory/"
	msgCopyObject := storagetypes.NewMsgCopyObject(user.GetAddr(), bucketName, dstBucketName, objectName, dstObjectName, math.MaxUint, nil)
	msgCopyObject.DstPrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCopyObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCopyObject)

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
}

func (s *StorageTestSuite) TestUpdateParams() {
	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()
	queryParamsRequest := storagetypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.StorageQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)

	newParams := queryParamsResponse.GetParams()
	newParams.VersionedParams.MaxSegmentSize = 2048
	newParams.VersionedParams.MinChargeSize = 4096

	msgUpdateParams := &storagetypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    newParams,
	}

	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test", "test", "test",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(s.Validator, msgProposal)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query proposal and get proposal ID
	var proposalId uint64
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if attr.Key == "proposal_id" {
					proposalId, err = strconv.ParseUint(attr.Value, 10, 0)
					s.Require().NoError(err)
					break
				}
			}
			break
		}
	}
	s.Require().True(proposalId != 0)

	queryProposal := &govtypesv1.QueryProposalRequest{ProposalId: proposalId}
	_, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := govtypesv1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(1 * time.Second)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED)

	statusRes, err := s.TmClient.TmClient.Status(context.Background())
	s.Require().NoError(err)
	blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()
	queryVersionedParamsRequest := storagetypes.QueryParamsByTimestampRequest{Timestamp: blockTime}
	queryVersionedParamsResponse, err := s.Client.StorageQueryClient.QueryParamsByTimestamp(ctx, &queryVersionedParamsRequest)
	s.Require().NoError(err)
	require.EqualValues(s.T(), queryVersionedParamsResponse.GetParams().VersionedParams.MaxSegmentSize, 2048)
}

func (s *StorageTestSuite) TestCreateAndUpdateGroupExtraField() {
	var err error
	ctx := context.Background()
	owner := s.GenAndChargeAccounts(1, 1000000)[0]

	// Create a group without members
	testGroupName := "appName/bucketName"
	extra := "{\"description\":\"no description\",\"imageUrl\":\"www.images.com/image1\"}"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, nil, extra)
	s.SendTxBlock(owner, msgCreateGroup)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.Require().Equal(headGroupResponse.GroupInfo.Extra, extra)
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// Update the extra to empty
	newExtra := ""
	msgUpdateGroup := storagetypes.NewMsgUpdateGroupExtra(owner.GetAddr(), owner.GetAddr(), testGroupName, newExtra)
	s.SendTxBlock(owner, msgUpdateGroup)

	// Head Group
	headGroupRequest = storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err = s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.Require().Equal(newExtra, headGroupResponse.GroupInfo.Extra)
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// Update the extra
	newExtra = "something"
	msgUpdateGroup = storagetypes.NewMsgUpdateGroupExtra(owner.GetAddr(), owner.GetAddr(), testGroupName, newExtra)
	s.SendTxBlock(owner, msgUpdateGroup)

	// Head Group
	headGroupRequest = storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err = s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.Require().Equal(newExtra, headGroupResponse.GroupInfo.Extra)
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())
}

func (s *StorageTestSuite) TestRejectSealObject() {
	var err error
	// CreateBucket
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpId, sp.Info.Id)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Creator, "")
	// RejectSealObject
	msgRejectSealObject := storagetypes.NewMsgRejectUnsealedObject(sp.SealKey.GetAddr(), bucketName, objectName)
	s.SendTxBlock(sp.SealKey, msgRejectSealObject)

	// HeadObject
	queryHeadObjectRequest1 := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	_, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest1)
	s.Require().Error(err)
	s.Require().True(strings.Contains(err.Error(), storagetypes.ErrNoSuchObject.Error()))
}

func (s *StorageTestSuite) TestMigrationBucket() {
	// construct bucket and object
	primarySP := s.StorageProviders[0]
	gvg, found := primarySP.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	objectName := storageutils.GenRandomObjectName()
	_, _, _, bucketInfo := s.BaseSuite.CreateObject(user, &primarySP, gvg.Id, bucketName, objectName)

	var err error
	dstPrimarySP := s.CreateNewStorageProvider()

	// migrate bucket
	msgMigrationBucket := storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstPrimarySP.Info.Id)
	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
	msgMigrationBucket.DstPrimarySpApproval.Sig, err = dstPrimarySP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())
	s.SendTxBlock(user, msgMigrationBucket)
	s.Require().NoError(err)

	// cancel migration bucket
	msgCancelMigrationBucket := storagetypes.NewMsgCancelMigrateBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgCancelMigrationBucket)
	s.Require().NoError(err)

	// complete migration bucket
	var secondarySPIDs []uint32
	var secondarySPs []core.StorageProvider

	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySP.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
			secondarySPs = append(secondarySPs, ssp)
		}
		if len(secondarySPIDs) == 5 {
			break
		}
	}
	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(dstPrimarySP, 0, secondarySPIDs, 1)
	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &types2.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	dstGVG := gvgResp.GlobalVirtualGroup
	s.Require().True(found)

	// construct the signatures
	var gvgMappings []*storagetypes.GVGMapping
	gvgMappings = append(gvgMappings, &storagetypes.GVGMapping{SrcGlobalVirtualGroupId: gvg.Id, DstGlobalVirtualGroupId: dstGVG.Id})
	for _, gvgMapping := range gvgMappings {
		migrationBucketSignHash := storagetypes.NewSecondarySpMigrationBucketSignDoc(s.GetChainID(), bucketInfo.Id, dstPrimarySP.Info.Id, gvgMapping.SrcGlobalVirtualGroupId, gvgMapping.DstGlobalVirtualGroupId).GetBlsSignHash()
		secondarySigs := make([][]byte, 0)
		secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
		for _, ssp := range secondarySPs {
			sig, err := core.BlsSignAndVerify(ssp, migrationBucketSignHash)
			s.Require().NoError(err)
			secondarySigs = append(secondarySigs, sig)
			pk, err := bls.PublicKeyFromBytes(ssp.BlsKey.PubKey().Bytes())
			s.Require().NoError(err)
			secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
		}
		aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, migrationBucketSignHash, secondarySigs)
		s.Require().NoError(err)
		gvgMapping.SecondarySpBlsSignature = aggBlsSig
	}

	msgCompleteMigrationBucket := storagetypes.NewMsgCompleteMigrateBucket(dstPrimarySP.OperatorKey.GetAddr(), bucketName, dstGVG.FamilyId, gvgMappings)
	s.SendTxBlockWithExpectErrorString(msgCompleteMigrationBucket, dstPrimarySP.OperatorKey, "The bucket is not been migrating")

	// send again
	msgMigrationBucket = storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstPrimarySP.Info.Id)
	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
	msgMigrationBucket.DstPrimarySpApproval.Sig, err = dstPrimarySP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())
	s.SendTxBlock(user, msgMigrationBucket)
	s.Require().NoError(err)

	// complete again
	msgCompleteMigrationBucket = storagetypes.NewMsgCompleteMigrateBucket(dstPrimarySP.OperatorKey.GetAddr(), bucketName, dstGVG.FamilyId, gvgMappings)
	s.SendTxBlock(dstPrimarySP.OperatorKey, msgCompleteMigrationBucket)

}
