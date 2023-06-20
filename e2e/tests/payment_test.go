package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	storagetestutil "github.com/bnb-chain/greenfield/testutil/storage"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

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
}

func (s *PaymentTestSuite) createObject() (keys.KeyManager, string, string, storagetypes.Uint, [][]byte) {
	var err error
	sp := s.StorageProviders[0]
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	// CreateBucket
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := "ch" + storagetestutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	// CreateObject
	objectName := storagetestutil.GenRandomObjectName()
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

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	signBz := storagetypes.NewSecondarySpSealObjectSignDoc(objectId, gvgId, storagetypes.GenerateHash(checksums[:])).GetSignBytes()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := blsSignAndVerify(s.StorageProviders[i], signBz)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
	}
	aggBlsSig, err := blsAggregateAndVerify(secondarySPBlsPubKeys, signBz, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.T().Logf("msg %s", msgSealObject.String())
	s.SendTxBlock(sp.SealKey, msgSealObject)

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
}

// TestVersionedParams_SealAfterReserveTimeChange will cover the following case:
// create an object, increase the reserve time, seal the object without error.
func (s *PaymentTestSuite) TestVersionedParams_SealObjectAfterReserveTimeChange() {
	ctx := context.Background()
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, &queryParamsRequest)
	s.Require().NoError(err)

	// create bucket, create object
	user, bucketName, objectName, objectId, checksums := s.createObject()

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
	user, bucketName, objectName, objectId, checksums := s.createObject()

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
	user, bucketName, objectName, objectId, checksums := s.createObject()

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

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}
