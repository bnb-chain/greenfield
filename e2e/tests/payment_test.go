package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/samber/lo"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	storagetestutils "github.com/bnb-chain/greenfield/testutil/storage"
	"github.com/bnb-chain/greenfield/types/common"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type StreamRecords struct {
	User      paymenttypes.StreamRecord
	GVGFamily paymenttypes.StreamRecord
	GVG       paymenttypes.StreamRecord
	Tax       paymenttypes.StreamRecord
}

type PaymentTestSuite struct {
	core.BaseSuite
}

func (s *PaymentTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *PaymentTestSuite) SetupTest() {}

func (s *PaymentTestSuite) TestPaymentAccount() {
	user := s.GenAndChargeAccounts(1, 100)[0]
	ctx := context.Background()
	// create a new payment account
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: user.GetAddr().String(),
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	// query user's payment accounts
	queryGetPaymentAccountsByOwnerRequest := paymenttypes.QueryGetPaymentAccountsByOwnerRequest{
		Owner: user.GetAddr().String(),
	}
	paymentAccounts, err := s.Client.GetPaymentAccountsByOwner(ctx, &queryGetPaymentAccountsByOwnerRequest)
	s.Require().NoError(err)
	s.T().Log(paymentAccounts)
	s.Require().Equal(1, len(paymentAccounts.PaymentAccounts))
	paymentAccountAddr := paymentAccounts.PaymentAccounts[0]
	// query this payment account
	queryGetPaymentAccountRequest := paymenttypes.QueryGetPaymentAccountRequest{
		Addr: paymentAccountAddr,
	}
	paymentAccount, err := s.Client.PaymentAccount(ctx, &queryGetPaymentAccountRequest)
	s.Require().NoError(err)
	s.T().Logf("payment account: %s", core.YamlString(paymentAccount.PaymentAccount))
	s.Require().Equal(user.GetAddr().String(), paymentAccount.PaymentAccount.Owner)
	s.Require().Equal(true, paymentAccount.PaymentAccount.Refundable)
	// set this payment account to non-refundable
	msgDisableRefund := &paymenttypes.MsgDisableRefund{
		Owner: user.GetAddr().String(),
		Addr:  paymentAccountAddr,
	}
	_ = s.SendTxBlock(user, msgDisableRefund)
	// query this payment account
	paymentAccount, err = s.Client.PaymentAccount(ctx, &queryGetPaymentAccountRequest)
	s.Require().NoError(err)
	s.T().Logf("payment account: %s", core.YamlString(paymentAccount.PaymentAccount))
	s.Require().Equal(false, paymentAccount.PaymentAccount.Refundable)
}

func (s *PaymentTestSuite) updateParams(params paymenttypes.Params) {
	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()

	ts := time.Now().Unix()
	queryParamsRequest := &paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, queryParamsRequest)
	s.Require().NoError(err)

	msgUpdateParams := &paymenttypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    params,
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

	queryParamsByTimestampRequest := &paymenttypes.QueryParamsByTimestampRequest{Timestamp: ts}
	queryParamsByTimestampResponse, err := s.Client.PaymentQueryClient.ParamsByTimestamp(ctx, queryParamsByTimestampRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryParamsResponse.Params.VersionedParams.ReserveTime,
		queryParamsByTimestampResponse.Params.VersionedParams.ReserveTime)
	s.T().Logf("new params: %s", params.String())
}

func (s *PaymentTestSuite) createBucketAndObject() (keys.KeyManager, string, string, storagetypes.Uint, [][]byte) {
	var err error
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	// CreateBucket
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := "ch" + storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// CreateObject
	objectName := storagetestutils.GenRandomObjectName()
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
		storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType,
		storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)

	return user, bucketName, objectName, queryHeadObjectResponse.ObjectInfo.Id, expectChecksum
}

func (s *PaymentTestSuite) createBucket() (keys.KeyManager, string) {
	var err error
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	// CreateBucket
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := "ch" + storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	return user, bucketName
}

func (s *PaymentTestSuite) createObject(user keys.KeyManager, bucketName string) (keys.KeyManager, string, string, storagetypes.Uint, [][]byte) {
	var err error
	sp := s.StorageProviders[0]

	// CreateObject
	objectName := storagetestutils.GenRandomObjectName()
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
		storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType,
		storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)

	return user, bucketName, objectName, queryHeadObjectResponse.ObjectInfo.Id, expectChecksum
}

func (s *PaymentTestSuite) sealObject(bucketName, objectName string, objectId storagetypes.Uint, checksums [][]byte) {
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	s.T().Log("GVG info: ", gvg.String())

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, objectId, storagetypes.GenerateHash(checksums[:])).GetBlsSignHash()
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

	queryHeadObjectRequest2 := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse2, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest2)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse2.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse2.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse2.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
}

// TestVersionedParams_SealAfterReserveTimeChange will cover the following case:
// create an object, increase the reserve time, seal the object without error.
func (s *PaymentTestSuite) TestVersionedParams_SealObjectAfterReserveTimeChange() {
	ctx := context.Background()
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)

	// create bucket, create object
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject()

	// update params
	params := queryParamsResponse.GetParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", oldReserveTime, oldValidatorTaxRate)

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)

	s.updateParams(params)
	queryParamsResponse, err = s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params = queryParamsResponse.GetParams()
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", params.VersionedParams.ReserveTime, params.VersionedParams.ValidatorTaxRate)

	// seal object
	s.sealObject(bucketName, objectName, objectId, checksums)

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	// delete object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// delete bucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)

	// revert params
	params.VersionedParams.ReserveTime = oldReserveTime
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate
	s.updateParams(params)
}

// TestVersionedParams_DeleteAfterValidatorTaxRateChange will cover the following case:
// create a bucket with non-zero read quota, change the validator tax rate, delete the bucket.
// The rate of the validator tax address should be correct.
func (s *PaymentTestSuite) TestVersionedParams_DeleteBucketAfterValidatorTaxRateChange() {
	ctx := context.Background()
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)

	validatorTaxPoolRate := sdk.ZeroInt()
	queryStreamRequest := paymenttypes.QueryGetStreamRecordRequest{Account: paymenttypes.ValidatorTaxPoolAddress.String()}
	queryStreamResponse, err := s.Client.PaymentQueryClient.StreamRecord(ctx, &queryStreamRequest)
	if err != nil {
		s.Require().ErrorContains(err, "key not found")
	} else {
		s.Require().NoError(err)
		validatorTaxPoolRate = queryStreamResponse.StreamRecord.NetflowRate
	}
	s.T().Logf("netflow, validatorTaxPoolRate: %s", validatorTaxPoolRate)

	// create bucket, create object
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject()

	// seal object
	s.sealObject(bucketName, objectName, objectId, checksums)

	// update params
	params := queryParamsResponse.GetParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", oldReserveTime, oldValidatorTaxRate)

	params.VersionedParams.ReserveTime = oldReserveTime / 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)

	s.updateParams(params)
	queryParamsResponse, err = s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params = queryParamsResponse.GetParams()
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", params.VersionedParams.ReserveTime, params.VersionedParams.ValidatorTaxRate)

	// delete object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// delete bucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)

	queryStreamResponse, err = s.Client.PaymentQueryClient.StreamRecord(ctx, &queryStreamRequest)
	s.Require().NoError(err)
	s.Require().Equal(validatorTaxPoolRate, queryStreamResponse.StreamRecord.NetflowRate)

	// revert params
	params.VersionedParams.ReserveTime = oldReserveTime
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate
	s.updateParams(params)
}

// TestVersionedParams_DeleteObjectAfterReserveTimeChange will cover the following case:
// create an object, change the reserve time, the object can be force deleted even the object's own has no enough balance.
func (s *PaymentTestSuite) TestVersionedParams_DeleteObjectAfterReserveTimeChange() {
	ctx := context.Background()
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)

	// create bucket, create object
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject()

	// seal object
	s.sealObject(bucketName, objectName, objectId, checksums)

	// for payment
	time.Sleep(2 * time.Second)

	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.SubRaw(5*types.DecimalGwei)),
	))

	simulateResponse := s.SimulateTx(msgSend, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	msgSend.Amount = sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.Sub(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))),
	)
	s.SendTxBlock(user, msgSend)
	queryBalanceResponse, err = s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)
	s.Require().Equal(int64(0), queryBalanceResponse.Balance.Amount.Int64())

	// update params
	params := queryParamsResponse.GetParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", oldReserveTime, oldValidatorTaxRate)

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)

	s.updateParams(params)
	queryParamsResponse, err = s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params = queryParamsResponse.GetParams()
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", params.VersionedParams.ReserveTime, params.VersionedParams.ValidatorTaxRate)

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	sp := s.StorageProviders[0]

	// force delete bucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueBucket)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime > deleteAt {
			break
		}
	}

	_, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().ErrorContains(err, "No such object")

	// revert params
	params.VersionedParams.ReserveTime = oldReserveTime
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate
	s.updateParams(params)
}

func (s *PaymentTestSuite) TestDepositAndResume_InOneBlock() {
	ctx := context.Background()
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	params, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("params %s, err: %v", params, err)
	s.Require().NoError(err)
	reserveTime := params.Params.VersionedParams.ReserveTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	totalUserRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	taxRateParam := params.Params.VersionedParams.ValidatorTaxRate
	taxStreamRate := taxRateParam.MulInt(totalUserRate).TruncateInt()
	expectedRate := totalUserRate.Add(taxStreamRate)
	paymentAccountBNBNeeded := expectedRate.Mul(sdkmath.NewIntFromUint64(reserveTime))

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: userAddr,
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	paymentAccountsReq := &paymenttypes.QueryGetPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.GetPaymentAccountsByOwner(ctx, paymentAccountsReq)
	s.Require().NoError(err)
	s.T().Logf("paymentAccounts %s", core.YamlString(paymentAccounts))
	paymentAddr := paymentAccounts.PaymentAccounts[0]
	s.Require().Lenf(paymentAccounts.PaymentAccounts, 1, "paymentAccounts %s", core.YamlString(paymentAccounts))

	// deposit BNB needed
	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAddr,
		Amount:  paymentAccountBNBNeeded,
	}
	_ = s.SendTxBlock(user, msgDeposit)

	// create bucket
	bucketName := storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil, bucketChargedReadQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// check payment account stream record
	paymentAccountStreamRecord := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), sdkmath.ZeroInt().String())

	// wait until settle time
	retryCount := 0
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)
		currentTimestamp := latestBlock.Block.Time.Unix()
		s.T().Logf("currentTimestamp %d, paymentAccountStreamRecord.SettleTimestamp %d", currentTimestamp, paymentAccountStreamRecord.SettleTimestamp)
		if currentTimestamp > paymentAccountStreamRecord.SettleTimestamp {
			break
		}
		time.Sleep(time.Second)
		retryCount++
		if retryCount > 60 {
			s.T().Fatalf("wait for settle time timeout")
		}
	}
	// check auto settle
	paymentStreamRecordAfterAutoSettle := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentStreamRecordAfterAutoSettle %s", core.YamlString(paymentStreamRecordAfterAutoSettle))
	s.Require().NotEqual(paymentStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	// deposit, balance not enough to resume
	depositAmount1 := sdk.NewInt(1)
	msgDeposit1 := &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  depositAmount1,
	}
	_ = s.SendTxBlock(user, msgDeposit1)

	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit1 := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit1 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit1))
	s.Require().NotEqual(paymentAccountStreamRecordAfterDeposit1.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	// deposit and resume
	depositAmount2 := sdk.NewInt(1e10)
	msgDeposit2 := &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  depositAmount2,
	}
	s.SendTxBlock(user, msgDeposit2)
	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit2 := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit2 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit2))
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.StaticBalance.Add(paymentAccountStreamRecordAfterDeposit2.BufferBalance).String(), paymentAccountStreamRecordAfterDeposit1.StaticBalance.Add(depositAmount2).String())
}

func (s *PaymentTestSuite) TestDepositAndResume_InBlocks() {
	ctx := context.Background()
	// update params
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params := queryParamsResponse.GetParams()
	oldMaxAutoResumeFlowCount := params.MaxAutoResumeFlowCount
	s.T().Logf("params, MaxAutoResumeFlowCount: %d", oldMaxAutoResumeFlowCount)

	params.MaxAutoResumeFlowCount = 1 // update to 1
	s.updateParams(params)
	queryParamsResponse, err = s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params = queryParamsResponse.GetParams()
	s.T().Logf("params: %s", params.String())

	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()

	reserveTime := params.VersionedParams.ReserveTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	totalUserRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	taxRateParam := params.VersionedParams.ValidatorTaxRate
	taxStreamRate := taxRateParam.MulInt(totalUserRate).TruncateInt()
	expectedRate := totalUserRate.Add(taxStreamRate)
	paymentAccountBNBNeeded := expectedRate.Mul(sdkmath.NewIntFromUint64(reserveTime))

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: userAddr,
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	paymentAccountsReq := &paymenttypes.QueryGetPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.GetPaymentAccountsByOwner(ctx, paymentAccountsReq)
	s.Require().NoError(err)
	s.T().Logf("paymentAccounts %s", core.YamlString(paymentAccounts))
	paymentAddr := paymentAccounts.PaymentAccounts[0]
	s.Require().Lenf(paymentAccounts.PaymentAccounts, 1, "paymentAccounts %s", core.YamlString(paymentAccounts))

	// deposit BNB needed
	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAddr,
		Amount:  paymentAccountBNBNeeded,
	}
	_ = s.SendTxBlock(user, msgDeposit)

	// create bucket
	bucketName := storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil, bucketChargedReadQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// check payment account stream record
	paymentAccountStreamRecord := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), sdkmath.ZeroInt().String())

	// wait until settle time
	retryCount := 0
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)
		currentTimestamp := latestBlock.Block.Time.Unix()
		s.T().Logf("currentTimestamp %d, paymentAccountStreamRecord.SettleTimestamp %d", currentTimestamp, paymentAccountStreamRecord.SettleTimestamp)
		if currentTimestamp > paymentAccountStreamRecord.SettleTimestamp {
			break
		}
		time.Sleep(time.Second)
		retryCount++
		if retryCount > 60 {
			s.T().Fatalf("wait for settle time timeout")
		}
	}
	// check auto settle
	paymentStreamRecordAfterAutoSettle := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentStreamRecordAfterAutoSettle %s", core.YamlString(paymentStreamRecordAfterAutoSettle))
	s.Require().NotEqual(paymentStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	// deposit and resume
	depositAmount := sdk.NewInt(1e10)
	msgDeposit = &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  depositAmount,
	}
	mode := tx.BroadcastMode_BROADCAST_MODE_ASYNC
	txOpt := types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.SendTxWithTxOpt(msgDeposit, user, txOpt)

	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit %s", core.YamlString(paymentAccountStreamRecordAfterDeposit))
	s.Require().NotEqual(paymentAccountStreamRecordAfterDeposit.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	// wait blocks
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)

		paymentAccountStreamRecordAfterDeposit = s.GetStreamRecord(paymentAddr)
		s.T().Logf("paymentAccountStreamRecordAfterDeposit %s at %d", core.YamlString(paymentAccountStreamRecordAfterDeposit), latestBlock.Block.Height)
		if paymentAccountStreamRecordAfterDeposit.Status == paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE {
			break
		}
		time.Sleep(time.Second)
		retryCount++
		if retryCount > 60 {
			s.T().Fatalf("wait for resume time timeout")
		}
	}

	// revert params
	params.MaxAutoResumeFlowCount = oldMaxAutoResumeFlowCount
	s.updateParams(params)
}

func (s *PaymentTestSuite) TestAutoSettle_InOneBlock() {
	ctx := context.Background()
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{
		StorageProviderId: sp.Info.Id,
		FamilyId:          gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily

	bucketChargedReadQuota := uint64(1000)
	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)
	reserveTime := paymentParams.Params.VersionedParams.ReserveTime
	forcedSettleTime := paymentParams.Params.ForcedSettleTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	totalUserRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	taxRateParam := paymentParams.Params.VersionedParams.ValidatorTaxRate
	taxStreamRate := taxRateParam.MulInt(totalUserRate).TruncateInt()
	expectedRate := totalUserRate.Add(taxStreamRate)
	paymentAccountBNBNeeded := expectedRate.Mul(sdkmath.NewIntFromUint64(reserveTime))

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: userAddr,
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
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
	_ = s.SendTxBlock(user, msgDeposit)

	// create bucket from payment account
	bucketName := storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil, bucketChargedReadQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)
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
	msgCreateBucket.BucketName = storagetestutils.GenRandomBucketName()
	msgCreateBucket.PaymentAddress = ""
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// check user stream record
	userStreamRecord := s.GetStreamRecord(userAddr)
	s.T().Logf("userStreamRecord %s", core.YamlString(userStreamRecord))
	s.Require().Equal(userStreamRecord.SettleTimestamp, userStreamRecord.CrudTimestamp+int64(reserveTime-forcedSettleTime))
	familyStreamRecord := s.GetStreamRecord(family.VirtualPaymentAddress)
	s.T().Logf("familyStreamRecord %s", core.YamlString(familyStreamRecord))
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
	familyStreamRecordAfterAutoSettle := s.GetStreamRecord(family.VirtualPaymentAddress)
	s.T().Logf("familyStreamRecordAfterAutoSettle %s", core.YamlString(familyStreamRecordAfterAutoSettle))
	paymentAccountStreamRecordAfterAutoSettle := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterAutoSettle %s", core.YamlString(paymentAccountStreamRecordAfterAutoSettle))
	// payment account become frozen
	s.Require().NotEqual(paymentAccountStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(familyStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
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
	_ = s.SendTxBlock(user, msgDeposit1)
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
	s.SendTxBlock(user, msgDeposit2)
	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit2 := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit2 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit2))
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.StaticBalance.Add(paymentAccountStreamRecordAfterDeposit2.BufferBalance).String(), paymentAccountStreamRecordAfterDeposit1.StaticBalance.Add(depositAmount2).String())
}

func (s *PaymentTestSuite) TestAutoSettle_InBlocks() {
	ctx := context.Background()
	// update params
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params := queryParamsResponse.GetParams()
	oldMaxAutoSettleFlowCount := params.MaxAutoSettleFlowCount
	s.T().Logf("params, MaxAutoSettleFlowCount: %d", oldMaxAutoSettleFlowCount)

	params.MaxAutoSettleFlowCount = 2 // update to 2
	s.updateParams(params)
	queryParamsResponse, err = s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)
	params = queryParamsResponse.GetParams()
	s.T().Logf("params: %s", params.String())

	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()

	reserveTime := params.VersionedParams.ReserveTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	totalUserRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	taxRateParam := params.VersionedParams.ValidatorTaxRate
	taxStreamRate := taxRateParam.MulInt(totalUserRate).TruncateInt()
	expectedRate := totalUserRate.Add(taxStreamRate)
	paymentAccountBNBNeeded := expectedRate.Mul(sdkmath.NewIntFromUint64(reserveTime))

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: userAddr,
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	paymentAccountsReq := &paymenttypes.QueryGetPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.GetPaymentAccountsByOwner(ctx, paymentAccountsReq)
	s.Require().NoError(err)
	s.T().Logf("paymentAccounts %s", core.YamlString(paymentAccounts))
	paymentAddr := paymentAccounts.PaymentAccounts[0]
	s.Require().Lenf(paymentAccounts.PaymentAccounts, 1, "paymentAccounts %s", core.YamlString(paymentAccounts))

	// deposit BNB needed
	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAddr,
		Amount:  paymentAccountBNBNeeded,
	}
	_ = s.SendTxBlock(user, msgDeposit)

	// create bucket
	bucketName := storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil, bucketChargedReadQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// check payment account stream record
	paymentAccountStreamRecord := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), sdkmath.ZeroInt().String())

	// wait until settle time
	retryCount := 0
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)
		currentTimestamp := latestBlock.Block.Time.Unix()
		s.T().Logf("currentTimestamp %d, paymentAccountStreamRecord.SettleTimestamp %d", currentTimestamp, paymentAccountStreamRecord.SettleTimestamp)
		if currentTimestamp > paymentAccountStreamRecord.SettleTimestamp {
			break
		}
		time.Sleep(time.Second)
		retryCount++
		if retryCount > 60 {
			s.T().Fatalf("wait for settle time timeout")
		}
	}
	// check auto settle
	for {
		paymentStreamRecordAfterAutoSettle := s.GetStreamRecord(paymentAddr)
		s.T().Logf("paymentStreamRecordAfterAutoSettle %s", core.YamlString(paymentStreamRecordAfterAutoSettle))
		if paymentStreamRecordAfterAutoSettle.NetflowRate.IsZero() {
			break
		}
		time.Sleep(500 * time.Millisecond)
		retryCount++
		if retryCount > 60 {
			s.T().Fatalf("wait for settle time timeout")
		}
	}
	paymentStreamRecordAfterAutoSettle := s.GetStreamRecord(paymentAddr)
	s.T().Logf("paymentStreamRecordAfterAutoSettle %s", core.YamlString(paymentStreamRecordAfterAutoSettle))
	s.Require().NotEqual(paymentStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	// revert params
	params.MaxAutoSettleFlowCount = oldMaxAutoSettleFlowCount
	s.updateParams(params)
}

func (s *PaymentTestSuite) TestDeleteBucketWithReadQuota() {
	var err error
	ctx := context.Background()
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{
		StorageProviderId: sp.Info.Id,
		FamilyId:          gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}

	// CreateBucket
	chargedReadQuota := uint64(100)
	bucketName := storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PUBLIC_READ, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, chargedReadQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	streamRecordsBeforeDelete := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeDelete: %s", core.YamlString(streamRecordsBeforeDelete))
	s.Require().NotEqual(streamRecordsBeforeDelete.User.NetflowRate.String(), "0")

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)

	// check the billing change
	streamRecordsAfterDelete := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeDelete: %s", core.YamlString(streamRecordsAfterDelete))
	s.Require().Equal(streamRecordsAfterDelete.User.NetflowRate.String(), "0")
}

func (s *PaymentTestSuite) TestStorageSmoke() {
	var err error
	ctx := context.Background()
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{
		StorageProviderId: sp.Info.Id,
		FamilyId:          gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBeforeCreateBucket := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeCreateBucket: %s", core.YamlString(streamRecordsBeforeCreateBucket))

	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	// create bucket
	bucketName := storagetestutils.GenRandomBucketName()
	bucketChargedReadQuota := uint64(1000)
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.ChargedReadQuota = bucketChargedReadQuota
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// check bill after creating bucket
	userBankAccount, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: user.GetAddr().String(),
		Denom:   s.Config.Denom,
	})
	s.Require().NoError(err)
	s.T().Logf("user bank account %s", userBankAccount)

	streamRecordsAfterCreateBucket := s.GetStreamRecords(streamAddresses)
	userStreamRecord := streamRecordsAfterCreateBucket.User
	s.Require().Equal(userStreamRecord.StaticBalance, sdkmath.ZeroInt())

	// check price and rate calculation
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: queryHeadBucketResponse.BucketInfo.CreateAt,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	readChargeRate := readPrice.MulInt(sdk.NewIntFromUint64(queryHeadBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	userTaxRate := paymentParams.Params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate := readChargeRate.Add(userTaxRate)
	s.Require().Equal(userStreamRecord.NetflowRate.Abs(), userTotalRate)
	expectedOutFlows := []paymenttypes.OutFlow{
		{ToAddress: family.VirtualPaymentAddress, Rate: readChargeRate},
		{ToAddress: paymenttypes.ValidatorTaxPoolAddress.String(), Rate: userTaxRate},
	}
	userOutFlowsResponse, err := s.Client.OutFlows(ctx, &paymenttypes.QueryOutFlowsRequest{Account: user.GetAddr().String()})
	s.Require().NoError(err)
	sort.Slice(userOutFlowsResponse.OutFlows, func(i, j int) bool {
		return userOutFlowsResponse.OutFlows[i].ToAddress < userOutFlowsResponse.OutFlows[j].ToAddress
	})
	sort.Slice(expectedOutFlows, func(i, j int) bool {
		return expectedOutFlows[i].ToAddress < expectedOutFlows[j].ToAddress
	})
	s.Require().Equal(expectedOutFlows, userOutFlowsResponse.OutFlows)

	// CreateObject
	objectName := storagetestutils.GenRandomObjectName()
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
	// simulate
	res := s.SimulateTx(msgCreateObject, user)
	s.T().Logf("res %v", res.Result)
	// check EventFeePreview in simulation result
	var feePreviewEventEmitted bool
	events := res.Result.Events
	for _, event := range events {
		if event.Type == "greenfield.payment.EventFeePreview" {
			s.T().Logf("event %v", event)
			feePreviewEventEmitted = true
		}
	}
	s.Require().True(feePreviewEventEmitted)
	s.SendTxBlock(user, msgCreateObject)

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

	queryGetSecondarySpStorePriceByTime, err := s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: queryHeadBucketResponse.BucketInfo.CreateAt,
	})
	s.T().Logf("queryGetSecondarySpStorePriceByTime %s, err: %v", queryGetSecondarySpStorePriceByTime, err)
	s.Require().NoError(err)
	primaryStorePrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice
	secondaryStorePrice := queryGetSecondarySpStorePriceByTime.SecondarySpStorePrice.StorePrice
	chargeSize := s.GetChargeSize(queryHeadObjectResponse.ObjectInfo.PayloadSize)
	expectedChargeRate := primaryStorePrice.Add(secondaryStorePrice.MulInt64(6)).MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt()
	expectedLockedBalance := expectedChargeRate.Mul(sdkmath.NewIntFromUint64(paymentParams.Params.VersionedParams.ReserveTime))

	streamRecordsAfterCreateObject := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsAfterCreateObject %s", core.YamlString(streamRecordsAfterCreateObject))
	userStreamAccountAfterCreateObj := streamRecordsAfterCreateObject.User

	s.Require().Equal(expectedLockedBalance.String(), userStreamAccountAfterCreateObj.LockBalance.String())

	// seal object
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(expectChecksum[:])).GetBlsSignHash()
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

	// check bill after seal
	streamRecordsAfterSeal := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsAfterSeal %s", core.YamlString(streamRecordsAfterSeal))
	s.Require().Equal(sdkmath.ZeroInt(), streamRecordsAfterSeal.User.LockBalance)
	s.CheckStreamRecordsBeforeAndAfter(streamRecordsAfterCreateObject, streamRecordsAfterSeal, readPrice, readChargeRate, primaryStorePrice, secondaryStorePrice, chargeSize, uint64(payloadSize))

	// query dynamic balance
	time.Sleep(3 * time.Second)
	queryDynamicBalanceRequest := paymenttypes.QueryDynamicBalanceRequest{
		Account: user.GetAddr().String(),
	}
	queryDynamicBalanceResponse, err := s.Client.DynamicBalance(ctx, &queryDynamicBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("queryDynamicBalanceResponse %s", core.YamlString(queryDynamicBalanceResponse))

	// create empty object
	streamRecordsBeforeCreateEmptyObject := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeCreateEmptyObject %s", core.YamlString(streamRecordsBeforeCreateEmptyObject))

	emptyObjectName := "sub_directory/"
	// create empty test buffer
	var emptyBuffer bytes.Buffer
	emptyPayloadSize := emptyBuffer.Len()
	emptyChecksum := sdk.Keccak256(emptyBuffer.Bytes())
	emptyExpectChecksum := [][]byte{emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum, emptyChecksum}
	msgCreateEmptyObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, emptyObjectName, uint64(emptyPayloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, emptyExpectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateEmptyObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateEmptyObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateEmptyObject)

	streamRecordsAfterCreateEmptyObject := s.GetStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsAfterCreateEmptyObject %s", core.YamlString(streamRecordsAfterCreateEmptyObject))
	chargeSize = s.GetChargeSize(uint64(emptyPayloadSize))
	s.CheckStreamRecordsBeforeAndAfter(streamRecordsBeforeCreateEmptyObject, streamRecordsAfterCreateEmptyObject, readPrice, readChargeRate, primaryStorePrice, secondaryStorePrice, chargeSize, uint64(emptyPayloadSize))

	// test query auto settle records
	queryAllAutoSettleRecordRequest := paymenttypes.QueryAllAutoSettleRecordRequest{}
	queryAllAutoSettleRecordResponse, err := s.Client.AutoSettleRecordAll(ctx, &queryAllAutoSettleRecordRequest)
	s.Require().NoError(err)
	s.T().Logf("queryAllAutoSettleRecordResponse %s", core.YamlString(queryAllAutoSettleRecordResponse))
	s.Require().True(len(queryAllAutoSettleRecordResponse.AutoSettleRecord) >= 1)

	// simulate delete object, check fee preview
	deleteObjectMsg := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	deleteObjectSimRes := s.SimulateTx(deleteObjectMsg, user)
	s.T().Logf("deleteObjectSimRes %v", deleteObjectSimRes.Result)
}

// TestForceDeletion_DeleteAfterPriceChange will cover the following case:
// create an object, sp increase the price a lot, the object can be force deleted even the object's own has no enough balance.
func (s *PaymentTestSuite) TestForceDeletion_AfterPriceChange() {
	ctx := context.Background()

	// set storage price
	sp := s.StorageProviders[0]
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// create bucket
	user, bucketName := s.createBucket()

	// create & seal objects
	_, _, objectName1, objectId1, checksums1 := s.createObject(user, bucketName)
	s.sealObject(bucketName, objectName1, objectId1, checksums1)

	_, _, objectName2, objectId2, checksums2 := s.createObject(user, bucketName)
	s.sealObject(bucketName, objectName2, objectId2, checksums2)

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// for payment
	time.Sleep(2 * time.Second)

	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.SubRaw(5*types.DecimalGwei)),
	))

	simulateResponse := s.SimulateTx(msgSend, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	msgSend.Amount = sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.Sub(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))),
	)
	s.SendTxBlock(user, msgSend)
	queryBalanceResponse, err = s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)
	s.Require().Equal(int64(0), queryBalanceResponse.Balance.Amount.Int64())

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName1,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	// force delete bucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueBucket)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime > deleteAt {
			break
		}
	}

	_, err = s.Client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucketName})
	s.Require().ErrorContains(err, "No such bucket")

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) GetStreamRecord(addr string) (sr paymenttypes.StreamRecord) {
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

func (s *PaymentTestSuite) GetStreamRecords(addrs []string) (streamRecords StreamRecords) {
	streamRecords.User = s.GetStreamRecord(addrs[0])
	streamRecords.GVGFamily = s.GetStreamRecord(addrs[1])
	streamRecords.GVG = s.GetStreamRecord(addrs[2])
	streamRecords.Tax = s.GetStreamRecord(addrs[3])
	return
}

func (s *PaymentTestSuite) CheckStreamRecordsBeforeAndAfter(streamRecordsBefore StreamRecords, streamRecordsAfter StreamRecords, readPrice sdk.Dec,
	readChargeRate sdkmath.Int, primaryStorePrice sdk.Dec, secondaryStorePrice sdk.Dec, chargeSize uint64, payloadSize uint64) {
	userRateDiff := streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate)
	gvgFamilyRateDiff := streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate)
	gvgRateDiff := streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate)
	taxRateDiff := streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate)
	s.Require().Equal(userRateDiff, gvgFamilyRateDiff.Add(gvgRateDiff).Add(taxRateDiff).Neg())

	outFlowsResponse, err := s.Client.OutFlows(context.Background(), &paymenttypes.QueryOutFlowsRequest{Account: streamRecordsAfter.User.Account})
	s.Require().NoError(err)
	userOutflowMap := lo.Reduce(outFlowsResponse.OutFlows, func(m map[string]sdkmath.Int, outflow paymenttypes.OutFlow, i int) map[string]sdkmath.Int {
		m[outflow.ToAddress] = outflow.Rate
		return m
	}, make(map[string]sdkmath.Int))
	if payloadSize != 0 {
		gvgFamilyRate := primaryStorePrice.MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt().Add(readChargeRate)
		s.Require().Equal(gvgFamilyRate, userOutflowMap[streamRecordsAfter.GVGFamily.Account])

		gvgRate := secondaryStorePrice.MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt().MulRaw(6)
		s.Require().Equal(gvgRate, userOutflowMap[streamRecordsAfter.GVG.Account])
	}
}

func (s *PaymentTestSuite) GetChargeSize(payloadSize uint64) uint64 {
	ctx := context.Background()
	storageParams, err := s.Client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	s.Require().NoError(err)
	s.T().Logf("storageParams %s", storageParams)
	minChargeSize := storageParams.Params.VersionedParams.MinChargeSize
	if payloadSize < minChargeSize {
		return minChargeSize
	} else {
		return payloadSize
	}
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}
