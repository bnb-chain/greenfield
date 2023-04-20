package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"

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

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user[0])

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
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
	s.SendTxBlock(msgPutPolicy, user[0])

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
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{Resource: grn.String(),
		PrincipalAddress: user[1].GetAddr().String()}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user[1].GetAddr(), bucketName)
	s.SendTxBlock(msgDeleteBucket, user[1])

}

func (s *StorageTestSuite) TestDeletePolicy() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user[0])

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
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
	s.SendTxBlock(msgPutPolicy, user[0])

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
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{Resource: grn.String(),
		PrincipalAddress: user[1].GetAddr().String()}
	queryPolicyForAccountResp, err := s.Client.QueryPolicyForAccount(ctx, &queryPolicyForAccountReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForAccountResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)

	// update read quota
	chargedReadQuota := uint64(100000)
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(user[1].GetAddr(), bucketName, &chargedReadQuota,
		sdk.MustAccAddressFromHex(queryHeadBucketResponse.BucketInfo.PaymentAddress), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlock(msgUpdateBucketInfo, user[1])

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
	s.SendTxBlock(msgDeletePolicy, user[0])

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

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user[0])

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
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

	// Put bucket policy
	statement := &types.Statement{
		Actions: []types.ActionType{types.ACTION_CREATE_OBJECT},
		Effect:  types.EFFECT_ALLOW,
	}
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(msgPutPolicy, user[0])

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
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{Resource: grn.String(),
		PrincipalAddress: user[1].GetAddr().String()}
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
		storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user[1])

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
	s.Require().Equal(len(queryHeadObjectResponse.ObjectInfo.SecondarySpAddresses), 1)

	// CancelCreateObject
	msgCancelCreateObject := storagetypes.NewMsgCancelCreateObject(user[2].GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgCancelCreateObject, user[2], "Only allowed owner/creator to do cancel create object")

	// CancelCreateObject
	msgCancelCreateObject = storagetypes.NewMsgCancelCreateObject(user[1].GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlock(msgCancelCreateObject, user[1])

	// CreateObject
	msgCreateObject = storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize),
		storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user[1])

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
	s.Require().Equal(len(queryHeadObjectResponse.ObjectInfo.SecondarySpAddresses), 1)

	// Owner cancel
	msgCancelCreateObject = storagetypes.NewMsgCancelCreateObject(user[0].GetAddr(), bucketName, objectName)
	s.Require().NoError(err)
	s.SendTxBlock(msgCancelCreateObject, user[0])

	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(msgDeletePolicy, user[0])

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

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user[0])

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
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
	expirationTime := time.Now().UTC()
	s.T().Logf("Time now: %s", expirationTime.String())
	principal := types.NewPrincipalWithAccount(user[1].GetAddr())
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, &expirationTime)
	s.SendTxBlock(msgPutPolicy, user[0])

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
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{Resource: grn.String(),
		PrincipalAddress: user[1].GetAddr().String()}
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())
	s.SendTxBlockWithExpectErrorString(msgCreateObject, user[1], "has no CreateObject permission of the bucket")

	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(msgDeletePolicy, user[0])

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

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user[0])

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
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
	s.SendTxBlock(msgPutPolicy, user[0])

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
	queryPolicyForAccountReq := storagetypes.QueryPolicyForAccountRequest{Resource: grn.String(),
		PrincipalAddress: user[1].GetAddr().String()}
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
	msgCreateObject := storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())
	s.SendTxBlock(msgCreateObject, user[1])

	objectName2 := storageutil.GenRandomObjectName()
	msgCreateObject = storagetypes.NewMsgCreateObject(user[1].GetAddr(), bucketName, objectName2, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PUBLIC_READ, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.T().Logf("Message: %s", msgCreateObject.String())
	s.SendTxBlockWithExpectErrorString(msgCreateObject, user[1], "has no CreateObject permission of the bucket")

	// Delete bucket Policy
	msgDeletePolicy := storagetypes.NewMsgDeletePolicy(user[0].GetAddr(), grn.String(), types.NewPrincipalWithAccount(user[1].GetAddr()))
	s.SendTxBlock(msgDeletePolicy, user[0])

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

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user[0])

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.PrimarySpAddress, sp.OperatorKey.GetAddr().String())
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
	msgCreateGroup := storagetypes.NewMsgCreateGroup(user[0].GetAddr(), testGroupName, []sdk.AccAddress{user[1].GetAddr()})
	s.SendTxBlock(msgCreateGroup, user[0])

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
	principal := types.NewPrincipalWithGroup(headGroupResponse.GroupInfo.Id)
	msgPutPolicy := storagetypes.NewMsgPutPolicy(user[0].GetAddr(), types2.NewBucketGRN(bucketName).String(),
		principal, []*types.Statement{statement}, nil)
	s.SendTxBlock(msgPutPolicy, user[0])

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
	queryPolicyForGroupReq := storagetypes.QueryPolicyForGroupRequest{Resource: grn.String(),
		PrincipalGroupId: headGroupResponse.GroupInfo.Id.String()}
	queryPolicyForGroupResp, err := s.Client.QueryPolicyForGroup(ctx, &queryPolicyForGroupReq)
	s.Require().NoError(err)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceType, resource.RESOURCE_TYPE_BUCKET)
	s.Require().Equal(queryPolicyForGroupResp.Policy.ResourceId, queryHeadBucketResponse.BucketInfo.Id)
	s.Require().Equal(queryPolicyForGroupResp.Policy.Statements[0].Effect, types.EFFECT_ALLOW)
}

func (s *StorageTestSuite) TestVisibilityPermission() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.StorageProviders[0]
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
		msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxBlock(msgCreateBucket, user[0])
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
		msgCreateObject0 := storagetypes.NewMsgCreateObject(user[0].GetAddr(), object.BucketName, object.ObjectName, uint64(payloadSize), object.PublicType, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
		msgCreateObject0.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject0.GetApprovalBytes())
		s.Require().NoError(err)
		s.SendTxBlock(msgCreateObject0, user[0])

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
