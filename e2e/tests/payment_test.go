package tests

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	"github.com/cosmos/cosmos-sdk/types/query"
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

func (s *PaymentTestSuite) TestQueryPaymentAccounts() {
	_, err := s.Client.PaymentAccounts(context.Background(), &paymenttypes.QueryPaymentAccountsRequest{
		Pagination: &query.PageRequest{
			Offset: 10, // offset is not allowed
		},
	})
	s.Require().Error(err)

	_, err = s.Client.PaymentAccounts(context.Background(), &paymenttypes.QueryPaymentAccountsRequest{
		Pagination: &query.PageRequest{
			CountTotal: true, // count total = true is not allowed
		},
	})
	s.Require().Error(err)

	_, err = s.Client.PaymentAccounts(context.Background(), &paymenttypes.QueryPaymentAccountsRequest{
		Pagination: &query.PageRequest{},
	})
	s.Require().NoError(err)
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
	queryGetPaymentAccountsByOwnerRequest := paymenttypes.QueryPaymentAccountsByOwnerRequest{
		Owner: user.GetAddr().String(),
	}
	paymentAccounts, err := s.Client.PaymentAccountsByOwner(ctx, &queryGetPaymentAccountsByOwnerRequest)
	s.Require().NoError(err)
	s.T().Log(paymentAccounts)
	s.Require().Equal(1, len(paymentAccounts.PaymentAccounts))
	paymentAccountAddr := paymentAccounts.PaymentAccounts[0]
	// query this payment account
	queryGetPaymentAccountRequest := paymenttypes.QueryPaymentAccountRequest{
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
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject(sp, gvg)

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
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject(sp, gvg)

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
	user, bucketName, objectName, objectId, checksums := s.createBucketAndObject(sp, gvg)

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
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
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

func (s *PaymentTestSuite) TestDeposit_FromBankAccount() {
	ctx := context.Background()
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	userAddr := user.GetAddr().String()
	var err error

	// derive payment account
	paymentAccount := derivePaymentAccount(user.GetAddr(), 0)
	// transfer BNB to derived payment account
	msgSend := banktypes.NewMsgSend(user.GetAddr(), paymentAccount, sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdk.NewInt(1e18)),
	))
	_ = s.SendTxBlock(user, msgSend)

	paymentBalanceBefore, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: paymentAccount.String(),
		Denom:   s.Config.Denom,
	})
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(1e18).String(), paymentBalanceBefore.GetBalance().Amount.String())

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: userAddr,
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
	s.Require().NoError(err)
	s.T().Logf("paymentAccounts %s", core.YamlString(paymentAccounts))
	paymentAddr := paymentAccounts.PaymentAccounts[0]
	s.Require().Lenf(paymentAccounts.PaymentAccounts, 1, "paymentAccounts %s", core.YamlString(paymentAccounts))

	// transfer BNB to payment account: should not success
	msgSend = banktypes.NewMsgSend(user.GetAddr(), sdk.MustAccAddressFromHex(paymentAddr), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdk.NewInt(1e18)),
	))
	s.SendTxBlockWithExpectErrorString(msgSend, user, "is not allowed to receive funds")

	// deposit BNB needed
	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAddr,
		Amount:  sdk.NewInt(1e18), // deposit more than needed
	}
	_ = s.SendTxBlock(user, msgDeposit)

	paymentBalanceAfter, err := s.Client.Balance(ctx, &banktypes.QueryBalanceRequest{
		Address: paymentAddr,
		Denom:   s.Config.Denom,
	})
	s.Require().NoError(err)
	s.Require().Equal(sdk.NewInt(0).String(), paymentBalanceAfter.GetBalance().Amount.String())
}

func derivePaymentAccount(owner sdk.AccAddress, index uint64) sdk.AccAddress {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, index)
	return address.Derive(owner.Bytes(), b)[:sdk.EthAddressLength]
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
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
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
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
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
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
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
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(1000)
	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
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
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	bucketChargedReadQuota := uint64(100000)
	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: userAddr}
	paymentAccounts, err := s.Client.PaymentQueryClient.PaymentAccountsByOwner(ctx, paymentAccountsReq)
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

	bucketName := s.createBucket(sp, gvg, user, 1024)
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
	bucketName := s.createBucket(sp, gvg, user, 0)

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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// create bucket
	bucketName := s.createBucket(sp, gvg, user, 1200987)

	// create & seal objects
	_, _, objectName1, objectId1, checksums1, _ := s.createObject(user, bucketName, false)
	s.sealObject(sp, gvg, bucketName, objectName1, objectId1, checksums1)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

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
	bucketName := s.createBucket(sp, gvg, user, 12780)

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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// create bucket
	bucketName := s.createBucket(sp, gvg, user, 0)

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
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

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
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(100), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))

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

func (s *PaymentTestSuite) TestDiscontinue_InBlocks_WithPriceChangeReserveTimeChange() {
	defer s.revertParams()

	ctx := context.Background()
	sp := s.PickStorageProvider()

	_, secondarySps := s.GetSecondarySP(sp)
	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(sp, 0, secondarySps, 1)
	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &virtualgrouptypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	gvg := gvgResp.GlobalVirtualGroup

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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// create bucket
	bucketName := s.createBucket(sp, gvg, user, 10200)

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
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

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
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(100), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))

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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// create bucket
	bucketName := s.createBucket(sp, gvg, user, 0)

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
	userStreamRecord := s.getStreamRecord(user.GetAddr().String())
	s.Require().True(userStreamRecord.LockBalance.IsPositive())

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

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
		userStream := s.getStreamRecord(user.GetAddr().String())
		s.Require().True(userStream.LockBalance.IsPositive())
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
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(100), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))

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

	s.Require().True(streamRecordsAfter.User.LockBalance.IsZero())
	s.Require().True(streamRecordsAfter.User.StaticBalance.Int64() == userStreamRecord.LockBalance.Int64())
}

func (s *PaymentTestSuite) TestDiscontinue_MultiObjects() {
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
	bucketName := s.createBucket(sp, gvg, user, 0)
	objectIds := []sdkmath.Uint{}

	// create objects
	for i := 0; i < 3; i++ {
		_, _, objectName, objectId, _, _ := s.createObject(user, bucketName, false)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
		time.Sleep(200 * time.Millisecond)
		objectIds = append(objectIds, objectId)
	}

	// create & seal objects
	for i := 0; i < 3; i++ {
		_, _, objectName, objectId, checksums, _ := s.createObject(user, bucketName, false)
		s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName,
			ObjectName: objectName,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
		time.Sleep(200 * time.Millisecond)
		objectIds = append(objectIds, objectId)
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

	// force objects
	msgs := make([]sdk.Msg, 0)
	for _, id := range objectIds {
		msgDiscontinueObject := storagetypes.NewMsgDiscontinueObject(sp.GcKey.GetAddr(), bucketName, []sdkmath.Uint{id}, "test")
		msgs = append(msgs, msgDiscontinueObject)
	}
	msgs = append(msgs, storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test"))
	txRes := s.SendTxBlock(sp.GcKey, msgs...)
	deleteAt := filterDiscontinueObjectEventFromTx(txRes).DeleteAt

	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime > deleteAt+5 {
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
	s.Require().True(streamRecordsAfter.User.LockBalance.IsZero())
}

func (s *PaymentTestSuite) TestDiscontinue_MultiBuckets() {
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

	bucketNames := []string{}
	// create bucket
	bucketName1 := s.createBucket(sp, gvg, user, 1023)
	bucketNames = append(bucketNames, bucketName1)

	// create objects
	for i := 0; i < 2; i++ {
		_, _, objectName, _, _, _ := s.createObject(user, bucketName1, false)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName1,
			ObjectName: objectName,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_CREATED)
		time.Sleep(200 * time.Millisecond)
	}

	// create & seal objects
	for i := 0; i < 2; i++ {
		_, _, objectName, objectId, checksums, _ := s.createObject(user, bucketName1, false)
		s.sealObject(sp, gvg, bucketName1, objectName, objectId, checksums)
		queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
			BucketName: bucketName1,
			ObjectName: objectName,
		}
		queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
		s.Require().NoError(err)
		s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)
		time.Sleep(200 * time.Millisecond)
	}

	// create bucket
	bucketName2 := s.createBucket(sp, gvg, user, 21023)
	bucketNames = append(bucketNames, bucketName2)

	// create bucket
	bucketName3 := s.createBucket(sp, gvg, user, 0)
	bucketNames = append(bucketNames, bucketName3)

	// create bucket
	bucketName4 := s.createBucket(sp, gvg, user, 55)
	bucketNames = append(bucketNames, bucketName4)

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

	// force objects
	msgs := make([]sdk.Msg, 0)
	for _, bucketName := range bucketNames {
		msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(sp.GcKey.GetAddr(), bucketName, "test")
		msgs = append(msgs, msgDiscontinueBucket)
	}
	txRes := s.SendTxBlock(sp.GcKey, msgs...)
	deleteAt := filterDiscontinueBucketEventFromTx(txRes).DeleteAt

	for {
		time.Sleep(200 * time.Millisecond)
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		blockTime := statusRes.SyncInfo.LatestBlockTime.Unix()

		s.T().Logf("current blockTime: %d, delete blockTime: %d", blockTime, deleteAt)

		if blockTime > deleteAt+5 {
			break
		}
	}

	for _, bucketName := range bucketNames {
		_, err = s.Client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucketName})
		s.Require().ErrorContains(err, "No such bucket")
	}
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
	s.Require().True(streamRecordsAfter.User.LockBalance.IsZero())
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

func (s *PaymentTestSuite) checkStreamRecordsBeforeAndAfter(streamRecordsBefore, streamRecordsAfter StreamRecords, readPrice sdk.Dec,
	readChargeRate sdkmath.Int, primaryStorePrice, secondaryStorePrice sdk.Dec, chargeSize, payloadSize uint64,
) {
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

func (s *PaymentTestSuite) calculateLockFee(bucketName, objectName string, payloadSize uint64) sdkmath.Int {
	ctx := context.Background()

	params := s.queryParams()

	headBucketExtraResponse, err := s.Client.HeadBucketExtra(ctx, &storagetypes.QueryHeadBucketExtraRequest{BucketName: bucketName})
	s.Require().NoError(err)

	storageParams, err := s.Client.StorageQueryClient.Params(ctx, &storagetypes.QueryParamsRequest{})
	s.T().Logf("storageParams %s, err: %v", storageParams, err)
	s.Require().NoError(err)
	secondarySpCount := storageParams.Params.VersionedParams.RedundantDataChunkNum + storageParams.Params.VersionedParams.RedundantParityChunkNum

	chargeSize := s.getChargeSize(payloadSize)
	_, primaryPrice, secondaryPrice := s.getPrices(headBucketExtraResponse.ExtraInfo.PriceTime)

	gvgFamilyRate := primaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate := secondaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate = gvgRate.MulRaw(int64(secondarySpCount))
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate.Add(gvgRate)).TruncateInt()
	return gvgFamilyRate.Add(gvgRate).Add(taxRate).MulRaw(int64(params.VersionedParams.ReserveTime))
}

func (s *PaymentTestSuite) getPrices(timestamp int64) (sdk.Dec, sdk.Dec, sdk.Dec) {
	ctx := context.Background()

	spStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: timestamp,
	})
	s.T().Logf("spStoragePriceByTimeResp %s, err: %v", spStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	return spStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice, spStoragePriceByTimeResp.GlobalSpStorePrice.PrimaryStorePrice,
		spStoragePriceByTimeResp.GlobalSpStorePrice.SecondaryStorePrice
}

func (s *PaymentTestSuite) calculateReadRates(bucketName string) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	ctx := context.Background()

	params := s.queryParams()

	headBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	headBucketResponse, err := s.Client.HeadBucket(ctx, &headBucketRequest)
	s.Require().NoError(err)

	readPrice, _, _ := s.getPrices(headBucketResponse.BucketInfo.CreateAt)

	gvgFamilyRate := readPrice.MulInt64(int64(headBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate).TruncateInt()
	return gvgFamilyRate, taxRate, gvgFamilyRate.Add(taxRate)
}

func (s *PaymentTestSuite) calculateReadRatesCurrentTimestamp(bucketName string) (sdkmath.Int, sdkmath.Int, sdkmath.Int) {
	ctx := context.Background()

	params := s.queryParams()

	headBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	headBucketResponse, err := s.Client.HeadBucket(ctx, &headBucketRequest)
	s.Require().NoError(err)

	readPrice, _, _ := s.getPrices(time.Now().Unix())

	gvgFamilyRate := readPrice.MulInt64(int64(headBucketResponse.BucketInfo.ChargedReadQuota)).TruncateInt()
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate).TruncateInt()
	return gvgFamilyRate, taxRate, gvgFamilyRate.Add(taxRate)
}

func (s *PaymentTestSuite) calculateStorageRates(bucketName, objectName string, payloadSize uint64, priceTime int64) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
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
	_, primaryPrice, secondaryPrice := s.getPrices(priceTime)
	s.T().Logf("===secondaryPrice: %v,primaryPrice: %v===", secondaryPrice, primaryPrice)
	gvgFamilyRate := primaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate := secondaryPrice.MulInt(sdkmath.NewIntFromUint64(chargeSize)).TruncateInt()
	gvgRate = gvgRate.MulRaw(int64(secondarySpCount))
	taxRate := params.VersionedParams.ValidatorTaxRate.MulInt(gvgFamilyRate.Add(gvgRate)).TruncateInt()
	return gvgFamilyRate, gvgRate, taxRate, gvgFamilyRate.Add(gvgRate).Add(taxRate)
}

func (s *PaymentTestSuite) calculateStorageRatesCurrentTimestamp(bucketName, objectName string, payloadSize uint64) (sdkmath.Int, sdkmath.Int, sdkmath.Int, sdkmath.Int) {
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
	_, primaryPrice, secondaryPrice := s.getPrices(time.Now().Unix())

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

func (s *PaymentTestSuite) createBucketAndObject(sp *core.StorageProvider, gvg *virtualgrouptypes.GlobalVirtualGroup) (keys.KeyManager, string, string, storagetypes.Uint, [][]byte) {
	var err error
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

func (s *PaymentTestSuite) createBucket(sp *core.StorageProvider, gvg *virtualgrouptypes.GlobalVirtualGroup, user keys.KeyManager, readQuota uint64) string {
	var err error
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

func (s *PaymentTestSuite) rejectSealObject(sp *core.StorageProvider, gvg *virtualgrouptypes.GlobalVirtualGroup, bucketName, objectName string) {
	msgRejectSealObject := storagetypes.NewMsgRejectUnsealedObject(sp.SealKey.GetAddr(), bucketName, objectName)

	s.T().Logf("msg %s", msgRejectSealObject.String())
	s.SendTxBlock(sp.SealKey, msgRejectSealObject)

	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	_, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	s.Require().Error(err)
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

func (s *PaymentTestSuite) updateGlobalSpPrice(readPrice, storePrice sdk.Dec) {
	ctx := context.Background()
	globalPriceResBefore, _ := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{Timestamp: 0})
	s.T().Log("globalPriceResBefore", core.YamlString(globalPriceResBefore))

	for _, sp := range s.BaseSuite.StorageProviders {
		msgUpdateSpStoragePrice := &sptypes.MsgUpdateSpStoragePrice{
			SpAddress:     sp.OperatorKey.GetAddr().String(),
			ReadPrice:     readPrice,
			StorePrice:    storePrice,
			FreeReadQuota: 1024 * 1024,
		}
		s.SendTxBlock(sp.OperatorKey, msgUpdateSpStoragePrice)
	}
	time.Sleep(2 * time.Second)

	globalPriceResAfter, _ := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{Timestamp: 0})
	s.T().Log("globalPriceResAfter1", core.YamlString(globalPriceResAfter))
}
