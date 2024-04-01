package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"

	types2 "github.com/bnb-chain/greenfield/sdk/types"

	storageutils "github.com/bnb-chain/greenfield/testutil/storage"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

func (s *StorageTestSuite) enableMessage() {
	msgSetBucketFlowRateLimit := sdk.MsgTypeURL(&storagetypes.MsgSetBucketFlowRateLimit{})
	msgMigrateBuketGasParams := gashubtypes.NewMsgGasParamsWithFixedGas(msgSetBucketFlowRateLimit, 1.2e3)

	msgUpdateGasParams := gashubtypes.NewMsgSetMsgGasParams(authtypes.NewModuleAddress(govtypes.ModuleName).String(), []*gashubtypes.MsgGasParams{msgMigrateBuketGasParams}, nil)

	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()

	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateGasParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types2.NewIntFromInt64WithDecimal(100, types2.DecimalBNB))},
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
}

func (s *StorageTestSuite) TestSetBucketRateLimitToZero() {
	s.enableMessage()

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

	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	s.Require().Equal(queryHeadBucketResponse.ExtraInfo.IsRateLimited, true)
	s.Require().Equal(queryHeadBucketResponse.ExtraInfo.FlowRateLimit.String(), "0")

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
	s.enableMessage()

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

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000000))
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
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// create object
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
	s.enableMessage()

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

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// UpdateBucketInfo
	var readQuota uint64 = 100
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlock(user, msgUpdateBucketInfo)

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	s.Require().Equal(queryHeadBucketResponse.ExtraInfo.IsRateLimited, true)

	// update bucket
	readQuota = 101
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "payment account is not changed but the bucket is limited")

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	s.Require().Equal(queryHeadBucketResponse.ExtraInfo.IsRateLimited, false)

	// update bucket
	readQuota = 102
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlock(user, msgUpdateBucketInfo)
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
	s.enableMessage()

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

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(s.User.GetAddr(), s.User.GetAddr(), s.User.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(s.User, msgSetBucketRateLimit)

	// SetBucketRateLimit
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAcc.GetAddr(), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "the flow rate limit is not set")

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(0))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	s.Require().Equal(queryHeadBucketResponse.ExtraInfo.IsRateLimited, true)

	// UpdateBucketInfo
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAcc.GetAddr(), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "greater than the flow rate limit")

	// SetBucketRateLimit
	msgSetBucketRateLimit = storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// UpdateBucketInfo
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, nil, paymentAcc.GetAddr(), storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlock(user, msgUpdateBucketInfo)

	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	s.Require().Equal(queryHeadBucketResponse.ExtraInfo.IsRateLimited, false)
}

func (s *StorageTestSuite) TestQueryBucketRateLimit() {
	s.enableMessage()

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

	// SetBucketRateLimit
	msgSetBucketRateLimit := storagetypes.NewMsgSetBucketFlowRateLimit(paymentAcc.GetAddr(), s.User.GetAddr(), paymentAcc.GetAddr(), bucketName, sdkmath.NewInt(100000000000000))
	s.SendTxBlock(paymentAcc, msgSetBucketRateLimit)

	// update bucket
	var readQuota uint64 = 100
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, nil, storagetypes.VISIBILITY_TYPE_PUBLIC_READ)
	s.SendTxBlock(user, msgUpdateBucketInfo)

	s.Require().NoError(err)

	// QueryBucketRateLimit
	ctx := context.Background()
	queryBucketRateLimitRequest := storagetypes.QueryPaymentAccountBucketFlowRateLimitRequest{
		PaymentAccount: paymentAcc.GetAddr().String(),
		BucketName:     bucketName,
		BucketOwner:    user.GetAddr().String(),
	}
	queryBucketRateLimitResponse, err := s.Client.QueryPaymentAccountBucketFlowRateLimit(ctx, &queryBucketRateLimitRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryBucketRateLimitResponse.IsSet, true)
	s.Require().Equal(queryBucketRateLimitResponse.FlowRateLimit, sdkmath.NewInt(100000000000000))
}
