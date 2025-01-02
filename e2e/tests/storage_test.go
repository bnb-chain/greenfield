package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	ctypes "github.com/cometbft/cometbft/rpc/core/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	types2 "github.com/bnb-chain/greenfield/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
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

var line = `1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,1234567890,
	1234567890,1234567890,1234567890,123`

func (s *StorageTestSuite) TestCreateBucket() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	queryQuotaUpdateTimeResponse, err := s.Client.QueryQuotaUpdateTime(ctx, &storagetypes.QueryQuoteUpdateTimeRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.CreateAt, queryQuotaUpdateTimeResponse.UpdateAt)

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

	// verify HeadBucketById
	queryHeadBucketResponseAfterUpdateBucket, err = s.Client.HeadBucketById(ctx, &storagetypes.QueryHeadBucketByIdRequest{BucketId: queryHeadBucketResponseAfterUpdateBucket.BucketInfo.Id.String()})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponseAfterUpdateBucket.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadBucketResponseAfterUpdateBucket.BucketInfo.BucketName, bucketName)

	// verify HeadBucketNFT
	headBucketNftResponse, err := s.Client.HeadBucketNFT(ctx, &storagetypes.QueryNFTRequest{
		TokenId: queryHeadBucketResponseAfterUpdateBucket.BucketInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(headBucketNftResponse.MetaData.BucketName, bucketName)

	// verify QueryIsPriceChanged
	isPriceChanged, err := s.Client.QueryIsPriceChanged(ctx, &storagetypes.QueryIsPriceChangedRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(isPriceChanged.Changed, false)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)
}

func (s *StorageTestSuite) TestCreateObject() {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
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

	// verify ListObjectsByBucketId
	queryListObjectsResponse, err = s.Client.ListObjectsByBucketId(ctx, &storagetypes.QueryListObjectsByBucketIdRequest{
		BucketId: queryHeadBucketResponse.BucketInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(len(queryListObjectsResponse.ObjectInfos), 1)
	s.Require().Equal(queryListObjectsResponse.ObjectInfos[0].ObjectName, objectName)

	// verify HeadObjectNFT
	headObjectNftResponse, err := s.Client.HeadObjectNFT(ctx, &storagetypes.QueryNFTRequest{
		TokenId: queryListObjectsResponse.ObjectInfos[0].Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(headObjectNftResponse.MetaData.ObjectName, objectName)

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

	// verify HeadObjectById
	queryHeadObjectAfterUpdateObjectResponse, err = s.Client.HeadObjectById(context.Background(), &storagetypes.QueryHeadObjectByIdRequest{ObjectId: queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Id.String()})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_INHERIT)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.ObjectName, objectName)

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
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())

	// 2.1. HeadGroupNFT
	headGroupNftResponse, err := s.Client.HeadGroupNFT(ctx, &storagetypes.QueryNFTRequest{
		TokenId: queryHeadGroupResp.GroupInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(headGroupNftResponse.MetaData.GroupName, groupName)

	// 3. ListGroup
	queryListGroupReq := storagetypes.QueryListGroupsRequest{GroupOwner: owner.GetAddr().String()}
	queryListGroupResp, err := s.Client.ListGroups(ctx, &queryListGroupReq)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(queryListGroupResp.GroupInfos), 1)

	// 4. UpdateGroupMember(add)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: member.GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// 4-1. HeadGroupMember(add)
	queryHeadGroupMemberReq := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	queryHeadGroupMemberResp, err := s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupMemberResp.GroupMember.GroupId, queryHeadGroupResp.GroupInfo.Id)

	// 5. UpdateGroupMember(delete)
	member2 := s.GenAndChargeAccounts(1, 1000000)[0]
	membersToAdd = []*storagetypes.MsgGroupMember{
		{Member: member2.GetAddr().String()},
	}
	membersToDelete = []sdk.AccAddress{member.GetAddr()}
	msgUpdateGroupMember = storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// 5-1. HeadGroupMember (delete)
	queryHeadGroupMemberReqDelete := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	_, err = s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReqDelete)
	s.Require().True(strings.Contains(err.Error(), storagetypes.ErrNoSuchGroupMember.Error()))

	// 6. Create a group with the same name
	msgCreateGroup = storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	s.SendTxBlockWithExpectErrorString(msgCreateGroup, owner, "exists")
}

func (s *StorageTestSuite) TestLeaveGroup() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: member.GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())

	// 2.1. HeadGroupNFT
	headGroupNftResponse, err := s.Client.HeadGroupNFT(ctx, &storagetypes.QueryNFTRequest{
		TokenId: queryHeadGroupResp.GroupInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(headGroupNftResponse.MetaData.GroupName, groupName)

	// 3. ListGroup
	queryListGroupReq := storagetypes.QueryListGroupsRequest{GroupOwner: owner.GetAddr().String()}
	queryListGroupResp, err := s.Client.ListGroups(ctx, &queryListGroupReq)
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
	membersToAdd = []*storagetypes.MsgGroupMember{
		{Member: member2.GetAddr().String()},
	}
	membersToDelete = []sdk.AccAddress{member.GetAddr()}
	msgUpdateGroupMember = storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// 5. leave group
	msgLeaveGroup := storagetypes.NewMsgLeaveGroup(member2.GetAddr(), owner.GetAddr(), groupName)
	s.SendTxBlock(member2, msgLeaveGroup)

	// 6. HeadGroupMember (leave)
	queryHeadGroupMemberReqDelete := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member2.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	_, err = s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReqDelete)
	s.Require().True(strings.Contains(err.Error(), storagetypes.ErrNoSuchGroupMember.Error()))
}

func (s *StorageTestSuite) TestDeleteBucket() {
	var err error
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	sp := s.BaseSuite.PickStorageProvider()
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
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
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
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user.GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// MirrorBucket using id
	msgMirrorBucket := storagetypes.NewMsgMirrorBucket(user.GetAddr(), sdk.ChainID(714), queryHeadBucketResponse.BucketInfo.Id, "")
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
	msgMirrorBucket = storagetypes.NewMsgMirrorBucket(user.GetAddr(), sdk.ChainID(714), sdk.NewUint(0), bucketName)
	s.SendTxBlock(user, msgMirrorBucket)
}

func (s *StorageTestSuite) TestMirrorObject() {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
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
	msgMirrorObject := storagetypes.NewMsgMirrorObject(user.GetAddr(), sdk.ChainID(714), queryHeadObjectResponse.ObjectInfo.Id, "", "")
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
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err = core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.SendTxBlock(sp.SealKey, msgSealObject)

	// MirrorObject using names
	msgMirrorObject = storagetypes.NewMsgMirrorObject(user.GetAddr(), sdk.ChainID(714), sdk.NewUint(0), bucketName, objectName)
	s.SendTxBlock(user, msgMirrorObject)
}

func (s *StorageTestSuite) TestMirrorGroup() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())

	// MirrorGroup using id
	msgMirrorGroup := storagetypes.NewMsgMirrorGroup(owner.GetAddr(), sdk.ChainID(714), queryHeadGroupResp.GroupInfo.Id, "")
	s.SendTxBlock(owner, msgMirrorGroup)

	// CreateGroup
	groupName = storageutils.GenRandomGroupName()
	msgCreateGroup = storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	s.SendTxBlock(owner, msgCreateGroup)

	// MirrorGroup using name
	msgMirrorGroup = storagetypes.NewMsgMirrorGroup(owner.GetAddr(), sdk.ChainID(714), sdk.NewUint(0), groupName)
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
	_ = s.SendTxBlock(sp.GcKey, msgDiscontinueObject)

	// DeleteObject before discontinue confirm window
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlockWithExpectErrorString(msgDeleteObject, user, "is discontined")
}

func (s *StorageTestSuite) TestDiscontinueBucket_Normal() {
	sp1, _, bucketName1, bucketId1, _, _ := s.createObject()
	sp2, _, bucketName2, bucketId2, _, _ := s.createObject()

	// DiscontinueBucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp1.GcKey.GetAddr(), bucketName1, "test")
	txRes1 := s.SendTxBlock(sp1.GcKey, msgDiscontinueBucket)
	deleteAt1 := filterDiscontinueBucketEventFromTx(txRes1).DeleteAt

	time.Sleep(3 * time.Second)
	msgDiscontinueBucket2 := storagetypes.NewMsgDiscontinueBucket(sp2.GcKey.GetAddr(), bucketName2, "test")
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

func (s *StorageTestSuite) GetSecondarySP(sps ...*core.StorageProvider) ([]*core.StorageProvider, []uint32) {
	var secondarySPs []*core.StorageProvider
	var secondarySPIDs []uint32

	for _, ssp := range s.StorageProviders {
		isSecondSP := true
		for _, sp := range sps {
			if ssp.Info.Id == sp.Info.Id {
				isSecondSP = false
				break
			}
		}
		if isSecondSP {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
			secondarySPs = append(secondarySPs, ssp)
		}
		if len(secondarySPIDs) == 6 {
			break
		}
	}
	return secondarySPs, secondarySPIDs
}

// createObject with default VISIBILITY_TYPE_PRIVATE
func (s *StorageTestSuite) createObject() (*core.StorageProvider, keys.KeyManager, string, storagetypes.Uint, string, storagetypes.Uint) {
	return s.createObjectWithVisibility(storagetypes.VISIBILITY_TYPE_PRIVATE)
}

func (s *StorageTestSuite) createObjectWithVisibility(v storagetypes.VisibilityType) (*core.StorageProvider, keys.KeyManager, string, storagetypes.Uint, string, storagetypes.Uint) {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
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
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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

	// Create a group
	testGroupName := "appName/bucketName"
	extra := "{\"description\":\"no description\",\"imageUrl\":\"www.images.com/image1\"}"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, extra)
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

func (s *StorageTestSuite) TestCreateAndRenewGroup() {
	var err error
	ctx := context.Background()
	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]

	// Create a group
	testGroupName := "appName/bucketName"
	extra := "{\"description\":\"no description\",\"imageUrl\":\"www.images.com/image1\"}"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, extra)
	s.SendTxBlock(owner, msgCreateGroup)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), testGroupName)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.Require().Equal(headGroupResponse.GroupInfo.Extra, extra)

	// Renew GroupMember
	expiration, err := time.Parse(time.RFC3339, "3023-12-31T23:59:59Z")
	s.Require().NoError(err)
	members := []*storagetypes.MsgGroupMember{
		{Member: member.GetAddr().String(), ExpirationTime: &expiration},
	}
	msgUpdateGroupMember := storagetypes.NewMsgRenewGroupMember(owner.GetAddr(), owner.GetAddr(), testGroupName, members)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// Head GroupMember
	queryHeadGroupMemberReq := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  testGroupName,
		GroupOwner: owner.GetAddr().String(),
	}
	queryHeadGroupMemberResp, err := s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupMemberResp.GroupMember.GroupId, headGroupResponse.GroupInfo.Id)
	s.Require().True(queryHeadGroupMemberResp.GroupMember.ExpirationTime.Equal(expiration))
}

func (s *StorageTestSuite) TestRejectSealObject() {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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

//func (s *StorageTestSuite) TestMigrationBucket() {
//	// construct bucket and object
//	primarySP := s.BaseSuite.PickStorageProvider()
//	gvg, found := primarySP.GetFirstGlobalVirtualGroup()
//	s.Require().True(found)
//	user := s.GenAndChargeAccounts(1, 1000000)[0]
//	bucketName := storageutils.GenRandomBucketName()
//	objectName := storageutils.GenRandomObjectName()
//	_, _, _, bucketInfo := s.BaseSuite.CreateObject(user, primarySP, gvg.Id, bucketName, objectName)
//
//	var err error
//	dstPrimarySP := s.CreateNewStorageProvider()
//
//	// migrate bucket
//	msgMigrationBucket := storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstPrimarySP.Info.Id)
//	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
//	msgMigrationBucket.DstPrimarySpApproval.Sig, err = dstPrimarySP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())
//	s.SendTxBlock(user, msgMigrationBucket)
//	s.Require().NoError(err)
//
//	// cancel migration bucket
//	msgCancelMigrationBucket := storagetypes.NewMsgCancelMigrateBucket(user.GetAddr(), bucketName)
//	s.SendTxBlock(user, msgCancelMigrationBucket)
//	s.Require().NoError(err)
//
//	// complete migration bucket
//	var secondarySPIDs []uint32
//	var secondarySPs []*core.StorageProvider
//
//	for _, ssp := range s.StorageProviders {
//		if ssp.Info.Id != primarySP.Info.Id {
//			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
//			secondarySPs = append(secondarySPs, ssp)
//		}
//		if len(secondarySPIDs) == 6 {
//			break
//		}
//	}
//	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(dstPrimarySP, 0, secondarySPIDs, 1)
//	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &types2.QueryGlobalVirtualGroupRequest{
//		GlobalVirtualGroupId: gvgID,
//	})
//	s.Require().NoError(err)
//	dstGVG := gvgResp.GlobalVirtualGroup
//	s.Require().True(found)
//
//	// construct the signatures
//	var gvgMappings []*storagetypes.GVGMapping
//	gvgMappings = append(gvgMappings, &storagetypes.GVGMapping{SrcGlobalVirtualGroupId: gvg.Id, DstGlobalVirtualGroupId: dstGVG.Id})
//	for _, gvgMapping := range gvgMappings {
//		migrationBucketSignHash := storagetypes.NewSecondarySpMigrationBucketSignDoc(s.GetChainID(), bucketInfo.Id, dstPrimarySP.Info.Id, gvgMapping.SrcGlobalVirtualGroupId, gvgMapping.DstGlobalVirtualGroupId).GetBlsSignHash()
//		secondarySigs := make([][]byte, 0)
//		secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
//		for _, ssp := range secondarySPs {
//			sig, err := core.BlsSignAndVerify(ssp, migrationBucketSignHash)
//			s.Require().NoError(err)
//			secondarySigs = append(secondarySigs, sig)
//			pk, err := bls.PublicKeyFromBytes(ssp.BlsKey.PubKey().Bytes())
//			s.Require().NoError(err)
//			secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
//		}
//		aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, migrationBucketSignHash, secondarySigs)
//		s.Require().NoError(err)
//		gvgMapping.SecondarySpBlsSignature = aggBlsSig
//	}
//
//	msgCompleteMigrationBucket := storagetypes.NewMsgCompleteMigrateBucket(dstPrimarySP.OperatorKey.GetAddr(), bucketName, dstGVG.FamilyId, gvgMappings)
//	s.SendTxBlockWithExpectErrorString(msgCompleteMigrationBucket, dstPrimarySP.OperatorKey, "The bucket is not been migrating")
//
//	// send again
//	msgMigrationBucket = storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstPrimarySP.Info.Id)
//	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
//	msgMigrationBucket.DstPrimarySpApproval.Sig, err = dstPrimarySP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())
//	s.SendTxBlock(user, msgMigrationBucket)
//	s.Require().NoError(err)
//
//	// complete again
//	msgCompleteMigrationBucket = storagetypes.NewMsgCompleteMigrateBucket(dstPrimarySP.OperatorKey.GetAddr(), bucketName, dstGVG.FamilyId, gvgMappings)
//	s.SendTxBlock(dstPrimarySP.OperatorKey, msgCompleteMigrationBucket)
//}

func (s *StorageTestSuite) TestUpdateStorageParams() {
	// 1. create proposal
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	queryParamsResp, err := s.Client.StorageQueryClient.Params(context.Background(), &storagetypes.QueryParamsRequest{})
	s.Require().NoError(err)

	updatedParams := queryParamsResp.Params
	updatedParams.MaxBucketsPerAccount = 10000
	msgUpdateParams := &storagetypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    updatedParams,
	}

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update storage params", "Test update storage params")
	s.Require().NoError(err)
	txBroadCastResp, err := s.SendTxBlockWithoutCheck(proposal, s.Validator)
	s.Require().NoError(err)
	s.T().Log("create proposal tx hash: ", txBroadCastResp.TxResponse.TxHash)

	// get proposal id
	proposalID := 0
	txResp, err := s.WaitForTx(txBroadCastResp.TxResponse.TxHash)
	s.Require().NoError(err)
	if txResp.Code == 0 && txResp.Height > 0 {
		for _, event := range txResp.Events {
			if event.Type == "submit_proposal" {
				proposalID, err = strconv.Atoi(event.GetAttributes()[0].Value)
				s.Require().NoError(err)
			}
		}
	}

	// 2. vote
	if proposalID == 0 {
		s.T().Errorf("proposalID is 0")
		return
	}
	s.T().Log("proposalID: ", proposalID)
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &types.TxOption{
		Mode:      &mode,
		Memo:      "",
		FeeAmount: sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
	}
	voteBroadCastResp, err := s.SendTxBlockWithoutCheckWithTxOpt(v1.NewMsgVote(s.Validator.GetAddr(), uint64(proposalID), v1.OptionYes, ""),
		s.Validator, txOpt)
	s.Require().NoError(err)
	voteResp, err := s.WaitForTx(voteBroadCastResp.TxResponse.TxHash)
	s.Require().NoError(err)
	s.T().Log("vote tx hash: ", voteResp.TxHash)
	if voteResp.Code > 0 {
		s.T().Errorf("voteTxResp.Code > 0")
		return
	}

	// 3. query proposal until it is end voting period
CheckProposalStatus:
	for {
		queryProposalResp, err := s.Client.Proposal(context.Background(), &v1.QueryProposalRequest{ProposalId: uint64(proposalID)})
		s.Require().NoError(err)
		if queryProposalResp.Proposal.Status != v1.StatusVotingPeriod {
			switch queryProposalResp.Proposal.Status {
			case v1.StatusDepositPeriod:
				s.T().Errorf("proposal deposit period")
				return
			case v1.StatusRejected:
				s.T().Errorf("proposal rejected")
				return
			case v1.StatusPassed:
				s.T().Logf("proposal passed")
				break CheckProposalStatus
			case v1.StatusFailed:
				s.T().Errorf("proposal failed, reason %s", queryProposalResp.Proposal.FailedReason)
				return
			}
		}
		time.Sleep(1 * time.Second)
	}

	// 4. check params updated
	err = s.WaitForNextBlock()
	s.Require().NoError(err)

	updatedQueryParamsResp, err := s.Client.StorageQueryClient.Params(context.Background(), &storagetypes.QueryParamsRequest{})
	s.Require().NoError(err)
	if reflect.DeepEqual(updatedQueryParamsResp.Params, updatedParams) {
		s.T().Logf("update params success")
	} else {
		s.T().Errorf("update params failed")
	}
}

// when a sp turn into maintenance mode, it should be able to create  bucket and object by its testing account.
func (s *StorageTestSuite) TestMaintenanceSPCreateBucketAndObject() {
	var err error
	ctx := context.Background()
	var sp *core.StorageProvider
	for _, tempSP := range s.BaseSuite.StorageProviders {
		exists, err := s.BaseSuite.ExistsSPMaintenanceRecords(tempSP.OperatorKey.GetAddr().String())
		s.Require().NoError(err)
		if !exists {
			sp = tempSP
			break
		}
	}
	spAddr := sp.OperatorKey.GetAddr()
	spMaintenanceAddr := sp.MaintenanceKey.GetAddr()

	req := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: spAddr.String(),
	}
	spResp, err := s.Client.StorageProviderByOperatorAddress(ctx, &req)
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_IN_SERVICE, spResp.StorageProvider.Status)

	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	msg := sptypes.NewMsgUpdateStorageProviderStatus(
		spAddr,
		sptypes.STATUS_IN_MAINTENANCE,
		1200,
	)
	txRes := s.SendTxBlock(sp.OperatorKey, msg)
	s.Require().Equal(txRes.Code, uint32(0))

	spResp, err = s.Client.StorageProviderByOperatorAddress(ctx, &req)
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_IN_MAINTENANCE, spResp.StorageProvider.Status)

	// create a bucket
	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		spMaintenanceAddr, bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(sp.MaintenanceKey, msgCreateBucket)

	// HeadBucket
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, spMaintenanceAddr.String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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
	msgCreateObject := storagetypes.NewMsgCreateObject(spMaintenanceAddr, bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(sp.MaintenanceKey, msgCreateObject)

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
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, spMaintenanceAddr.String())
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
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
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
		spMaintenanceAddr, bucketName, objectName, storagetypes.VISIBILITY_TYPE_INHERIT)
	s.Require().NoError(err)
	s.SendTxBlock(sp.MaintenanceKey, updateObjectInfo)
	s.Require().NoError(err)

	// verify modified objectinfo
	// head object
	queryHeadObjectAfterUpdateObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_INHERIT)

	// verify HeadObjectById
	queryHeadObjectAfterUpdateObjectResponse, err = s.Client.HeadObjectById(context.Background(), &storagetypes.QueryHeadObjectByIdRequest{ObjectId: queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Id.String()})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_INHERIT)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.ObjectName, objectName)

	// DeleteObject
	msgDeleteObject := storagetypes.NewMsgDeleteObject(spMaintenanceAddr, bucketName, objectName)
	s.SendTxBlock(sp.MaintenanceKey, msgDeleteObject)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(spMaintenanceAddr, bucketName)
	s.SendTxBlock(sp.MaintenanceKey, msgDeleteBucket)

	// revert back
	msg = sptypes.NewMsgUpdateStorageProviderStatus(
		spAddr,
		sptypes.STATUS_IN_SERVICE,
		0,
	)
	txRes = s.SendTxBlock(sp.OperatorKey, msg)
	s.Require().Equal(txRes.Code, uint32(0))
	spResp, err = s.Client.StorageProviderByOperatorAddress(ctx, &req)
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_IN_SERVICE, spResp.StorageProvider.Status)
}

//func (s *StorageTestSuite) TestRejectMigrateBucket() {
//	// construct bucket and object
//	primarySP := s.BaseSuite.PickStorageProvider()
//	gvg, found := primarySP.GetFirstGlobalVirtualGroup()
//	s.Require().True(found)
//	user := s.GenAndChargeAccounts(1, 1000000)[0]
//	bucketName := storageutils.GenRandomBucketName()
//	objectName := storageutils.GenRandomObjectName()
//	s.BaseSuite.CreateObject(user, primarySP, gvg.Id, bucketName, objectName)
//
//	var err error
//	dstPrimarySP := s.CreateNewStorageProvider()
//
//	// migrate bucket
//	msgMigrationBucket := storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstPrimarySP.Info.Id)
//	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
//	msgMigrationBucket.DstPrimarySpApproval.Sig, err = dstPrimarySP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())
//	s.SendTxBlock(user, msgMigrationBucket)
//	s.Require().NoError(err)
//
//	ctx := context.Background()
//	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
//		BucketName: bucketName,
//	}
//	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
//	s.Require().NoError(err)
//	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
//	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketStatus, storagetypes.BUCKET_STATUS_MIGRATING)
//
//	// Dest SP reject the migration
//	rejectMigration := storagetypes.NewMsgRejectMigrateBucket(dstPrimarySP.OperatorKey.GetAddr(), bucketName)
//	s.SendTxBlock(dstPrimarySP.OperatorKey, rejectMigration)
//	s.Require().NoError(err)
//
//	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
//		BucketName: bucketName,
//	}
//	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
//	s.Require().NoError(err)
//	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketStatus, storagetypes.BUCKET_STATUS_CREATED)
//
//	// migrate bucket again
//	msgMigrationBucket = storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstPrimarySP.Info.Id)
//	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
//	msgMigrationBucket.DstPrimarySpApproval.Sig, err = dstPrimarySP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())
//	s.SendTxBlock(user, msgMigrationBucket)
//	s.Require().NoError(err)
//
//	// cancel migration by user
//	msgCancelMigrationBucket := storagetypes.NewMsgCancelMigrateBucket(user.GetAddr(), bucketName)
//	s.SendTxBlock(user, msgCancelMigrationBucket)
//	s.Require().NoError(err)
//
//	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
//	s.Require().NoError(err)
//	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketStatus, storagetypes.BUCKET_STATUS_CREATED)
//
//	// dest SP should fail to reject
//	s.Client.SetKeyManager(dstPrimarySP.OperatorKey)
//	_, err = s.Client.BroadcastTx(context.Background(), []sdk.Msg{rejectMigration}, nil)
//	s.Require().Error(err)
//	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketStatus, storagetypes.BUCKET_STATUS_CREATED)
//}

func (s *StorageTestSuite) TestCreateBucketAndSetTag() {
	var err error
	user := s.GenAndChargeAccounts(1, 1000000)

	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	bucketName := storageutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)

	grn := types2.NewBucketGRN(bucketName)
	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	msgSetTag := storagetypes.NewMsgSetTag(user[0].GetAddr(), grn.String(), &tags)
	s.SendTxBlock(user[0], msgCreateBucket, msgSetTag)

	// HeadBucket
	req := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	resp, err := s.Client.HeadBucket(context.Background(), &req)
	s.Require().NoError(err)
	s.Require().Equal(tags, *resp.BucketInfo.Tags)
}

func (s *StorageTestSuite) TestCreateObjectAndSetTag() {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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

	grn := types2.NewObjectGRN(bucketName, objectName)
	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	msgSetTag := storagetypes.NewMsgSetTag(user.GetAddr(), grn.String(), &tags)
	s.SendTxBlock(user, msgCreateObject, msgSetTag)

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
	s.Require().Equal(*queryHeadObjectResponse.ObjectInfo.Tags, tags)

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
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

	// verify ListObjectsByBucketId
	queryListObjectsResponse, err = s.Client.ListObjectsByBucketId(ctx, &storagetypes.QueryListObjectsByBucketIdRequest{
		BucketId: queryHeadBucketResponse.BucketInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(len(queryListObjectsResponse.ObjectInfos), 1)
	s.Require().Equal(queryListObjectsResponse.ObjectInfos[0].ObjectName, objectName)

	// verify HeadObjectNFT
	headObjectNftResponse, err := s.Client.HeadObjectNFT(ctx, &storagetypes.QueryNFTRequest{
		TokenId: queryListObjectsResponse.ObjectInfos[0].Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(headObjectNftResponse.MetaData.ObjectName, objectName)

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

	// verify HeadObjectById
	queryHeadObjectAfterUpdateObjectResponse, err = s.Client.HeadObjectById(context.Background(), &storagetypes.QueryHeadObjectByIdRequest{ObjectId: queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Id.String()})
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_INHERIT)
	s.Require().Equal(queryHeadObjectAfterUpdateObjectResponse.ObjectInfo.ObjectName, objectName)

	// DeleteObject
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)
}

func (s *StorageTestSuite) TestCreateGroupAndSetTag() {
	ctx := context.Background()

	owner := s.GenAndChargeAccounts(1, 1000000)[0]
	member := s.GenAndChargeAccounts(1, 1000000)[0]
	groupName := storageutils.GenRandomGroupName()

	// 1. CreateGroup
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	grn := types2.NewGroupGRN(owner.GetAddr(), groupName)
	var tags storagetypes.ResourceTags
	tags.Tags = append(tags.Tags, storagetypes.ResourceTags_Tag{Key: "key1", Value: "value1"})
	msgSetTag := storagetypes.NewMsgSetTag(owner.GetAddr(), grn.String(), &tags)
	s.SendTxBlock(owner, msgCreateGroup, msgSetTag)
	s.T().Logf("CerateGroup success, owner: %s, group name: %s", owner.GetAddr().String(), groupName)

	// 2. HeadGroup
	queryHeadGroupReq := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: groupName}
	queryHeadGroupResp, err := s.Client.HeadGroup(ctx, &queryHeadGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.GroupName, groupName)
	s.Require().Equal(queryHeadGroupResp.GroupInfo.Owner, owner.GetAddr().String())
	s.Require().Equal(*queryHeadGroupResp.GroupInfo.Tags, tags)

	// 2.1. HeadGroupNFT
	headGroupNftResponse, err := s.Client.HeadGroupNFT(ctx, &storagetypes.QueryNFTRequest{
		TokenId: queryHeadGroupResp.GroupInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(headGroupNftResponse.MetaData.GroupName, groupName)

	// 3. ListGroup
	queryListGroupReq := storagetypes.QueryListGroupsRequest{GroupOwner: owner.GetAddr().String()}
	queryListGroupResp, err := s.Client.ListGroups(ctx, &queryListGroupReq)
	s.Require().NoError(err)
	s.Require().GreaterOrEqual(len(queryListGroupResp.GroupInfos), 1)

	// 4. UpdateGroupMember(add)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: member.GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// 4-1. HeadGroupMember(add)
	queryHeadGroupMemberReq := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	queryHeadGroupMemberResp, err := s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReq)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadGroupMemberResp.GroupMember.GroupId, queryHeadGroupResp.GroupInfo.Id)

	// 5. UpdateGroupMember(delete)
	member2 := s.GenAndChargeAccounts(1, 1000000)[0]
	membersToAdd = []*storagetypes.MsgGroupMember{
		{Member: member2.GetAddr().String()},
	}
	membersToDelete = []sdk.AccAddress{member.GetAddr()}
	msgUpdateGroupMember = storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), groupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// 5-1. HeadGroupMember (delete)
	queryHeadGroupMemberReqDelete := storagetypes.QueryHeadGroupMemberRequest{
		Member:     member.GetAddr().String(),
		GroupName:  groupName,
		GroupOwner: owner.GetAddr().String(),
	}
	_, err = s.Client.HeadGroupMember(ctx, &queryHeadGroupMemberReqDelete)
	s.Require().True(strings.Contains(err.Error(), storagetypes.ErrNoSuchGroupMember.Error()))

	// 6. Create a group with the same name
	msgCreateGroup = storagetypes.NewMsgCreateGroup(owner.GetAddr(), groupName, "")
	s.SendTxBlockWithExpectErrorString(msgCreateGroup, owner, "exists")
}

func (s *StorageTestSuite) TestDeleteCreateObject_InCreatedStatus() {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
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
	msgDeleteCreateObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteCreateObject)

	_, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().EqualError(err, "rpc error: code = Unknown desc = No such object: unknown request")
}

func (s *StorageTestSuite) TestDisallowChangePaymentAccount() {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
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
	_, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	// create a new payment account
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: user.GetAddr().String(),
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	// query user's payment accounts
	queryGetPaymentAccountsByOwnerRequest := paymenttypes.QueryPaymentAccountsByOwnerRequest{
		Owner: user.GetAddr().String(),
	}
	paymentAccounts, err := s.Client.PaymentAccountsByOwner(ctx, &queryGetPaymentAccountsByOwnerRequest)
	s.Require().NoError(err)
	s.T().Log(paymentAccounts)
	s.Require().Equal(1, len(paymentAccounts.PaymentAccounts))
	paymentAccountAddr := sdk.MustAccAddressFromHex(paymentAccounts.PaymentAccounts[0])

	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAccountAddr.String(),
		Amount:  types.NewIntFromInt64WithDecimal(2, types.DecimalBNB),
	}
	_ = s.SendTxBlock(user, msgDeposit)

	// UpdateBucketInfo is fine for no created object
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAccountAddr, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgUpdateBucketInfo)
	s.Require().NoError(err)

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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize),
		storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
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

	// UpdateBucketInfo is not fine for there is a created object
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "has unseald objects")

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvgId, nil)

	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig

	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

	// UpdateBucketInfo is fine for there is no created object
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.Require().NoError(err)
	s.SendTxBlock(user, msgUpdateBucketInfo)
	s.Require().NoError(err)
}

func (s *StorageTestSuite) TestToggleBucketSpAsDelegatedAgents() {
	var err error
	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
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

	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	ctx := context.Background()
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(false, queryHeadBucketResponse.BucketInfo.SpAsDelegatedAgentDisabled)

	MsgToggleSPAsDelegatedAgent := storagetypes.NewMsgToggleSPAsDelegatedAgent(
		user.GetAddr(),
		bucketName)
	s.SendTxBlock(user, MsgToggleSPAsDelegatedAgent)

	// HeadBucket
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(true, queryHeadBucketResponse.BucketInfo.SpAsDelegatedAgentDisabled)
}

func (s *StorageTestSuite) TestCreateObjectByDelegatedAgents() {
	var err error
	ctx := context.Background()

	// CreateBucket
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	bucketOwner := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := storageutils.GenRandomBucketName()
	objectName := storageutils.GenRandomObjectName()

	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		bucketOwner.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)

	s.SendTxBlock(bucketOwner, msgCreateBucket)

	// HeadBucket
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(false, queryHeadBucketResponse.BucketInfo.SpAsDelegatedAgentDisabled)

	// DelegateCreate for user2, who does not have permission
	var buffer bytes.Buffer
	// Create 1MiB content where each line contains 1024 characters.
	for i := 0; i < 1024; i++ {
		buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	payloadSize := buffer.Len()
	contextType := "text/event-stream"
	msgDelegateCreateObject := storagetypes.NewMsgDelegateCreateObject(
		sp.OperatorKey.GetAddr(),
		bucketOwner.GetAddr(),
		bucketName,
		objectName,
		uint64(payloadSize),
		storagetypes.VISIBILITY_TYPE_PRIVATE,
		nil,
		contextType,
		storagetypes.REDUNDANCY_EC_TYPE)
	s.SendTxBlock(sp.OperatorKey, msgDelegateCreateObject)

	headObjectReq := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	headObjectResp, err := s.Client.HeadObject(ctx, &headObjectReq)
	s.Require().NoError(err)
	s.Require().Equal(objectName, headObjectResp.ObjectInfo.ObjectName)
	s.Require().Equal(bucketOwner.GetAddr().String(), headObjectResp.ObjectInfo.Owner)
	s.Require().Equal(0, len(headObjectResp.ObjectInfo.Checksums))

	// SP seal object, and update the object checksum
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}

	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObjectV2(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil, expectChecksum)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, headObjectResp.ObjectInfo.Id, storagetypes.GenerateHash(expectChecksum[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

	headObjectResp, err = s.Client.HeadObject(ctx, &headObjectReq)
	s.Require().NoError(err)
	s.Require().Equal(objectName, headObjectResp.ObjectInfo.ObjectName)
	s.Require().Equal(bucketOwner.GetAddr().String(), headObjectResp.ObjectInfo.Owner)
	s.Require().Equal(expectChecksum, headObjectResp.ObjectInfo.Checksums)

	// delegate update
	var newBuffer bytes.Buffer
	for i := 0; i < 2048; i++ {
		newBuffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
	}
	newPayloadSize := uint64(newBuffer.Len())
	newChecksum := sdk.Keccak256(newBuffer.Bytes())
	newExpectChecksum := [][]byte{newChecksum, newChecksum, newChecksum, newChecksum, newChecksum, newChecksum, newChecksum}

	msgUpdateObject := storagetypes.NewMsgDelegateUpdateObjectContent(sp.OperatorKey.GetAddr(),
		bucketOwner.GetAddr(), bucketName, objectName, newPayloadSize, nil)
	s.SendTxBlock(sp.OperatorKey, msgUpdateObject)
	s.T().Logf("msgUpdateObject %s", msgUpdateObject.String())

	// every secondary sp signs the checksums
	newSecondarySigs := make([][]byte, 0)
	newBlsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, headObjectResp.ObjectInfo.Id, storagetypes.GenerateHash(newExpectChecksum[:])).GetBlsSignHash()
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], newBlsSignHash)
		s.Require().NoError(err)
		newSecondarySigs = append(newSecondarySigs, sig)
	}
	aggBlsSig, err = core.BlsAggregateAndVerify(secondarySPBlsPubKeys, newBlsSignHash, newSecondarySigs)
	s.Require().NoError(err)
	msgSealObject = storagetypes.NewMsgSealObjectV2(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil, newExpectChecksum)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.T().Logf("msgSealObject %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)
}
