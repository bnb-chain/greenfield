package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	sdktype "github.com/bnb-chain/greenfield/sdk/types"
	storageutil "github.com/bnb-chain/greenfield/testutil/storage"
	types2 "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/common"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/x/permission/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *StorageTestSuite) TestDeleteBucketPermission() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// verify permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_DELETE_BUCKET,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())

	// Put bucket policy
	statement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(user[0], msgPutPolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_DELETE_BUCKET,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_ALLOW)

	// Query bucket policy
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user[1].GetAddr(), bucketName)
	s.SendTxBlock(user[1], msgDeleteBucket)
}

func (s *StorageTestSuite) TestDeletePolicy() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// verify permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())

	// Put bucket policy
	statement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_UPDATE_BUCKET_INFO, types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(user[0], msgPutPolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_ALLOW)

	// Query bucket policy
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)

	// update read quota
	chargedReadQuota := uint64(100000)
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(user[1].GetAddr(), bucketName, &chargedReadQuota,
		sdk.MustAccAddressFromHex(queryHeadBucketResponse.BucketInfo.PaymentAddress), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlock(user[1], msgUpdateBucketInfo)

	// Query BucketInfo
	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.ChargedReadQuota, uint64(100000))
	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(user[0], msgDeletePolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
}

func (s *StorageTestSuite) TestCreateObjectByOthers() {
	var err error
	user := s.GenAndChargeAccounts(3, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.T().Logf("BucketInfo: %s", queryHeadBucketResponse.BucketInfo.String())

	// verify permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())

	// Put object policy
	statement1 := &types.Statement{
		Actions: []types.ActionType{types.ACTION_CREATE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	statement2 := &types.Statement{
		Actions: []types.ActionType{types.ACTION_UPDATE_OBJECT_INFO},
		Effect:  types.EFFECT_ALLOW,
	}
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement1, statement2}, nil)
	s.SendTxBlock(user[0], msgPutPolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_ALLOW)

	// Query bucket policy
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)

	// CreateObject
	objectName := storageutil.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize),
		storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[1], msgCreateObject)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.T().Logf("ObjectInfo: %s", queryHeadObjectResponse.ObjectInfo.String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.PayloadSize, uint64(payloadSize))
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Creator, user[1].GetAddr().String())

	// CancelCreateObject
	msgCancelCreateObject := storagetypes.NewMsgCancelCreateObject(user[2].GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgCancelCreateObject, user[2], "Only allowed owner/creator to do cancel create object")

	// CancelCreateObject
	msgCancelCreateObject = storagetypes.NewMsgCancelCreateObject(user[1].GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlock(user[1], msgCancelCreateObject)

	// CreateObject
	msgCreateObject = storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize),
		storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[1], msgCreateObject)

	// HeadObject
	queryHeadObjectRequest = storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.T().Logf("ObjectInfo: %s", queryHeadObjectResponse.ObjectInfo.String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.PayloadSize, uint64(payloadSize))
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Checksums, expectChecksum)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.RedundancyType, storagetypes.REDUNDANCY_EC_TYPE)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ContentType, contextType)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.Creator, user[1].GetAddr().String())

	// Owner cancel
	msgCancelCreateObject = storagetypes.NewMsgCancelCreateObject(user[0].GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCancelCreateObject)

	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(user[0], msgDeletePolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
}

func (s *StorageTestSuite) TestCreateObjectByOthersExpiration() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.T().Logf("BucketInfo: %s", queryHeadBucketResponse.BucketInfo.String())

	// verify permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())

	// Put bucket policy
	statement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_CREATE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	// Add 5 seconds to the current time, because current BlockTime is later than the current time about 3 sec
	expirationTime := time.Now().UTC().Add(5 * time.Second)
	s.T().Logf("Time now: %s", expirationTime.String())
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, &expirationTime)
	s.SendTxBlock(user[0], msgPutPolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_ALLOW)

	// Query bucket policy
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)

	time.Sleep(5 * time.Second)
	// CreateObject
	objectName := storageutil.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())
	s.SendTxBlockWithExpectErrorString(msgCreateObject, user[1], "has no CreateObject permission of the bucket")

	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(user[0], msgDeletePolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
}

func (s *StorageTestSuite) TestCreateObjectByOthersLimitSize() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)
	s.T().Logf("BucketInfo: %s", queryHeadBucketResponse.BucketInfo.String())

	// verify permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())

	// Put bucket policy, create object size limit to 2K
	statement := &types.Statement{
		Actions:   []types.ActionType{types.ACTION_CREATE_OBJECT},
		Effect:    types.EFFECT_ALLOW,
		LimitSize: &common.UInt64Value{Value: 1.5 * 1024 * 1024},
	}
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(user[0], msgPutPolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s, err %s", verifyPermReq.String(), verifyPermResp.String(), err)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_ALLOW)

	// Query bucket policy
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.T().Logf("Policy: %s", queryPolicyForAccountResp.Policy.String())
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)
	s.Require().Equal(queryPolicyForAccountResp.Policy.Statements[0].LimitSize, statement.LimitSize)

	// CreateObject
	objectName := storageutil.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())
	s.SendTxBlock(user[1], msgCreateObject)

	objectName2 := storageutil.GenRandomObjectName()
	msgCreateObject = storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName2, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())
	s.SendTxBlockWithExpectErrorString(msgCreateObject, user[1], "has no CreateObject permission of the bucket")

	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(user[0], msgDeletePolicy)

	// verify permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_CREATE_OBJECT,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
}

func (s *StorageTestSuite) TestGrantsPermissionToGroup() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// verify deny permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())

	// Create Group
	testGroupName := "testGroup"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(user[0].GetAddr(), testGroupName, "")
	s.SendTxBlock(user[0], msgCreateGroup)

	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: user[1].GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(user[0].GetAddr(), user[0].GetAddr(), testGroupName, membersToAdd, membersToDelete)
	s.SendTxBlock(user[0], msgUpdateGroupMember)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: user[0].GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(user[0].GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// Head Group member
	headGroupMemberRequest := storagetypes.QueryHeadGroupMemberRequest{Member: user[1].GetAddr().String(), GroupOwner: user[0].GetAddr().String(), GroupName: testGroupName}
	headGroupMemberResponse, err := s.Client.HeadGroupMember(ctx, &headGroupMemberRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupMemberResponse.GroupMember.GroupId, headGroupResponse.GetGroupInfo().Id)

	// Put bucket policy
	statement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_UPDATE_BUCKET_INFO, types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	principal := types.NewPrincipalWithGroupInfo(user[0].GetAddr(), headGroupResponse.GroupInfo.GroupName)
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(user[0], msgPutPolicy)

	// verify allow permission
	verifyPermReq = storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[1].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_UPDATE_BUCKET_INFO,
	}
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_ALLOW)

	// Query policy for group
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForGroupReq := storagetypes.QueryPolicyForGroupRequest{
		Resource:         grn.String(),
		PrincipalGroupId: headGroupResponse.GroupInfo.Id.String(),
	}
	queryPolicyForGroupResp, err := s.Client.QueryPolicyForGroup(ctx, &queryPolicyForGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)
	s.Require().Equal(queryPolicyForGroupResp.Policy.Statements[0].Effect, types.EFFECT_ALLOW)
}

func (s *StorageTestSuite) TestVisibilityPermission() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket bucket0:public bucket1:private bucket2:default
	bucketName0 := storageutil.GenRandomBucketName()
	bucketName1 := storageutil.GenRandomBucketName()
	bucketName2 := storageutil.GenRandomBucketName()
	buckets := []struct {
		BucketName string
		PublicType storagetypes.VisibilityType
	}{
		{
			BucketName: bucketName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
		},
		{
			BucketName: bucketName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
		},
		{
			BucketName: bucketName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
		},
	}

	for _, bucket := range buckets {
		msgCreateBucket := storagetypes.NewMsgCreateBucket(
			user[0].GetAddr(), bucket.BucketName, bucket.PublicType, sp.OperatorKey.GetAddr(),
			nil, math.MaxUint, nil, 0)
		msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
		msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxBlock(user[0], msgCreateBucket)
	}

	// object0:public object1:private object2:default
	objectName0 := storageutil.GenRandomObjectName()
	objectName1 := storageutil.GenRandomObjectName()
	objectName2 := storageutil.GenRandomObjectName()

	objects := []struct {
		BucketName string
		ObjectName string
		PublicType storagetypes.VisibilityType
		Effect     types.Effect
	}{
		{
			BucketName: bucketName0,
			ObjectName: objectName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName0,
			ObjectName: objectName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName0,
			ObjectName: objectName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName1,
			ObjectName: objectName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName1,
			ObjectName: objectName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName1,
			ObjectName: objectName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName2,
			ObjectName: objectName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName2,
			ObjectName: objectName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName2,
			ObjectName: objectName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_DENY,
		},
	}

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

	// Create content contains 1024 characters.
	buffer.WriteString(line)
	payloadSize := buffer.Len()
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"

	ctx := context.Background()

	for _, object := range objects {
		msgCreateObject0 := storagetypes.NewMsgCreateObject(user[0].GetAddr(), object.BucketName, object.ObjectName, uint64(payloadSize), object.PublicType, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
		msgCreateObject0.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject0.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxBlock(user[0], msgCreateObject0)

		// verify permission
		verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
			Operator:   user[1].GetAddr().String(),
			BucketName: object.BucketName,
			ObjectName: object.ObjectName,
			ActionType: types.ACTION_GET_OBJECT,
		}
		verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
		s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
		s.Require().NoError(err)
		s.Require().Equal(verifyPermResp.Effect, object.Effect)
	}
}

func (s *StorageTestSuite) TestEmptyPermission() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	ctx := context.Background()

	// CreateBucket bucket0:public bucket1:private bucket2:default
	bucketName0 := storageutil.GenRandomBucketName()
	bucketName1 := storageutil.GenRandomBucketName()
	bucketName2 := storageutil.GenRandomBucketName()
	buckets := []struct {
		BucketName string
		PublicType storagetypes.VisibilityType
		Effect     types.Effect
	}{
		{
			BucketName: bucketName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_DENY,
		},
	}

	for _, bucket := range buckets {
		msgCreateBucket := storagetypes.NewMsgCreateBucket(
			user[0].GetAddr(), bucket.BucketName, bucket.PublicType, sp.OperatorKey.GetAddr(),
			nil, math.MaxUint, nil, 0)
		msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
		msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxBlock(user[0], msgCreateBucket)

		// verify permission for empty operator address
		verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
			Operator:   "",
			BucketName: bucket.BucketName,
			ActionType: types.ACTION_GET_OBJECT,
		}
		verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
		s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
		s.Require().NoError(err)
		s.Require().Equal(verifyPermResp.Effect, bucket.Effect)
	}

	// object0:public object1:private object2:default
	objectName0 := storageutil.GenRandomObjectName()
	objectName1 := storageutil.GenRandomObjectName()
	objectName2 := storageutil.GenRandomObjectName()

	objects := []struct {
		BucketName string
		ObjectName string
		PublicType storagetypes.VisibilityType
		Effect     types.Effect
	}{
		{
			BucketName: bucketName0,
			ObjectName: objectName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName0,
			ObjectName: objectName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName0,
			ObjectName: objectName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName1,
			ObjectName: objectName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName1,
			ObjectName: objectName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName1,
			ObjectName: objectName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName2,
			ObjectName: objectName0,
			PublicType: storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
			Effect:     types.EFFECT_ALLOW,
		},
		{
			BucketName: bucketName2,
			ObjectName: objectName1,
			PublicType: storagetypes.VISIBILITY_TYPE_PRIVATE,
			Effect:     types.EFFECT_DENY,
		},
		{
			BucketName: bucketName2,
			ObjectName: objectName2,
			PublicType: storagetypes.VISIBILITY_TYPE_INHERIT,
			Effect:     types.EFFECT_DENY,
		},
	}

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

	// Create content contains 1024 characters.
	buffer.WriteString(line)
	payloadSize := buffer.Len()
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"

	for _, object := range objects {
		msgCreateObject0 := storagetypes.NewMsgCreateObject(user[0].GetAddr(), object.BucketName, object.ObjectName, uint64(payloadSize), object.PublicType, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
		msgCreateObject0.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject0.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxBlock(user[0], msgCreateObject0)

		// verify permission
		verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
			Operator:   "",
			BucketName: object.BucketName,
			ObjectName: object.ObjectName,
			ActionType: types.ACTION_GET_OBJECT,
		}
		verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
		s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
		s.Require().NoError(err)
		s.Require().Equal(verifyPermResp.Effect, object.Effect)
	}
}

// When resources are deleted, policies which associated with personal account(address) and resources(Bucket and Object)
// will also be garbage collected.
func (s *StorageTestSuite) TestStalePermissionForAccountGC() {
	var err error
	ctx := context.Background()
	user1 := s.GenAndChargeAccounts(1, 1000000)[0]

	_, owner, bucketName, bucketId, objectName, objectId := s.createObjectWithVisibility(storagetypes.VISIBILITY_TYPE_PUBLIC_READ)

	principal := types.NewPrincipalWithAccount(user1.GetAddr())

	// Put bucket policy
	bucketStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutBucketPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{bucketStatement}, nil)
	s.SendTxBlock(owner, msgPutBucketPolicy)

	// Put Object policy
	objectStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutObjectPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewObjectGRN(bucketName, objectName).String(),
		principal, []*types.Statement{objectStatement}, nil)
	s.SendTxBlock(owner, msgPutObjectPolicy)

	// Query the policy which is enforced on bucket and object
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user1.GetAddr().String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(bucketId, queryPolicyForAccountResp.Policy.ResourceId)
	bucketPolicyId := queryPolicyForAccountResp.Policy.Id

	grn2 := types2.NewObjectGRN(bucketName, objectName)
	queryPolicyForAccountResp, err = s.Client.QueryPolicyForAccount(ctx, &storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn2.String(),
		PrincipalAddress: user1.GetAddr().String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(objectId, queryPolicyForAccountResp.Policy.ResourceId)
	objectPolicyId := queryPolicyForAccountResp.Policy.Id
	s.T().Log(queryPolicyForAccountResp.Policy.String())

	// user1 deletes the object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user1.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user1, msgDeleteObject)

	// user1 deletes the bucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user1.GetAddr(), bucketName)
	s.SendTxBlock(user1, msgDeleteBucket)

	// bucket and object dont exist after deletion
	headObjectReq := storagetypes.QueryHeadObjectRequest{
		BucketName: objectName,
	}
	_, err = s.Client.HeadObject(ctx, &headObjectReq)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such object")

	headBucketReq := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	_, err = s.Client.HeadBucket(ctx, &headBucketReq)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such bucket")

	// policy is GC
	_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: bucketPolicyId.String()})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such Policy")

	_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: objectPolicyId.String()})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such Policy")
}

func (s *StorageTestSuite) TestDeleteObjectPolicy() {
	var err error
	ctx := context.Background()
	user1 := s.GenAndChargeAccounts(1, 1000000)[0]

	_, owner, bucketName, _, objectName, objectId := s.createObjectWithVisibility(storagetypes.VISIBILITY_TYPE_PUBLIC_READ)

	principal := types.NewPrincipalWithAccount(user1.GetAddr())

	// Put bucket policy
	bucketStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutBucketPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{bucketStatement}, nil)
	s.SendTxBlock(owner, msgPutBucketPolicy)

	// Put Object policy
	objectStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutObjectPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewObjectGRN(bucketName, objectName).String(),
		principal, []*types.Statement{objectStatement}, nil)
	s.SendTxBlock(owner, msgPutObjectPolicy)

	// Query the policy which is enforced on bucket and object
	grn1 := types2.NewObjectGRN(bucketName, objectName)
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn1.String(),
		PrincipalAddress: user1.GetAddr().String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(objectId, queryPolicyForAccountResp.Policy.ResourceId)

	// Delete object policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(owner.GetAddr(), grn1.String(), types.NewPrincipalWithAccount(user1.GetAddr()))
	s.SendTxBlock(owner, msgDeletePolicy)

	// verify permission
	verifyPermReq := storagetypes.QueryVerifyPermissionRequest{
		Operator:   user1.GetAddr().String(),
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: types.ACTION_DELETE_OBJECT,
	}
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &verifyPermReq)
	s.T().Logf("resp: %s, rep %s", verifyPermReq.String(), verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(verifyPermResp.Effect, types.EFFECT_DENY)
}

func (s *StorageTestSuite) TestDeleteGroupPolicy() {
	var err error
	ctx := context.Background()

	user := s.GenAndChargeAccounts(4, 1000000)
	owner := user[0]
	_ = s.BaseSuite.PickStorageProvider()

	// Create Group
	testGroupName := "testGroup"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: user[1].GetAddr().String()},
		{Member: user[2].GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), testGroupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// Put policy
	groupStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_UPDATE_GROUP_MEMBER},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutGroupPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewGroupGRN(owner.GetAddr(), testGroupName).String(),
		types.NewPrincipalWithAccount(user[1].GetAddr()), []*types.Statement{groupStatement}, nil)
	s.SendTxBlock(owner, msgPutGroupPolicy)

	// Query for policy
	grn := types2.NewGroupGRN(owner.GetAddr(), testGroupName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}

	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_GROUP)
	s.T().Logf("policy is %s", queryPolicyForAccountResp.Policy.String())

	// Delete policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(owner.GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(owner, msgDeletePolicy)

	// verify permission
	_, err = s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such Policy")
}

// When resources are deleted, policies which associated with group and resources(Bucket and Object)
// will also be garbage collected.
func (s *StorageTestSuite) TestStalePermissionForGroupGC() {
	ctx := context.Background()
	user := s.GenAndChargeAccounts(3, 10000)
	_, owner, bucketName, bucketId, objectName, objectId := s.createObjectWithVisibility(storagetypes.VISIBILITY_TYPE_PUBLIC_READ)

	// Create Group
	testGroupName := "testGroup"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: user[1].GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), testGroupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	principal := types.NewPrincipalWithGroupId(headGroupResponse.GroupInfo.Id)
	// Put bucket policy for group
	bucketStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutBucketPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{bucketStatement}, nil)
	s.SendTxBlock(owner, msgPutBucketPolicy)

	// Put Object policy for group
	objectStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutObjectPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewObjectGRN(bucketName, objectName).String(),
		principal, []*types.Statement{objectStatement}, nil)
	s.SendTxBlock(owner, msgPutObjectPolicy)

	// Query bucket policy for group
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForGroupReq := storagetypes.QueryPolicyForGroupRequest{
		Resource:         grn.String(),
		PrincipalGroupId: headGroupResponse.GroupInfo.Id.String(),
	}

	queryPolicyForGroupResp, err := s.Client.QueryPolicyForGroup(ctx, &queryPolicyForGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(bucketId, queryPolicyForGroupResp.Policy.ResourceId)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(types.EFFECT_ALLOW, queryPolicyForGroupResp.Policy.Statements[0].Effect)
	bucketPolicyId := queryPolicyForGroupResp.Policy.Id

	// Query object policy for group
	grn2 := types2.NewObjectGRN(bucketName, objectName)
	queryPolicyForGroupResp, err = s.Client.QueryPolicyForGroup(ctx, &storagetypes.QueryPolicyForGroupRequest{
		Resource:         grn2.String(),
		PrincipalGroupId: headGroupResponse.GroupInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(objectId, queryPolicyForGroupResp.Policy.ResourceId)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceType, resource.RESOURCE_TYPE_OBJECT)
	s.Require().Equal(types.EFFECT_ALLOW, queryPolicyForGroupResp.Policy.Statements[0].Effect)
	objectPolicyId := queryPolicyForGroupResp.Policy.Id

	// user1 deletes the object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user[1].GetAddr(), bucketName, objectName)
	s.SendTxBlock(user[1], msgDeleteObject)

	// user1 deletes the bucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user[1].GetAddr(), bucketName)
	s.SendTxBlock(user[1], msgDeleteBucket)

	// bucket and object dont exist after deletion
	headObjectReq := storagetypes.QueryHeadObjectRequest{
		BucketName: objectName,
	}
	_, err = s.Client.HeadObject(ctx, &headObjectReq)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such object")

	headBucketReq := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	_, err = s.Client.HeadBucket(ctx, &headBucketReq)
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such bucket")

	// policy is GC
	_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: objectPolicyId.String()})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such Policy")

	_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: bucketPolicyId.String()})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such Policy")
}

// When a group is deleted, a. Policies associated with group members and group, b. group members
// will be garbage collected.
func (s *StorageTestSuite) TestGroupMembersAndPolicyGC() {
	var err error
	ctx := context.Background()

	user := s.GenAndChargeAccounts(4, 1000000)
	owner := user[0]
	_ = s.BaseSuite.PickStorageProvider()

	// Create Group
	testGroupName := "testGroup"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: user[1].GetAddr().String()},
		{Member: user[2].GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), testGroupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// Put policy
	groupStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_UPDATE_GROUP_MEMBER},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutGroupPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewGroupGRN(owner.GetAddr(), testGroupName).String(),
		types.NewPrincipalWithAccount(user[1].GetAddr()), []*types.Statement{groupStatement}, nil)
	s.SendTxBlock(owner, msgPutGroupPolicy)

	// Query for policy
	grn := types2.NewGroupGRN(owner.GetAddr(), testGroupName)
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{
		Resource:         grn.String(),
		PrincipalAddress: user[1].GetAddr().String(),
	}

	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_GROUP)
	s.T().Logf("policy is %s", queryPolicyForAccountResp.Policy.String())
	policyID := queryPolicyForAccountResp.Policy.Id

	// Head Group member
	headGroupMemberRequest := storagetypes.QueryHeadGroupMemberRequest{Member: user[2].GetAddr().String(), GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupMemberResponse, err := s.Client.HeadGroupMember(ctx, &headGroupMemberRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupMemberResponse.GroupMember.GroupId, headGroupResponse.GetGroupInfo().Id)

	// list group
	queryListGroupReq := storagetypes.QueryListGroupsRequest{GroupOwner: owner.GetAddr().String()}
	queryListGroupResp, err := s.Client.ListGroups(ctx, &queryListGroupReq)
	s.Require().NoError(err)
	s.T().Log(queryListGroupResp.String())

	// the owner deletes the group
	msgDeleteGroup := storagetypes.NewMsgDeleteGroup(owner.GetAddr(), testGroupName)
	s.SendTxBlock(owner, msgDeleteGroup)

	// policy is GC
	_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: policyID.String()})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such Policy")
}

func (s *StorageTestSuite) TestExceedEachBlockLimitGC() {
	var err error
	ctx := context.Background()
	owner := s.GenAndChargeAccounts(1, 10000)[0]
	user := s.GenAndChargeAccounts(1, 10000)[0]
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	s.Client.SetKeyManager(owner)

	nonce, _ := s.Client.GetNonce(ctx)
	bucketNames := make([]string, 0)

	// Create 250 Buckets
	bucketNumber := 250

	feeAmt := sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(int64(15000000000000))))
	txOpt := sdktype.TxOption{
		NoSimulate: true,
		GasLimit:   3000,
		FeeAmount:  feeAmt,
	}

	for i := 0; i < bucketNumber; i++ {
		txOpt.Nonce = nonce
		bucketName := storageutil.GenRandomBucketName()
		bucketNames = append(bucketNames, bucketName)
		msgCreateBucket := storagetypes.NewMsgCreateBucket(
			owner.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
			nil, math.MaxUint, nil, 0)
		msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
		msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxWithTxOpt(msgCreateBucket, owner, txOpt)
		nonce++
	}
	err = s.WaitForNextBlock()
	s.Require().NoError(err)

	// ListBuckets
	queryListBucketsRequest := storagetypes.QueryListBucketsRequest{}
	queryListBucketResponse, err := s.Client.ListBuckets(ctx, &queryListBucketsRequest)
	s.Require().NoError(err)
	s.T().Logf("number of buckes is %d", queryListBucketResponse.Pagination.Total)

	principal := types.NewPrincipalWithAccount(user.GetAddr())

	for i := 0; i < bucketNumber; i++ {
		txOpt.Nonce = nonce
		// Put bucket policy
		bucketStatement := &types.Statement{
			Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
			Effect:  types.EFFECT_ALLOW,
		}
		msgPutBucketPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewBucketGRN(bucketNames[i]).String(),
			principal, []*types.Statement{bucketStatement}, nil)
		s.SendTxWithTxOpt(msgPutBucketPolicy, owner, txOpt)
		nonce++
	}
	// wait for 2 blocks
	for i := 0; i < 2; i++ {
		_ = s.WaitForNextBlock()
	}

	policyIds := make([]sdkmath.Uint, 0)
	// policies are present for buckets
	for i := 0; i < bucketNumber; i++ {
		queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &storagetypes.QueryPolicyForAccountRequest{
			Resource:         types2.NewBucketGRN(bucketNames[i]).String(),
			PrincipalAddress: user.GetAddr().String(),
		})
		s.Require().NoError(err)
		policyIds = append(policyIds, queryPolicyForAccountResp.Policy.Id)
	}

	// delete batch of buckets
	for i := 0; i < bucketNumber; i++ {
		txOpt.Nonce = nonce
		// the owner deletes buckets
		msgDeleteBucket := storagetypes.NewMsgDeleteBucket(owner.GetAddr(), bucketNames[i])
		s.SendTxWithTxOpt(msgDeleteBucket, owner, txOpt)
		nonce++
	}

	// Garbage collection wont be done within the block since the total number of policies to be deleted exceed the
	// handling ability of each block
	notAllPoliciesGC := false
	for i := 0; i < bucketNumber; i++ {
		_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: policyIds[i].String()})
		if err == nil {
			// if there is at least 1 policy still exist, that means GC is not fully done yet.
			notAllPoliciesGC = true
		}
	}
	s.Require().True(notAllPoliciesGC)

	// wait for another 2 block, all policies should be GC
	for i := 0; i < 2; i++ {
		_ = s.WaitForNextBlock()
	}

	for i := 0; i < bucketNumber; i++ {
		// policy is GC
		_, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: policyIds[i].String()})
		s.Require().Error(err)
		s.Require().ErrorContains(err, "No such Policy")
	}
}

func (s *StorageTestSuite) TestUpdateGroupExtraWithPermission() {
	var err error
	ctx := context.Background()

	user := s.GenAndChargeAccounts(4, 1000000)
	owner := user[0]

	// Create Group
	testGroupName := "testGroup"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, "")
	s.SendTxBlock(owner, msgCreateGroup)
	membersToAdd := []*storagetypes.MsgGroupMember{
		{Member: user[1].GetAddr().String()},
		{Member: user[2].GetAddr().String()},
	}
	membersToDelete := []sdk.AccAddress{}
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), testGroupName, membersToAdd, membersToDelete)
	s.SendTxBlock(owner, msgUpdateGroupMember)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// Put policy
	groupStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_UPDATE_GROUP_EXTRA},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutGroupPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewGroupGRN(owner.GetAddr(), testGroupName).String(),
		types.NewPrincipalWithAccount(user[1].GetAddr()), []*types.Statement{groupStatement}, nil)
	s.SendTxBlock(owner, msgPutGroupPolicy)

	// user1 update the extra of group is allowed
	newExtra := "newExtra"
	msgUpdateGroup := storagetypes.NewMsgUpdateGroupExtra(user[1].GetAddr(), owner.GetAddr(), testGroupName, newExtra)
	s.SendTxBlock(user[1], msgUpdateGroup)

	// Head Group
	headGroupRequest = storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err = s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.Require().Equal(newExtra, headGroupResponse.GroupInfo.Extra)
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	// user2 update the extra of group is not allowed
	newExtra = "newExtra2"
	msgUpdateGroup2 := storagetypes.NewMsgUpdateGroupExtra(user[2].GetAddr(), owner.GetAddr(), testGroupName, newExtra)
	_, err = s.SendTxBlockWithoutCheck(msgUpdateGroup2, user[2])
	s.Require().Error(err)
}

func (s *StorageTestSuite) TestPutPolicy_ObjectWithSlash() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	ctx := context.Background()
	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user[0], msgCreateBucket)

	// HeadBucket
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.GlobalVirtualGroupFamilyId, gvg.FamilyId)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PaymentAddress, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Visibility, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.SourceType, storagetypes.SOURCE_TYPE_ORIGIN)

	// CreateObject
	objectName := "test/" + storageutil.GenRandomObjectName()
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user[0].GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())

	time.Sleep(3 * time.Second)
	s.SendTxBlock(user[0], msgCreateObject)

	time.Sleep(3 * time.Second)
	// Put object policy
	statement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_GET_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewObjectGRN(bucketName, objectName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(user[0], msgPutPolicy)
}

func (s *StorageTestSuite) TestVerifyStaleGroupPermission() {
	ctx := context.Background()

	// set the params, not to delete stale policy
	queryParamsRequest := storagetypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.StorageQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)

	newParams := queryParamsResponse.GetParams()
	newParams.StalePolicyCleanupMax = 1
	s.UpdateParams(&newParams)

	defer func() {
		newParams.StalePolicyCleanupMax = 100
		s.UpdateParams(&newParams)
	}()

	user := s.GenAndChargeAccounts(3, 10000)
	_, owner, bucketName, bucketId, objectName, objectId := s.createObjectWithVisibility(storagetypes.VISIBILITY_TYPE_PUBLIC_READ)

	// Create Group with 3 group member
	testGroupName := "testGroup"
	msgCreateGroup := storagetypes.NewMsgCreateGroup(owner.GetAddr(), testGroupName, "")
	msgUpdateGroupMember := storagetypes.NewMsgUpdateGroupMember(owner.GetAddr(), owner.GetAddr(), testGroupName,
		[]*storagetypes.MsgGroupMember{
			{
				Member: user[0].GetAddr().String(),
			}, {
				Member: user[1].GetAddr().String(),
			}, {
				Member: user[2].GetAddr().String(),
			},
		},
		[]sdk.AccAddress{})
	s.SendTxBlock(owner, msgCreateGroup, msgUpdateGroupMember)

	// Head Group
	headGroupRequest := storagetypes.QueryHeadGroupRequest{GroupOwner: owner.GetAddr().String(), GroupName: testGroupName}
	headGroupResponse, err := s.Client.HeadGroup(ctx, &headGroupRequest)
	s.Require().NoError(err)
	s.Require().Equal(headGroupResponse.GroupInfo.GroupName, testGroupName)
	s.Require().True(owner.GetAddr().Equals(sdk.MustAccAddressFromHex(headGroupResponse.GroupInfo.Owner)))
	s.T().Logf("GroupInfo: %s", headGroupResponse.GetGroupInfo().String())

	principal := types.NewPrincipalWithGroupId(headGroupResponse.GroupInfo.Id)
	// Put bucket policy for group
	bucketStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_BUCKET},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutBucketPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{bucketStatement}, nil)
	s.SendTxBlock(owner, msgPutBucketPolicy)

	// Put Object policy for group
	objectStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutObjectPolicy := storagetypes.NewMsgPutPolicy(owner.GetAddr(), types2.NewObjectGRN(bucketName, objectName).String(),
		principal, []*types.Statement{objectStatement}, nil)
	s.SendTxBlock(owner, msgPutObjectPolicy)

	// Query bucket policy for group
	grn := types2.NewBucketGRN(bucketName)
	queryPolicyForGroupReq := storagetypes.QueryPolicyForGroupRequest{
		Resource:         grn.String(),
		PrincipalGroupId: headGroupResponse.GroupInfo.Id.String(),
	}

	queryPolicyForGroupResp, err := s.Client.QueryPolicyForGroup(ctx, &queryPolicyForGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(bucketId, queryPolicyForGroupResp.Policy.ResourceId)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(types.EFFECT_ALLOW, queryPolicyForGroupResp.Policy.Statements[0].Effect)
	bucketPolicyID := queryPolicyForGroupResp.Policy.Id

	// Query object policy for group
	grn2 := types2.NewObjectGRN(bucketName, objectName)
	queryPolicyForGroupResp, err = s.Client.QueryPolicyForGroup(ctx, &storagetypes.QueryPolicyForGroupRequest{
		Resource:         grn2.String(),
		PrincipalGroupId: headGroupResponse.GroupInfo.Id.String(),
	})
	s.Require().NoError(err)
	s.Require().Equal(objectId, queryPolicyForGroupResp.Policy.ResourceId)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceType, resource.RESOURCE_TYPE_OBJECT)
	s.Require().Equal(types.EFFECT_ALLOW, queryPolicyForGroupResp.Policy.Statements[0].Effect)
	objectPolicyID := queryPolicyForGroupResp.Policy.Id

	// verify group policy
	verifyPermResp, err := s.Client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[2].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_DELETE_BUCKET,
	})
	s.T().Logf("Verify Bucket Permission, %s", verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(types.EFFECT_ALLOW, verifyPermResp.Effect)
	// verify group policy
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[2].GetAddr().String(),
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: types.ACTION_DELETE_OBJECT,
	})
	s.T().Logf("Verify Object Permission, %s", verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(types.EFFECT_ALLOW, verifyPermResp.Effect)

	// user1 deletes the group
	msgDeleteGroup := storagetypes.NewMsgDeleteGroup(owner.GetAddr(), testGroupName)
	s.SendTxBlock(owner, msgDeleteGroup)

	// group don't exist after deletion
	_, err = s.Client.HeadGroup(ctx, &storagetypes.QueryHeadGroupRequest{
		GroupOwner: owner.GetAddr().String(),
		GroupName:  testGroupName,
	})
	s.Require().Error(err)
	s.Require().ErrorContains(err, "No such group")

	// stale permission is still exist
	queryPolicyByIDResp, err := s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: bucketPolicyID.String()})
	s.T().Logf("Qyery policy by id resp: %s", queryPolicyByIDResp)
	s.Require().NoError(err)

	queryPolicyByIDResp, err = s.Client.QueryPolicyById(ctx, &storagetypes.QueryPolicyByIdRequest{PolicyId: objectPolicyID.String()})
	s.T().Logf("Qyery policy by id resp: %s", queryPolicyByIDResp)
	s.Require().NoError(err)

	// verify group policy
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[2].GetAddr().String(),
		BucketName: bucketName,
		ActionType: types.ACTION_DELETE_BUCKET,
	})
	s.T().Logf("Verify Bucket Permission, %s", verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(types.EFFECT_DENY, verifyPermResp.Effect)
	// verify group policy
	verifyPermResp, err = s.Client.VerifyPermission(ctx, &storagetypes.QueryVerifyPermissionRequest{
		Operator:   user[2].GetAddr().String(),
		BucketName: bucketName,
		ObjectName: objectName,
		ActionType: types.ACTION_DELETE_OBJECT,
	})
	s.T().Logf("Verify Object Permission, %s", verifyPermResp.String())
	s.Require().NoError(err)
	s.Require().Equal(types.EFFECT_DENY, verifyPermResp.Effect)
}

func (s *StorageTestSuite) UpdateParams(newParams *storagetypes.Params) {
	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()

	msgUpdateParams := &storagetypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    *newParams,
	}

	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, sdktype.NewIntFromInt64WithDecimal(100, sdktype.DecimalBNB))},
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

	queryParamsResponse, err := s.Client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	s.Require().NoError(err)
	s.T().Logf("QueryParmas: %s", queryParamsResponse.Params.String())
	s.Require().Equal(queryParamsResponse.Params, *newParams)
}

func (s *StorageTestSuite) TestGrantsPermissionToObjectWithWildcardInName() {
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.BaseSuite.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	bucketName := storageutil.GenRandomBucketName()
	objectName := "*.jpg"
	s.CreateObject(user[0], sp, gvg.Id, bucketName, objectName)

	// grant permission to *.jpg
	objectStatement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_DELETE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	msgPutObjectPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewObjectGRN(bucketName, objectName).String(), types.NewPrincipalWithAccount(user[1].GetAddr()), []*types.Statement{objectStatement}, nil)
	s.SendTxBlock(user[0], msgPutObjectPolicy)

	// Delete object
	s.SendTxBlock(user[1], storagetypes.NewMsgDeleteObject(user[1].GetAddr(), bucketName, objectName))

	// head object
	_, err := s.Client.HeadObject(context.Background(), &storagetypes.QueryHeadObjectRequest{BucketName: bucketName, ObjectName: objectName})
	s.Require().True(strings.Contains(err.Error(), "No such object"))
}
