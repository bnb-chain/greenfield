package tests

import (
	"context"
	"math"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/bnb-chain/greenfield/e2e/core"
	storageutil "github.com/bnb-chain/greenfield/testutil/storage"
	types2 "github.com/bnb-chain/greenfield/types"
	"github.com/bnb-chain/greenfield/types/resource"
	"github.com/bnb-chain/greenfield/x/permission/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type PermissionTestSuite struct {
	core.BaseSuite
}

func (s *PermissionTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *PermissionTestSuite) SetupTest() {
}

func (s *StorageTestSuite) TestDeleteBucketPermission() {
	var err error
	user := s.GenAndChargeAccounts(2, 1000000)

	sp := s.StorageProviders[0]
	// CreateBucket
	bucketName := storageutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user[0].GetAddr(), bucketName, false, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.IsPublic, false)
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
		principal, []*types.Statement{statement})
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
		user[0].GetAddr(), bucketName, false, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
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
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.IsPublic, false)
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
		principal, []*types.Statement{statement})
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
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(user[1].GetAddr(), bucketName, 10000000, sdk.MustAccAddressFromHex(queryHeadBucketResponse.BucketInfo.PaymentAddress))
	s.SendTxBlock(msgUpdateBucketInfo, user[1])

	// Query BucketInfo
	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.Owner, user[0].GetAddr().String())
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.ReadQuota, uint64(10000000))
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
