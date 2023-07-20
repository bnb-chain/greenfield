package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
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
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
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
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type StreamRecords struct {
	User      paymenttypes.StreamRecord
	GVGFamily paymenttypes.StreamRecord
	GVG       paymenttypes.StreamRecord
	Tax       paymenttypes.StreamRecord
}

type PaymentTestSuite struct {
	core.BaseSuite
	defaultParams paymenttypes.Params
}

func (s *PaymentTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.defaultParams = s.queryParams()

}

func (s *PaymentTestSuite) SetupTest() {
	s.RefreshGVGFamilies()
}

func (s *PaymentTestSuite) TestCreatePaymentAccount() {
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

// TestVersionedParams_SealAfterReserveTimeChange will cover the following case:
// create an object, increase the reserve time, seal the object without error.
func (s *PaymentTestSuite) TestVersionedParams_SealObjectAfterReserveTimeChange() {
	defer s.revertParams()

	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	// create bucket, create object
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject(sp)

	// update params
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

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
}

// TestVersionedParams_DeleteAfterValidatorTaxRateChange will cover the following case:
// create a bucket with non-zero read quota, change the validator tax rate, delete the bucket.
// The rate of the validator tax address should be correct.
func (s *PaymentTestSuite) TestVersionedParams_DeleteBucketAfterValidatorTaxRateChange() {
	defer s.revertParams()

	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

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
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject(sp)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// update params
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldForceSettleTime := params.ForcedSettleTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate
	s.T().Logf("params, ReserveTime: %d, ValidatorTaxRate: %s", oldReserveTime, oldValidatorTaxRate)

	params.VersionedParams.ReserveTime = oldReserveTime / 2
	params.ForcedSettleTime = oldForceSettleTime / 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	// delete object
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// delete bucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)

	queryStreamResponse, err = s.Client.PaymentQueryClient.StreamRecord(ctx, &queryStreamRequest)
	s.Require().NoError(err)
	s.Require().Equal(validatorTaxPoolRate, queryStreamResponse.StreamRecord.NetflowRate)
}

// TestVersionedParams_DeleteObjectAfterReserveTimeChange will cover the following case:
// create an object, change the reserve time, the object can be force deleted even the object's own has no enough balance.
func (s *PaymentTestSuite) TestVersionedParams_DeleteObjectAfterReserveTimeChange() {
	defer s.revertParams()

	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	// create bucket, create object
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject(sp)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

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
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
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

	_, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().ErrorContains(err, "No such object")
}

func (s *PaymentTestSuite) TestDeposit_ActiveAccount() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	params := s.queryParams()
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
	s.T().Log("paymentAccountBNBNeeded", paymentAccountBNBNeeded.String())

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
		Amount:  paymentAccountBNBNeeded.MulRaw(2), // deposit more than needed
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
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), paymentAccountBNBNeeded.String())

	time.Sleep(5 * time.Second)

	// deposit
	msgDeposit = &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  paymentAccountBNBNeeded,
	}
	_ = s.SendTxBlock(user, msgDeposit)

	// check payment account stream record
	paymentAccountStreamRecordAfter := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfter %s", core.YamlString(paymentAccountStreamRecordAfter))
	s.Require().Equal(paymentAccountStreamRecordAfter.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	settledTime := paymentAccountStreamRecordAfter.CrudTimestamp - paymentAccountStreamRecord.CrudTimestamp
	settledBalance := expectedRate.MulRaw(settledTime)
	paymentBalanceChange := paymentAccountStreamRecordAfter.StaticBalance.Sub(paymentAccountStreamRecord.StaticBalance).
		Add(paymentAccountStreamRecordAfter.BufferBalance.Sub(paymentAccountStreamRecord.BufferBalance))
	s.Require().Equal(settledBalance.Add(paymentBalanceChange).Int64(), paymentAccountBNBNeeded.Int64())
	s.Require().Equal(paymentAccountBNBNeeded.MulRaw(3), settledBalance.Add(paymentAccountStreamRecordAfter.StaticBalance.Add(paymentAccountStreamRecordAfter.BufferBalance)))
}

func (s *PaymentTestSuite) TestDeposit_ResumeInOneBlock() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	params := s.queryParams()
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
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
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
	paymentStreamRecordAfterAutoSettle := s.getStreamRecord(paymentAddr)
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
	paymentAccountStreamRecordAfterDeposit1 := s.getStreamRecord(paymentAddr)
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
	paymentAccountStreamRecordAfterDeposit2 := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit2 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit2))
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.StaticBalance.Add(paymentAccountStreamRecordAfterDeposit2.BufferBalance).String(), paymentAccountStreamRecordAfterDeposit1.StaticBalance.Add(depositAmount2).String())
}

func (s *PaymentTestSuite) TestDeposit_ResumeInBlocks() {
	defer s.revertParams()

	ctx := context.Background()
	// update params
	params := s.queryParams()
	params.MaxAutoResumeFlowCount = 1 // update to 1
	s.updateParams(params)

	sp := s.PickStorageProvider()
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
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
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
	paymentStreamRecordAfterAutoSettle := s.getStreamRecord(paymentAddr)
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
	paymentAccountStreamRecordAfterDeposit := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit %s", core.YamlString(paymentAccountStreamRecordAfterDeposit))
	s.Require().NotEqual(paymentAccountStreamRecordAfterDeposit.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	// wait blocks
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)

		paymentAccountStreamRecordAfterDeposit = s.getStreamRecord(paymentAddr)
		s.T().Logf("paymentAccountStreamRecordAfterDeposit %s at %d", core.YamlString(paymentAccountStreamRecordAfterDeposit), latestBlock.Block.Height)
		if paymentAccountStreamRecordAfterDeposit.Status == paymenttypes.STREAM_ACCOUNT_STATUS_FROZEN &&
			!paymentAccountStreamRecordAfterDeposit.NetflowRate.IsZero() {
			s.T().Log("trying to deposit, which will error")
			msgDeposit = &paymenttypes.MsgDeposit{
				Creator: user.GetAddr().String(),
				To:      paymentAddr,
				Amount:  paymentAccountBNBNeeded,
			}
			s.SendTxBlockWithExpectErrorString(msgDeposit, user, "resuming")
		}
		if paymentAccountStreamRecordAfterDeposit.Status == paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE {
			break
		}
		time.Sleep(time.Second)
		retryCount++
		if retryCount > 60 {
			s.T().Fatalf("wait for resume time timeout")
		}
	}
}

func (s *PaymentTestSuite) TestAutoSettle_InOneBlock() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily

	bucketChargedReadQuota := uint64(1000)
	params := s.queryParams()
	reserveTime := params.VersionedParams.ReserveTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)
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
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), sdkmath.ZeroInt().String())
	govStreamRecord := s.getStreamRecord(paymenttypes.GovernanceAddress.String())
	s.T().Logf("govStreamRecord %s", core.YamlString(govStreamRecord))

	// increase bucket charged read quota is not allowed since the balance is not enough
	msgUpdateBucketInfo := &storagetypes.MsgUpdateBucketInfo{
		Operator:         user.GetAddr().String(),
		BucketName:       bucketName,
		ChargedReadQuota: &common.UInt64Value{Value: bucketChargedReadQuota + 1},
		Visibility:       storagetypes.VISIBILITY_TYPE_PUBLIC_READ,
	}
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "balance not enough, lack of")

	// wait until settle time
	retryCount := 0
	for {
		latestBlock, err := s.TmClient.TmClient.Block(ctx, nil)
		s.Require().NoError(err)
		currentTimestamp := latestBlock.Block.Time.Unix()
		s.T().Logf("currentTimestamp %d, userStreamRecord.SettleTimestamp %d", currentTimestamp, paymentAccountStreamRecord.SettleTimestamp)
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
	userStreamRecordAfterAutoSettle := s.getStreamRecord(userAddr)
	s.T().Logf("userStreamRecordAfterAutoSettle %s", core.YamlString(userStreamRecordAfterAutoSettle))
	familyStreamRecordAfterAutoSettle := s.getStreamRecord(family.VirtualPaymentAddress)
	s.T().Logf("familyStreamRecordAfterAutoSettle %s", core.YamlString(familyStreamRecordAfterAutoSettle))
	paymentAccountStreamRecordAfterAutoSettle := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterAutoSettle %s", core.YamlString(paymentAccountStreamRecordAfterAutoSettle))
	// payment account become frozen
	s.Require().NotEqual(paymentAccountStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(familyStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(userStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)

	govStreamRecordAfterSettle := s.getStreamRecord(paymenttypes.GovernanceAddress.String())
	s.T().Logf("govStreamRecordAfterSettle %s", core.YamlString(govStreamRecordAfterSettle))
	s.Require().NotEqual(govStreamRecordAfterSettle.StaticBalance.String(), govStreamRecord.StaticBalance.String())
	govStreamRecordStaticBalanceDelta := govStreamRecordAfterSettle.StaticBalance.Sub(govStreamRecord.StaticBalance)
	expectedGovBalanceDelta := paymentAccountStreamRecord.NetflowRate.Neg().MulRaw(paymentAccountStreamRecordAfterAutoSettle.CrudTimestamp - paymentAccountStreamRecord.CrudTimestamp)
	s.Require().True(govStreamRecordStaticBalanceDelta.Int64() >= expectedGovBalanceDelta.Int64())

	// deposit, balance not enough to resume
	depositAmount1 := sdk.NewInt(1)
	msgDeposit1 := &paymenttypes.MsgDeposit{
		Creator: userAddr,
		To:      paymentAddr,
		Amount:  depositAmount1,
	}
	_ = s.SendTxBlock(user, msgDeposit1)
	// check payment account stream record
	paymentAccountStreamRecordAfterDeposit1 := s.getStreamRecord(paymentAddr)
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
	paymentAccountStreamRecordAfterDeposit2 := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterDeposit2 %s", core.YamlString(paymentAccountStreamRecordAfterDeposit2))
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
	s.Require().Equal(paymentAccountStreamRecordAfterDeposit2.StaticBalance.Add(paymentAccountStreamRecordAfterDeposit2.BufferBalance).String(), paymentAccountStreamRecordAfterDeposit1.StaticBalance.Add(depositAmount2).String())
}

func (s *PaymentTestSuite) TestAutoSettle_InBlocks() {
	defer s.revertParams()

	ctx := context.Background()
	// update params
	params := s.queryParams()
	params.MaxAutoSettleFlowCount = 2 // update to 2
	s.updateParams(params)

	sp := s.PickStorageProvider()
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
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
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
		paymentStreamRecordAfterAutoSettle := s.getStreamRecord(paymentAddr)
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
	paymentStreamRecordAfterAutoSettle := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentStreamRecordAfterAutoSettle %s", core.YamlString(paymentStreamRecordAfterAutoSettle))
	s.Require().NotEqual(paymentStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_ACTIVE)
}

func (s *PaymentTestSuite) TestWithdraw() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	params := s.queryParams()
	reserveTime := params.VersionedParams.ReserveTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr: sp.OperatorKey.GetAddr().String(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(100000)
	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	totalUserRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	taxRateParam := params.VersionedParams.ValidatorTaxRate
	taxStreamRate := taxRateParam.MulInt(totalUserRate).TruncateInt()
	expectedRate := totalUserRate.Add(taxStreamRate)
	paymentAccountBNBNeeded := expectedRate.Mul(sdkmath.NewIntFromUint64(reserveTime))
	s.T().Log("paymentAccountBNBNeeded", paymentAccountBNBNeeded.String())

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
		Amount:  paymentAccountBNBNeeded.MulRaw(2), // deposit more than needed
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
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecord %s", core.YamlString(paymentAccountStreamRecord))
	s.Require().Equal(expectedRate.String(), paymentAccountStreamRecord.NetflowRate.Neg().String())
	s.Require().Equal(paymentAccountStreamRecord.BufferBalance.String(), paymentAccountBNBNeeded.String())
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.String(), paymentAccountBNBNeeded.String())

	time.Sleep(5 * time.Second)

	dynamicBalanceResp, err := s.Client.DynamicBalance(ctx, &paymenttypes.QueryDynamicBalanceRequest{Account: user.GetAddr().String()})
	s.Require().NoError(err)
	s.Require().True(dynamicBalanceResp.DynamicBalance.LT(paymentAccountBNBNeeded))

	// withdraw more than static balance
	withdrawMsg := paymenttypes.NewMsgWithdraw(userAddr, paymentAddr, paymentAccountBNBNeeded)
	s.SendTxBlockWithExpectErrorString(withdrawMsg, user, "not enough")

	// withdraw less than static balance
	amount := sdk.NewInt(1000)
	withdrawMsg = paymenttypes.NewMsgWithdraw(userAddr, paymentAddr, amount)
	s.SendTxBlock(user, withdrawMsg)
	paymentAccountStreamRecordAfter := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfter %s", core.YamlString(paymentAccountStreamRecordAfter))

	staticBalanceChange := paymentAccountStreamRecord.NetflowRate.MulRaw(paymentAccountStreamRecordAfter.CrudTimestamp - paymentAccountStreamRecord.CrudTimestamp).Neg()
	s.Require().Equal(paymentAccountStreamRecord.StaticBalance.Sub(paymentAccountStreamRecordAfter.StaticBalance).Int64(), amount.Add(staticBalanceChange).Int64())
}

func (s *PaymentTestSuite) TestWithdraw_WithoutEnoughBalance() {}

func (s *PaymentTestSuite) TestStorageBill_DeleteBucket_WithReadQuota() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	streamRecordsBeforeDelete := s.getStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeDelete: %s", core.YamlString(streamRecordsBeforeDelete))
	s.Require().NotEqual(streamRecordsBeforeDelete.User.NetflowRate.String(), "0")

	// DeleteBucket
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteBucket)

	// check the billing change
	streamRecordsAfterDelete := s.getStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeDelete: %s", core.YamlString(streamRecordsAfterDelete))
	s.Require().Equal(streamRecordsAfterDelete.User.NetflowRate.String(), "0")
}

func (s *PaymentTestSuite) TestStorageBill_Smoke() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBeforeCreateBucket := s.getStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsBeforeCreateBucket: %s", core.YamlString(streamRecordsBeforeCreateBucket))

	params := s.queryParams()

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

	streamRecordsAfterCreateBucket := s.getStreamRecords(streamAddresses)
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
	userTaxRate := params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
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
	chargeSize := s.getChargeSize(queryHeadObjectResponse.ObjectInfo.PayloadSize)
	expectedChargeRate := primaryStorePrice.Add(secondaryStorePrice.MulInt64(6)).MulInt(sdk.NewIntFromUint64(chargeSize)).TruncateInt()
	expectedChargeRate = params.VersionedParams.ValidatorTaxRate.MulInt(expectedChargeRate).TruncateInt().Add(expectedChargeRate)
	expectedLockedBalance := expectedChargeRate.Mul(sdkmath.NewIntFromUint64(params.VersionedParams.ReserveTime))

	streamRecordsAfterCreateObject := s.getStreamRecords(streamAddresses)
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

	// check bill after seal
	streamRecordsAfterSeal := s.getStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsAfterSeal %s", core.YamlString(streamRecordsAfterSeal))
	s.Require().Equal(sdkmath.ZeroInt(), streamRecordsAfterSeal.User.LockBalance)
	s.checkStreamRecordsBeforeAndAfter(streamRecordsAfterCreateObject, streamRecordsAfterSeal, readPrice, readChargeRate, primaryStorePrice, secondaryStorePrice, chargeSize, uint64(payloadSize))

	// query dynamic balance
	time.Sleep(3 * time.Second)
	queryDynamicBalanceRequest := paymenttypes.QueryDynamicBalanceRequest{
		Account: user.GetAddr().String(),
	}
	queryDynamicBalanceResponse, err := s.Client.DynamicBalance(ctx, &queryDynamicBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("queryDynamicBalanceResponse %s", core.YamlString(queryDynamicBalanceResponse))

	// create empty object
	streamRecordsBeforeCreateEmptyObject := s.getStreamRecords(streamAddresses)
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

	streamRecordsAfterCreateEmptyObject := s.getStreamRecords(streamAddresses)
	s.T().Logf("streamRecordsAfterCreateEmptyObject %s", core.YamlString(streamRecordsAfterCreateEmptyObject))
	chargeSize = s.getChargeSize(uint64(emptyPayloadSize))
	s.checkStreamRecordsBeforeAndAfter(streamRecordsBeforeCreateEmptyObject, streamRecordsAfterCreateEmptyObject, readPrice, readChargeRate, primaryStorePrice, secondaryStorePrice, chargeSize, uint64(emptyPayloadSize))

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

func (s *PaymentTestSuite) TestStorageBill_DeleteObjectBucket_WithoutPriceChange() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// create bucket
	bucketName := s.createBucket(sp, user, 256)

	//simulate delete bucket gas
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	simulateResponse := s.SimulateTx(msgDeleteBucket, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas := gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit)))
	s.T().Log("total gas", "gas", gas)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

	// for payment
	time.Sleep(2 * time.Second)

	//transfer gas
	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdkmath.NewInt(5*types.DecimalGwei)),
	))
	simulateResponse = s.SimulateTx(msgSend, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	//delete object gas
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName1)
	simulateResponse = s.SimulateTx(msgDeleteObject, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	// transfer out user's balance
	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	msgSend.Amount = sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.Sub(gas)),
	)
	s.SendTxBlock(user, msgSend)
	_, err = s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	s.SendTxBlock(user, msgDeleteObject)
	s.SendTxBlock(user, msgDeleteBucket)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

func (s *PaymentTestSuite) TestStorageBill_DeleteObjectBucket_WithPriceChange() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// create bucket
	bucketName := s.createBucket(sp, user, 256)

	//simulate delete bucket gas
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	simulateResponse := s.SimulateTx(msgDeleteBucket, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas := gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit)))
	s.T().Log("total gas", "gas", gas)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

	// for payment
	time.Sleep(2 * time.Second)

	//transfer gas
	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdkmath.NewInt(5*types.DecimalGwei)),
	))
	simulateResponse = s.SimulateTx(msgSend, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	//delete object gas
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName1)
	simulateResponse = s.SimulateTx(msgDeleteObject, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	// transfer out user's balance
	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	msgSend.Amount = sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.Sub(gas)),
	)
	s.SendTxBlock(user, msgSend)
	_, err = s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	// sp price changes
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(1000),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	s.SendTxBlock(user, msgDeleteObject)
	s.SendTxBlock(user, msgDeleteBucket)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) TestStorageBill_DeleteObjectBucket_WithPriceChangeReserveTimeChange() {
	defer s.revertParams()

	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// create bucket
	bucketName := s.createBucket(sp, user, 256)

	//simulate delete bucket gas
	msgDeleteBucket := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	simulateResponse := s.SimulateTx(msgDeleteBucket, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas := gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit)))
	s.T().Log("total gas", "gas", gas)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

	// for payment
	time.Sleep(2 * time.Second)

	//transfer gas
	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdkmath.NewInt(5*types.DecimalGwei)),
	))
	simulateResponse = s.SimulateTx(msgSend, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	//delete object gas
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName1)
	simulateResponse = s.SimulateTx(msgDeleteObject, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	// transfer out user's balance
	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	msgSend.Amount = sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.Sub(gas)),
	)
	s.SendTxBlock(user, msgSend)
	_, err = s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)

	// sp price changes
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(1000),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// update params
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	s.SendTxBlock(user, msgDeleteObject)
	s.SendTxBlock(user, msgDeleteBucket)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) TestStorageBill_DeleteObject_WithStoreLessThanReserveTime() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	params := s.queryParams()
	reserveTime := params.VersionedParams.ReserveTime

	// create bucket
	bucketName := s.createBucket(sp, user, 256)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, payloadSize := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

	headObjectRes, err := s.Client.HeadObject(ctx, &storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName1,
	})
	s.Require().NoError(err)
	s.T().Log("headObjectRes", headObjectRes)

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, userRateRead := s.calculateReadRates(sp, bucketName)
	_, _, _, userRateStore := s.calculateStorageRates(sp, bucketName, objectName1, payloadSize, 0)

	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName1)
	s.SendTxBlock(user, msgDeleteObject)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)

	settledTime := streamRecordsAfter.User.CrudTimestamp - streamRecordsBefore.User.CrudTimestamp
	timeToPay := int64(reserveTime) + headObjectRes.ObjectInfo.CreateAt - streamRecordsAfter.User.CrudTimestamp
	balanceDelta := userRateRead.Add(userRateStore).MulRaw(settledTime).Add(userRateStore.MulRaw(timeToPay))

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), userRateStore.Int64())
	userBalanceChange := streamRecordsAfter.User.BufferBalance.Add(streamRecordsAfter.User.StaticBalance).
		Sub(streamRecordsBefore.User.BufferBalance.Add(streamRecordsBefore.User.StaticBalance))
	s.Require().Equal(userBalanceChange.Neg().Int64(), balanceDelta.Int64())

	familyDelta := streamRecordsAfter.GVGFamily.StaticBalance.Sub(streamRecordsBefore.GVGFamily.StaticBalance)
	gvgDelta := streamRecordsAfter.GVG.StaticBalance.Sub(streamRecordsBefore.GVG.StaticBalance)
	taxPoolDelta := streamRecordsAfter.Tax.StaticBalance.Sub(streamRecordsBefore.Tax.StaticBalance)
	s.T().Log("familyDelta", familyDelta, "gvgDelta", gvgDelta, "taxPoolDelta", taxPoolDelta)
	s.Require().True(familyDelta.Add(gvgDelta).Add(taxPoolDelta).Int64() >= balanceDelta.Int64()) // could exist other buckets/objects on the gvg & family
}

func (s *PaymentTestSuite) TestStorageBill_DeleteObject_WithStoreMoreThanReserveTime() {
	defer s.revertParams()

	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	params := s.queryParams()
	params.VersionedParams.ReserveTime = 5
	params.ForcedSettleTime = 2
	s.updateParams(params)

	// create bucket
	bucketName := s.createBucket(sp, user, 256)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, payloadSize := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

	headObjectRes, err := s.Client.HeadObject(ctx, &storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName1,
	})
	s.Require().NoError(err)
	s.T().Log("headObjectRes", headObjectRes)

	time.Sleep(5 * time.Second)
	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)
	balanceBefore := queryBalanceResponse.Balance.Amount

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, userRateRead := s.calculateReadRates(sp, bucketName)
	_, _, _, userRateStore := s.calculateStorageRates(sp, bucketName, objectName1, payloadSize, 0)

	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName1)
	simulateResponse := s.SimulateTx(msgDeleteObject, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)
	gas := gasPrice.Amount.MulRaw(int64(gasLimit))

	// delete object
	s.SendTxBlock(user, msgDeleteObject)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	queryBalanceRequest = banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err = s.Client.BankQueryClient.Balance(ctx, &queryBalanceRequest)
	s.Require().NoError(err)
	balanceAfter := queryBalanceResponse.Balance.Amount

	settledTime := streamRecordsAfter.User.CrudTimestamp - streamRecordsBefore.User.CrudTimestamp
	balanceDelta := userRateRead.Add(userRateStore).MulRaw(settledTime)

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), userRateStore.Int64())
	userBalanceChange := streamRecordsBefore.User.BufferBalance.Add(streamRecordsBefore.User.StaticBalance).
		Sub(streamRecordsAfter.User.BufferBalance.Add(streamRecordsAfter.User.StaticBalance))
	userBalanceChange = userBalanceChange.Add(balanceBefore.Sub(balanceAfter)).Sub(gas)
	s.Require().Equal(userBalanceChange.Int64(), balanceDelta.Int64())

	familyDelta := streamRecordsAfter.GVGFamily.StaticBalance.Sub(streamRecordsBefore.GVGFamily.StaticBalance)
	gvgDelta := streamRecordsAfter.GVG.StaticBalance.Sub(streamRecordsBefore.GVG.StaticBalance)
	taxPoolDelta := streamRecordsAfter.Tax.StaticBalance.Sub(streamRecordsBefore.Tax.StaticBalance)
	s.T().Log("familyDelta", familyDelta, "gvgDelta", gvgDelta, "taxPoolDelta", taxPoolDelta)
	s.Require().True(familyDelta.Add(gvgDelta).Add(taxPoolDelta).Int64() >= balanceDelta.Int64()) // could exist other buckets/objects on the gvg & family
}

func (s *PaymentTestSuite) TestStorageBill_CreateBucket_WithZeroNoneZeroReadQuota() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	params := s.queryParams()

	// case: create bucket with zero read quota
	bucketName := s.createBucket(sp, user, 0)

	// bucket created
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	_, err = s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate, streamRecordsBefore.User.NetflowRate)
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate, streamRecordsBefore.GVGFamily.NetflowRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate, streamRecordsBefore.Tax.NetflowRate)

	// case: create bucket with none zero read quota
	bucketName = s.createBucket(sp, user, 10240)

	// bucket created
	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	// check price and rate calculation
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: queryHeadBucketResponse.BucketInfo.CreateAt,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	readChargeRate := readPrice.MulInt(sdk.NewIntFromUint64(queryHeadBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate := readChargeRate.Add(taxRate)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), readChargeRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)

	expectedOutFlows := []paymenttypes.OutFlow{
		{ToAddress: family.VirtualPaymentAddress, Rate: readChargeRate},
		{ToAddress: paymenttypes.ValidatorTaxPoolAddress.String(), Rate: taxRate},
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
}

func (s *PaymentTestSuite) TestStorageBill_CreateObject_WithZeroNoneZeroPayload() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	bucketName := s.createBucket(sp, user, 0)

	// case: create object with zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize := s.createObject(user, bucketName, true)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(sp, bucketName, objectName, payloadSize, 0)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)

	// case: create object with none zero payload size
	streamRecordsBefore = s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize = s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(sp, bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

func (s *PaymentTestSuite) TestStorageBill_CreateObject_WithReserveTimeValidatorTaxRateChange() {
	defer s.revertParams()

	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	bucketName := s.createBucket(sp, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(sp, bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// update params

	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	// create another object after parameter changes
	streamRecordsBefore = s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize = s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFeeAfterParameterChange := s.calculateLockFee(sp, bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFeeAfterParameterChange)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
	s.Require().True(lockFeeAfterParameterChange.GT(lockFee.MulRaw(2)))
}

func (s *PaymentTestSuite) TestStorageBill_CancelCreateObject() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	bucketName := s.createBucket(sp, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(sp, bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// cancel create object
	s.cancelCreateObject(user, bucketName, objectName)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, lockFee)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

func (s *PaymentTestSuite) TestStorageBill_SealObject_WithoutPriceChange() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	bucketName := s.createBucket(sp, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(sp, bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// case: seal object without price change
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(sp, bucketName, objectName, payloadSize, 0)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)
}

func (s *PaymentTestSuite) TestStorageBill_SealObject_WithPriceChange() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	bucketName := s.createBucket(sp, user, 102400)

	// case: seal object with read price change and storage price change
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(2),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(2),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadBefore, taxRateReadBefore, userTotalRateReadBefore := s.calculateReadRates(sp, bucketName)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadAfter, taxRateReadAfter, userTotalRateReadAfter := s.calculateReadRatesCurrentTimestamp(sp, bucketName)
	gvgFamilyRateStore, gvgRateStore, taxRateStore, userTotalRateStore := s.calculateStorageRatesCurrentTimestamp(sp, bucketName, objectName, payloadSize)

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRateReadAfter.Sub(userTotalRateReadBefore).Add(userTotalRateStore).Neg())
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRateStore)
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRateReadAfter.Sub(gvgFamilyRateReadBefore).Add(gvgFamilyRateStore))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRateReadAfter.Sub(taxRateReadBefore).Add(taxRateStore))
}

func (s *PaymentTestSuite) TestStorageBill_SealObject_WithPriceChangeValidatorTaxRateChange() {
	defer s.revertParams()

	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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

	bucketName := s.createBucket(sp, user, 102400)

	// case: seal object with read price change and storage price change
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(2),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(2),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadBefore, taxRateReadBefore, userTotalRateReadBefore := s.calculateReadRates(sp, bucketName)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadAfter, taxRateReadAfter, userTotalRateReadAfter := s.calculateReadRatesCurrentTimestamp(sp, bucketName)
	gvgFamilyRateStore, gvgRateStore, taxRateStore, userTotalRateStore := s.calculateStorageRatesCurrentTimestamp(sp, bucketName, objectName, payloadSize)

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRateReadAfter.Sub(userTotalRateReadBefore).Add(userTotalRateStore).Neg())
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRateStore)
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRateReadAfter.Sub(gvgFamilyRateReadBefore).Add(gvgFamilyRateStore))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRateReadAfter.Sub(taxRateReadBefore).Add(taxRateStore))

	// update params
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	_, _, objectName, objectId, checksums, payloadSize = s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadAfter, taxRateReadAfter, userTotalRateReadAfter = s.calculateReadRatesCurrentTimestamp(sp, bucketName)
	gvgFamilyRateStore, gvgRateStore, taxRateStore, userTotalRateStore = s.calculateStorageRatesCurrentTimestamp(sp, bucketName, objectName, payloadSize)

	gvgFamilyRateStore = gvgFamilyRateStore.MulRaw(2)
	gvgRateStore = gvgRateStore.MulRaw(2)
	taxRateStore = taxRateStore.MulRaw(2)
	userTotalRateStore = userTotalRateStore.MulRaw(2)

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRateReadAfter.Sub(userTotalRateReadBefore).Add(userTotalRateStore).Neg())
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRateStore)
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRateReadAfter.Sub(gvgFamilyRateReadBefore).Add(gvgFamilyRateStore))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRateReadAfter.Sub(taxRateReadBefore).Add(taxRateStore))
}

func (s *PaymentTestSuite) TestStorageBill_FullLifecycle() {
	defer s.revertParams()

	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	// query storage price
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// full lifecycle
	bucketName1 := s.createBucket(sp, user, 0)
	_, _, objectName1, _, _, _ := s.createObject(user, bucketName1, true)
	_, _, objectName2, objectId2, checksums2, _ := s.createObject(user, bucketName1, false)
	s.sealObject(sp, gvg, bucketName1, objectName2, objectId2, checksums2)

	bucketName2 := s.createBucket(sp, user, 1024)
	_, _, objectName3, objectId3, checksums3, _ := s.createObject(user, bucketName2, false)
	s.sealObject(sp, gvg, bucketName2, objectName3, objectId3, checksums3)

	// update params
	params := s.queryParams()
	params.VersionedParams.ReserveTime = params.VersionedParams.ReserveTime * 3
	params.ForcedSettleTime = params.ForcedSettleTime * 2
	s.updateParams(params)

	_, _, objectName4, objectId4, checksums4, _ := s.createObject(user, bucketName2, false)
	s.sealObject(sp, gvg, bucketName2, objectName4, objectId4, checksums4)

	bucketName3 := s.createBucket(sp, user, 1024)
	_, _, objectName5, objectId5, checksums5, _ := s.createObject(user, bucketName3, false)
	s.sealObject(sp, gvg, bucketName3, objectName5, objectId5, checksums5)

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// update params
	params = s.queryParams()
	params.VersionedParams.ReserveTime = params.VersionedParams.ReserveTime / 2
	params.ForcedSettleTime = params.ForcedSettleTime / 3
	s.updateParams(params)

	_, _, objectName6, objectId6, checksums6, _ := s.createObject(user, bucketName3, false)
	s.sealObject(sp, gvg, bucketName3, objectName6, objectId6, checksums6)

	bucketName4 := s.createBucket(sp, user, 1024)
	_, _, objectName7, objectId7, checksums7, _ := s.createObject(user, bucketName4, false)
	s.sealObject(sp, gvg, bucketName4, objectName7, objectId7, checksums7)

	// update params
	params = s.queryParams()
	params.VersionedParams.ValidatorTaxRate = params.VersionedParams.ValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	_, _, objectName8, objectId8, checksums8, _ := s.createObject(user, bucketName4, false)
	s.sealObject(sp, gvg, bucketName4, objectName8, objectId8, checksums8)

	time.Sleep(3 * time.Second)

	_ = s.deleteObject(user, bucketName1, objectName1)
	_ = s.deleteObject(user, bucketName1, objectName2)

	// update params
	params = s.queryParams()
	params.VersionedParams.ValidatorTaxRate = params.VersionedParams.ValidatorTaxRate.MulInt64(3)
	s.updateParams(params)

	_ = s.deleteObject(user, bucketName2, objectName3)
	_ = s.deleteObject(user, bucketName2, objectName4)
	err = s.deleteBucket(user, bucketName1)
	s.Require().Error(err)
	err = s.deleteBucket(user, bucketName2)
	s.Require().Error(err)

	_ = s.deleteObject(user, bucketName3, objectName5)
	_ = s.deleteObject(user, bucketName3, objectName6)
	_ = s.deleteObject(user, bucketName4, objectName7)
	_ = s.deleteObject(user, bucketName4, objectName8)
	err = s.deleteBucket(user, bucketName3)
	s.Require().Error(err)
	err = s.deleteBucket(user, bucketName4)
	s.Require().Error(err)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().True(!streamRecordsAfter.User.StaticBalance.IsZero())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) TestVirtualGroup_Settle() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	bucketName := s.createBucket(sp, user, 1024)
	_, _, objectName, objectId, checksums, _ := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// sleep seconds
	time.Sleep(3 * time.Second)

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// settle gvg family
	msgSettle := virtualgrouptypes.MsgSettle{
		StorageProvider:            sp.FundingKey.GetAddr().String(),
		GlobalVirtualGroupFamilyId: family.Id,
	}
	s.SendTxBlock(sp.FundingKey, &msgSettle)

	// settle gvg
	var secondarySp *core.StorageProvider
	for _, sp := range s.StorageProviders {
		for _, id := range gvg.SecondarySpIds {
			if sp.Info.Id == id {
				secondarySp = sp
				break
			}
		}
	}
	msgSettle = virtualgrouptypes.MsgSettle{
		StorageProvider:            secondarySp.FundingKey.GetAddr().String(),
		GlobalVirtualGroupFamilyId: 0,
		GlobalVirtualGroupIds:      []uint32{gvg.Id},
	}
	s.SendTxBlock(secondarySp.FundingKey, &msgSettle)

	// assertions - balance has been checked in other tests in virtual group
	streamRecordsAfter := s.getStreamRecords(streamAddresses)

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

func (s *PaymentTestSuite) TestVirtualGroup_SwapOut() {
	ctx := context.Background()
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	successorSp := s.PickStorageProvider()

	// create a new storage provider
	sp := s.BaseSuite.CreateNewStorageProvider()
	s.T().Logf("new SP Info: %s", sp.Info.String())

	// create a new gvg group for this storage provider
	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != successorSp.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
		if len(secondarySPIDs) == 6 {
			break
		}
	}

	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(sp, 0, secondarySPIDs, 1)

	// create object
	s.BaseSuite.CreateObject(user, sp, gvgID, storagetestutils.GenRandomBucketName(), storagetestutils.GenRandomObjectName())

	// Create another gvg contains this new sp
	anotherSP := s.PickDifferentStorageProvider(successorSp.Info.Id)
	var anotherSecondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != successorSp.Info.Id && ssp.Info.Id != anotherSP.Info.Id {
			anotherSecondarySPIDs = append(anotherSecondarySPIDs, ssp.Info.Id)
		}
		if len(anotherSecondarySPIDs) == 5 {
			break
		}
	}
	anotherSecondarySPIDs = append(anotherSecondarySPIDs, sp.Info.Id)

	anotherGVGID, _ := s.BaseSuite.CreateGlobalVirtualGroup(anotherSP, 0, anotherSecondarySPIDs, 1)

	familyResp, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{FamilyId: familyID})
	s.Require().NoError(err)
	gvgResp, err := s.Client.GlobalVirtualGroup(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: anotherGVGID})
	s.Require().NoError(err)

	streamAddresses := []string{
		user.GetAddr().String(),
		familyResp.GlobalVirtualGroupFamily.VirtualPaymentAddress,
		gvgResp.GlobalVirtualGroup.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	//  sp exit
	s.SendTxBlock(sp.OperatorKey, &virtualgrouptypes.MsgStorageProviderExit{
		StorageProvider: sp.OperatorKey.GetAddr().String(),
	})

	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// swap out, as secondary sp
	msgSwapOut2 := virtualgrouptypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID}, successorSp.Info.Id)
	msgSwapOut2.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
	msgSwapOut2.SuccessorSpApproval.Sig, err = successorSp.ApprovalKey.Sign(msgSwapOut2.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(sp.OperatorKey, msgSwapOut2)

	// complete swap out
	msgCompleteSwapOut2 := virtualgrouptypes.NewMsgCompleteSwapOut(successorSp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID})
	s.Require().NoError(err)
	s.SendTxBlock(successorSp.OperatorKey, msgCompleteSwapOut2)

	// swap out, as primary sp
	msgSwapOut := virtualgrouptypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), familyID, nil, successorSp.Info.Id)
	msgSwapOut.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
	msgSwapOut.SuccessorSpApproval.Sig, err = successorSp.ApprovalKey.Sign(msgSwapOut.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(sp.OperatorKey, msgSwapOut)

	// complete swap out, as primary sp
	msgCompleteSwapOut := virtualgrouptypes.NewMsgCompleteSwapOut(successorSp.OperatorKey.GetAddr(), familyID, nil)
	s.Require().NoError(err)
	s.SendTxBlock(successorSp.OperatorKey, msgCompleteSwapOut)

	// sp complete exit success
	s.SendTxBlock(
		sp.OperatorKey,
		&virtualgrouptypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
	)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

func (s *PaymentTestSuite) TestDiscontinue_InOneBlock_WithoutPriceChange() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// create bucket
	bucketName := s.createBucket(sp, user, 0)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
	_ = s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

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

	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

func (s *PaymentTestSuite) TestDiscontinue_InOneBlock_WithPriceChange() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// query storage price
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// create bucket
	bucketName := s.createBucket(sp, user, 1200987)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

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

	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) TestDiscontinue_InBlocks_WithoutPriceChange() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// create bucket
	bucketName := s.createBucket(sp, user, 12780)

	// create & seal objects
	for i := 0; i < 4; i++ {
		_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
		_ = s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
		time.Sleep(200 * time.Millisecond)
	}

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
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
}

// TestDiscontinue_InBlocks_WithPriceChange will cover the following case:
// create an object, sp increase the price a lot, the object can be force deleted even the object's own has no enough balance.
func (s *PaymentTestSuite) TestDiscontinue_InBlocks_WithPriceChange() {
	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// query storage price
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// create bucket
	bucketName := s.createBucket(sp, user, 0)

	// create objects
	for i := 0; i < 2; i++ {
		_, _, objectName1, _, _, _ := s.createObject(user, bucketName, false)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
		time.Sleep(200 * time.Millisecond)
	}

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// create & seal objects
	for i := 0; i < 2; i++ {
		_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
		s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
		time.Sleep(200 * time.Millisecond)
	}

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

	// for payment
	time.Sleep(2 * time.Second)

	// force delete bucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueBucket)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	// update new price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(100),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

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
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) TestDiscontinue_InBlocks_WithPriceChangeReserveTimeChange() {
	defer s.revertParams()

	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
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
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// query storage price
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// create bucket
	bucketName := s.createBucket(sp, user, 10200)

	// create objects
	for i := 0; i < 2; i++ {
		_, _, objectName1, _, _, _ := s.createObject(user, bucketName, false)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
		time.Sleep(200 * time.Millisecond)
	}

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// create & seal objects
	for i := 0; i < 2; i++ {
		_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
		s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
		time.Sleep(200 * time.Millisecond)
	}

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
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	// for payment
	time.Sleep(2 * time.Second)

	// force delete bucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueBucket)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	// update new price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(100),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

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
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().True(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64() <= int64(0)) // there are other auto settling

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func (s *PaymentTestSuite) TestDiscontinue_InBlocks_WithPriceChangeReserveTimeChange_FrozenAccount() {
	defer s.revertParams()

	ctx := context.Background()
	sp := s.PickStorageProvider()
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	// params
	params := s.queryParams()
	oldReserveTime := params.VersionedParams.ReserveTime
	oldValidatorTaxRate := params.VersionedParams.ValidatorTaxRate

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// query storage price
	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.SpStoragePrice)

	// create bucket
	bucketName := s.createBucket(sp, user, 0)

	// create objects
	for i := 0; i < 2; i++ {
		_, _, objectName1, _, _, _ := s.createObject(user, bucketName, false)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
		time.Sleep(200 * time.Millisecond)
	}

	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// update params
	params.VersionedParams.ReserveTime = 8
	params.ForcedSettleTime = 5
	s.updateParams(params)

	// create & seal objects
	for i := 0; i < 2; i++ {
		_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
		s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName1,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
		time.Sleep(200 * time.Millisecond)
	}

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

	// wait account to be frozen
	time.Sleep(8 * time.Second)
	streamRecord := s.getStreamRecord(user.GetAddr().String())
	s.Require().True(streamRecord.Status == paymenttypes.STREAM_ACCOUNT_STATUS_FROZEN)

	// update params
	params.VersionedParams.ReserveTime = oldReserveTime * 2
	params.VersionedParams.ValidatorTaxRate = oldValidatorTaxRate.MulInt64(2)
	s.updateParams(params)

	// for payment
	time.Sleep(2 * time.Second)

	// force delete bucket
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(sp.GcKey, msgDiscontinueBucket)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	// update new price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice.MulInt64(100),
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(10000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

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
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().True(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64() <= int64(0)) // there are other auto settling

	// revert price
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
}

func TestPaymentTestSuite(t *testing.T) {
	suite.Run(t, new(PaymentTestSuite))
}

func (s *PaymentTestSuite) getStreamRecord(addr string) (sr paymenttypes.StreamRecord) {
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

func (s *PaymentTestSuite) getStreamRecords(addrs []string) (streamRecords StreamRecords) {
	streamRecords.User = s.getStreamRecord(addrs[0])
	streamRecords.GVGFamily = s.getStreamRecord(addrs[1])
	streamRecords.GVG = s.getStreamRecord(addrs[2])
	streamRecords.Tax = s.getStreamRecord(addrs[3])
	s.T().Logf("streamRecords: %s", core.YamlString(streamRecords))
	return
}

func (s *PaymentTestSuite) checkStreamRecordsBeforeAndAfter(streamRecordsBefore StreamRecords, streamRecordsAfter StreamRecords, readPrice sdk.Dec,
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

func (s *PaymentTestSuite) getChargeSize(payloadSize uint64) uint64 {
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

func (s *PaymentTestSuite) calculateLockFee(sp *core.StorageProvider, bucketName, objectName string, payloadSize uint64) sdkmath.Int {
	ctx := context.Background()

	params := s.queryParams()

	headBucketExtraResponse, err := s.Client.HeadBucketExtra(ctx, &storagetypes.QueryHeadBucketExtraRequest{BucketName: bucketName})
	s.Require().NoError(err)

	storageParams, err := s.Client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	s.T().Logf("storageParams %s, err: %v", storageParams, err)
	s.Require().NoError(err)
	secondarySpCount := storageParams.Params.VersionedParams.RedundantDataChunkNum + storageParams.Params.VersionedParams.RedundantParityChunkNum

	chargeSize := s.getChargeSize(payloadSize)
	_, primaryPrice, secondaryPrice := s.getPrices(sp, headBucketExtraResponse.ExtraInfo.PriceTime)

	gvgFamilyRate := primaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate := secondaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate = gvgRate.MulRaw(int64(secondarySpCount))
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate.Add(gvgRate)).TruncateInt()
	return gvgFamilyRate.Add(gvgRate).Add(taxRate).MulRaw(int64(params.VersionedParams.ReserveTime))
}

func (s *PaymentTestSuite) getPrices(sp *core.StorageProvider, timestamp int64) (sdk.Dec, sdk.Dec, sdk.Dec) {
	ctx := context.Background()

	spStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: timestamp,
	})
	s.T().Logf("spStoragePriceByTimeResp %s, err: %v", spStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	secondaryStoragePriceByTimeResp, err := s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: timestamp,
	})
	s.T().Logf("spStoragePriceByTimeResp %s, err: %v", spStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	return spStoragePriceByTimeResp.SpStoragePrice.ReadPrice, spStoragePriceByTimeResp.SpStoragePrice.StorePrice,
		secondaryStoragePriceByTimeResp.SecondarySpStorePrice.StorePrice
}

func (s *PaymentTestSuite) calculateReadRates(sp *core.StorageProvider, bucketName string) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	ctx := context.Background()

	params := s.queryParams()

	headBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	headBucketResponse, err := s.Client.HeadBucket(ctx, &headBucketRequest)
	s.Require().NoError(err)

	readPrice, _, _ := s.getPrices(sp, headBucketResponse.BucketInfo.CreateAt)

	gvgFamilyRate := readPrice.MulInt64(int64(headBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate).TruncateInt()
	return gvgFamilyRate, taxRate, gvgFamilyRate.Add(taxRate)
}

func (s *PaymentTestSuite) calculateReadRatesCurrentTimestamp(sp *core.StorageProvider, bucketName string) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	ctx := context.Background()

	params := s.queryParams()

	headBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	headBucketResponse, err := s.Client.HeadBucket(ctx, &headBucketRequest)
	s.Require().NoError(err)

	readPrice, _, _ := s.getPrices(sp, time.Now().Unix())

	gvgFamilyRate := readPrice.MulInt64(int64(headBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate).TruncateInt()
	return gvgFamilyRate, taxRate, gvgFamilyRate.Add(taxRate)
}

func (s *PaymentTestSuite) calculateStorageRates(sp *core.StorageProvider, bucketName, objectName string, payloadSize uint64, priceTime int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	params := s.queryParams()

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	headObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	secondarySpCount := len(headObjectResponse.GlobalVirtualGroup.SecondarySpIds)
	fmt.Println("secondarySpCount", secondarySpCount)
	if priceTime == 0 {
		headBucketRequest := storagetypes.QueryHeadBucketRequest{
			BucketName: bucketName,
		}
		headBucketResponse, err := s.Client.HeadBucket(context.Background(), &headBucketRequest)
		s.Require().NoError(err)
		priceTime = headBucketResponse.BucketInfo.CreateAt
	}

	chargeSize := s.getChargeSize(payloadSize)
	_, primaryPrice, secondaryPrice := s.getPrices(sp, priceTime)

	gvgFamilyRate := primaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate := secondaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate = gvgRate.MulRaw(int64(secondarySpCount))
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate.Add(gvgRate)).TruncateInt()
	return gvgFamilyRate, gvgRate, taxRate, gvgFamilyRate.Add(gvgRate).Add(taxRate)
}

func (s *PaymentTestSuite) calculateStorageRatesCurrentTimestamp(sp *core.StorageProvider, bucketName, objectName string, payloadSize uint64) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	params := s.queryParams()

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	headObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	secondarySpCount := len(headObjectResponse.GlobalVirtualGroup.SecondarySpIds)
	fmt.Println("secondarySpCount", secondarySpCount)

	chargeSize := s.getChargeSize(payloadSize)
	_, primaryPrice, secondaryPrice := s.getPrices(sp, time.Now().Unix())

	gvgFamilyRate := primaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate := secondaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate = gvgRate.MulRaw(int64(secondarySpCount))
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate.Add(gvgRate)).TruncateInt()
	return gvgFamilyRate, gvgRate, taxRate, gvgFamilyRate.Add(gvgRate).Add(taxRate)
}

func (s *PaymentTestSuite) updateParams(params paymenttypes.Params) {
	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()

	ts := time.Now().Unix()
	queryParamsRequest := &paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(ctx, queryParamsRequest)
	s.Require().NoError(err)
	s.T().Log("params before", core.YamlString(queryParamsResponse.Params))

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
	s.T().Log("params by timestamp", core.YamlString(queryParamsResponse.Params))
	s.Require().Equal(queryParamsResponse.Params.VersionedParams.ReserveTime,
		queryParamsByTimestampResponse.Params.VersionedParams.ReserveTime)

	queryParamsRequest = &paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err = s.Client.PaymentQueryClient.Params(ctx, queryParamsRequest)
	s.Require().NoError(err)
	s.T().Log("params after", core.YamlString(queryParamsResponse.Params))
}

func (s *PaymentTestSuite) createBucketAndObject(sp *core.StorageProvider) (keys.KeyManager, string, string, storagetypes.Uint, [][]byte) {
	var err error
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

func (s *PaymentTestSuite) createBucket(sp *core.StorageProvider, user keys.KeyManager, readQuota uint64) string {
	var err error
	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)

	// CreateBucket
	bucketName := "ch" + storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, readQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	_, err = s.Client.HeadBucket(context.Background(), &queryHeadBucketRequest)
	s.Require().NoError(err)

	return bucketName
}

func (s *PaymentTestSuite) createObject(user keys.KeyManager, bucketName string, empty bool) (keys.KeyManager, string, string, storagetypes.Uint, [][]byte, uint64) {
	var err error
	sp := s.BaseSuite.PickStorageProviderByBucketName(bucketName)

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
	if !empty {
		for i := 0; i < 1024; i++ {
			buffer.WriteString(fmt.Sprintf("[%05d] %s\n", i, line))
		}
	}
	payloadSize := uint64(buffer.Len())
	checksum := sdk.Keccak256(buffer.Bytes())
	expectChecksum := [][]byte{checksum, checksum, checksum, checksum, checksum, checksum, checksum}
	contextType := "text/event-stream"
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, payloadSize,
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
	headObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(headObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(headObjectResponse.ObjectInfo.BucketName, bucketName)

	return user, bucketName, objectName, headObjectResponse.ObjectInfo.Id, expectChecksum, payloadSize
}

func (s *PaymentTestSuite) cancelCreateObject(user keys.KeyManager, bucketName, objectName string) {
	msgCancelCreateObject := storagetypes.NewMsgCancelCreateObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgCancelCreateObject)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	_, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().Error(err)
}

func (s *PaymentTestSuite) sealObject(sp *core.StorageProvider, gvg *virtualgrouptypes.GlobalVirtualGroup, bucketName, objectName string, objectId storagetypes.Uint, checksums [][]byte) *virtualgrouptypes.GlobalVirtualGroup {
	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)
	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, objectId, storagetypes.GenerateHash(checksums[:])).GetBlsSignHash()
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

	queryHeadObjectRequest2 := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse2, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest2)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse2.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse2.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse2.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	return gvg
}

func (s *PaymentTestSuite) deleteObject(user keys.KeyManager, bucketName, objectName string) error {
	msgDeleteObject := storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName)
	s.SendTxBlock(user, msgDeleteObject)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	_, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	return err
}

func (s *PaymentTestSuite) deleteBucket(user keys.KeyManager, bucketName string) error {
	msgDeleteObject := storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName)
	s.SendTxBlock(user, msgDeleteObject)

	// HeadObject
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	_, err := s.Client.HeadBucket(context.Background(), &queryHeadBucketRequest)
	return err
}

func (s *PaymentTestSuite) TestUpdatePaymentParams() {
	// 1. create proposal
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	queryParamsResp, err := s.Client.PaymentQueryClient.Params(context.Background(), &paymenttypes.QueryParamsRequest{})
	s.Require().NoError(err)

	updatedParams := queryParamsResp.Params
	updatedParams.PaymentAccountCountLimit = 300
	msgUpdateParams := &paymenttypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    updatedParams,
	}

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update Payment params", "Test update Payment params")
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

	updatedQueryParamsResp, err := s.Client.PaymentQueryClient.Params(context.Background(), &paymenttypes.QueryParamsRequest{})
	s.Require().NoError(err)
	if reflect.DeepEqual(updatedQueryParamsResp.Params, updatedParams) {
		s.T().Logf("update params success")
	} else {
		s.T().Errorf("update params failed")
	}
}

func (s *PaymentTestSuite) revertParams() {
	s.updateParams(s.defaultParams)
}

func (s *PaymentTestSuite) queryParams() paymenttypes.Params {
	queryParamsRequest := paymenttypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.PaymentQueryClient.Params(context.Background(), &queryParamsRequest)
	s.Require().NoError(err)
	s.T().Log("params", core.YamlString(queryParamsResponse.Params))
	return queryParamsResponse.Params
}
