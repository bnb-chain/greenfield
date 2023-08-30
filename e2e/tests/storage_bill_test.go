package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	storagetestutils "github.com/bnb-chain/greenfield/testutil/storage"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgrouptypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

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

	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: queryHeadBucketResponse.BucketInfo.CreateAt,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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

	queryGlobalSpStorePriceByTime, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: queryHeadBucketResponse.BucketInfo.CreateAt,
	})
	s.T().Logf("queryGlobalSpStorePriceByTime %s, err: %v", queryGlobalSpStorePriceByTime, err)
	s.Require().NoError(err)
	primaryStorePrice := queryGlobalSpStorePriceByTime.GlobalSpStorePrice.PrimaryStorePrice
	secondaryStorePrice := queryGlobalSpStorePriceByTime.GlobalSpStorePrice.SecondaryStorePrice
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
	queryAllAutoSettleRecordRequest := paymenttypes.QueryAutoSettleRecordsRequest{}
	queryAllAutoSettleRecordResponse, err := s.Client.AutoSettleRecords(ctx, &queryAllAutoSettleRecordRequest)
	s.Require().NoError(err)
	s.T().Logf("queryAllAutoSettleRecordResponse %s", core.YamlString(queryAllAutoSettleRecordResponse))
	s.Require().True(len(queryAllAutoSettleRecordResponse.AutoSettleRecords) >= 1)

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
	bucketName := s.createBucket(sp, gvg, user, 256)

	// simulate delete bucket gas
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

	// transfer gas
	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdkmath.NewInt(5*types.DecimalGwei)),
	))
	simulateResponse = s.SimulateTx(msgSend, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	// delete object gas
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
	bucketName := s.createBucket(sp, gvg, user, 256)

	// simulate delete bucket gas
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

	// transfer gas
	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdkmath.NewInt(5*types.DecimalGwei)),
	))
	simulateResponse = s.SimulateTx(msgSend, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	// delete object gas
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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(1000), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

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
	bucketName := s.createBucket(sp, gvg, user, 256)

	// simulate delete bucket gas
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

	// transfer gas
	msgSend := banktypes.NewMsgSend(user.GetAddr(), core.GenRandomAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sdkmath.NewInt(5*types.DecimalGwei)),
	))
	simulateResponse = s.SimulateTx(msgSend, user)
	gasLimit = simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err = sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)

	gas = gas.Add(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit))))
	s.T().Log("total gas", "gas", gas)

	// delete object gas
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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(1000), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

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
	bucketName := s.createBucket(sp, gvg, user, 256)

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
	_, _, userRateRead := s.calculateReadRates(bucketName)
	_, _, _, userRateStore := s.calculateStorageRates(bucketName, objectName1, payloadSize, 0)

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
	bucketName := s.createBucket(sp, gvg, user, 256)

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
	_, _, userRateRead := s.calculateReadRates(bucketName)
	_, _, _, userRateStore := s.calculateStorageRates(bucketName, objectName1, payloadSize, 0)

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
	bucketName := s.createBucket(sp, gvg, user, 0)

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
	bucketName = s.createBucket(sp, gvg, user, 10240)

	// bucket created
	queryHeadBucketRequest = storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)

	// check price and rate calculation
	queryGlobalSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: queryHeadBucketResponse.BucketInfo.CreateAt,
	})
	s.T().Logf("queryGlobalSpStoragePriceByTimeResp %s, err: %v", queryGlobalSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGlobalSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
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

	bucketName := s.createBucket(sp, gvg, user, 0)

	// case: create object with zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize := s.createObject(user, bucketName, true)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(bucketName, objectName, payloadSize, 0)
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
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
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

	bucketName := s.createBucket(sp, gvg, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
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
	lockFeeAfterParameterChange := s.calculateLockFee(bucketName, objectName, payloadSize)
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

	bucketName := s.createBucket(sp, gvg, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, _, _, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
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

	bucketName := s.createBucket(sp, gvg, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
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
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(bucketName, objectName, payloadSize, 0)
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

	bucketName := s.createBucket(sp, gvg, user, 102400)

	// case: seal object with read price change and storage price change
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(2), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(2))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadBefore, taxRateReadBefore, userTotalRateReadBefore := s.calculateReadRates(bucketName)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadAfter, taxRateReadAfter, userTotalRateReadAfter := s.calculateReadRatesCurrentTimestamp(bucketName)
	gvgFamilyRateStore, gvgRateStore, taxRateStore, userTotalRateStore := s.calculateStorageRatesCurrentTimestamp(bucketName, objectName, payloadSize)

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

	bucketName := s.createBucket(sp, gvg, user, 102400)

	// case: seal object with read price change and storage price change
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(2), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(2))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadBefore, taxRateReadBefore, userTotalRateReadBefore := s.calculateReadRates(bucketName)

	// seal object
	s.sealObject(sp, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	gvgFamilyRateReadAfter, taxRateReadAfter, userTotalRateReadAfter := s.calculateReadRatesCurrentTimestamp(bucketName)
	gvgFamilyRateStore, gvgRateStore, taxRateStore, userTotalRateStore := s.calculateStorageRatesCurrentTimestamp(bucketName, objectName, payloadSize)

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
	gvgFamilyRateReadAfter, taxRateReadAfter, userTotalRateReadAfter = s.calculateReadRatesCurrentTimestamp(bucketName)
	gvgFamilyRateStore, gvgRateStore, taxRateStore, userTotalRateStore = s.calculateStorageRatesCurrentTimestamp(bucketName, objectName, payloadSize*2)

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRateReadAfter.Sub(userTotalRateReadBefore).Add(userTotalRateStore).Neg())
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRateStore)
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRateReadAfter.Sub(gvgFamilyRateReadBefore).Add(gvgFamilyRateStore))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRateReadAfter.Sub(taxRateReadBefore).Add(taxRateStore))
}

func (s *PaymentTestSuite) TestStorageBill_RejectSealObject_WithPriceChange() {
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

	bucketName := s.createBucket(sp, gvg, user, 102400)

	// case: seal object with read price change and storage price change
	_, _, objectName, _, _, _ := s.createObject(user, bucketName, false)

	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(2), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(2))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	s.Require().True(streamRecordsBefore.User.LockBalance.IsPositive())

	// reject seal object
	s.rejectSealObject(sp, gvg, bucketName, objectName)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().True(streamRecordsAfter.User.StaticBalance.IsPositive())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())

	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))
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
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	// full lifecycle
	bucketName1 := s.createBucket(sp, gvg, user, 0)
	_, _, objectName1, _, _, _ := s.createObject(user, bucketName1, true)
	_, _, objectName2, objectId2, checksums2, _ := s.createObject(user, bucketName1, false)
	s.sealObject(sp, gvg, bucketName1, objectName2, objectId2, checksums2)

	bucketName2 := s.createBucket(sp, gvg, user, 1024)
	_, _, objectName3, objectId3, checksums3, _ := s.createObject(user, bucketName2, false)
	s.sealObject(sp, gvg, bucketName2, objectName3, objectId3, checksums3)

	// update params
	params := s.queryParams()
	params.VersionedParams.ReserveTime = params.VersionedParams.ReserveTime * 3
	params.ForcedSettleTime = params.ForcedSettleTime * 2
	s.updateParams(params)

	_, _, objectName4, objectId4, checksums4, _ := s.createObject(user, bucketName2, false)
	s.sealObject(sp, gvg, bucketName2, objectName4, objectId4, checksums4)

	bucketName3 := s.createBucket(sp, gvg, user, 1024)
	_, _, objectName5, objectId5, checksums5, _ := s.createObject(user, bucketName3, false)
	s.sealObject(sp, gvg, bucketName3, objectName5, objectId5, checksums5)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(50), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	// update params
	params = s.queryParams()
	params.VersionedParams.ReserveTime = params.VersionedParams.ReserveTime / 2
	params.ForcedSettleTime = params.ForcedSettleTime / 3
	s.updateParams(params)

	_, _, objectName6, objectId6, checksums6, _ := s.createObject(user, bucketName3, false)
	s.sealObject(sp, gvg, bucketName3, objectName6, objectId6, checksums6)

	bucketName4 := s.createBucket(sp, gvg, user, 1024)
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
}

// TestStorageBill_CopyObject_WithoutPriceChange
func (s *PaymentTestSuite) TestStorageBill_CopyObject_WithoutPriceChange() {
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

	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	bucketName := s.createBucket(sp, gvg, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
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
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(bucketName, objectName, payloadSize, 0)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)

	distBucketName := bucketName
	distObjectName := storagetestutils.GenRandomObjectName()

	objectIfo, err := s.copyObject(user, sp, bucketName, objectName, distBucketName, distObjectName)
	s.Require().NoError(err)
	s.sealObject(sp, gvg, distBucketName, distObjectName, objectIfo.Id, objectIfo.Checksums)
	// assertions
	streamRecordsAfterCopy := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfterCopy.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfterCopy.User.LockBalance, sdkmath.ZeroInt())
	// gvgFamilyRate1, gvgRate1, taxRate1, userTotalRate1 := s.calculateStorageRates(sp,distBucketName, distObjectName, payloadSize)
	s.Require().Equal(streamRecordsAfterCopy.User.NetflowRate.Sub(streamRecordsAfter.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfterCopy.GVGFamily.NetflowRate.Sub(streamRecordsAfter.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfterCopy.GVG.NetflowRate.Sub(streamRecordsAfter.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfterCopy.Tax.NetflowRate.Sub(streamRecordsAfter.Tax.NetflowRate), taxRate)
}

// TestStorageBill_CopyObject_WithoutPriceChange
func (s *PaymentTestSuite) TestStorageBill_CopyObject_WithPriceChange() {
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

	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	bucketName := s.createBucket(sp, gvg, user, 0)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
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
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(bucketName, objectName, payloadSize, 0)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)

	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(1000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	distBucketName := s.createBucket(sp, gvg, user, 0)
	distObjectName := storagetestutils.GenRandomObjectName()
	objectIfo, err := s.copyObject(user, sp, bucketName, objectName, distBucketName, distObjectName)
	s.Require().NoError(err)
	s.sealObject(sp, gvg, distBucketName, distObjectName, objectIfo.Id, objectIfo.Checksums)

	// assertions
	streamRecordsAfterCopy := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfterCopy.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfterCopy.User.LockBalance, sdkmath.ZeroInt())
	gvgFamilyRate1, gvgRate1, taxRate1, userTotalRate1 := s.calculateStorageRates(distBucketName, distObjectName, payloadSize, 0)
	s.Require().Equal(streamRecordsAfterCopy.GVGFamily.NetflowRate.Sub(streamRecordsAfter.GVGFamily.NetflowRate), gvgFamilyRate1)
	s.Require().Equal(streamRecordsAfterCopy.GVG.NetflowRate.Sub(streamRecordsAfter.GVG.NetflowRate), gvgRate1)
	s.Require().Equal(streamRecordsAfterCopy.Tax.NetflowRate.Sub(streamRecordsAfter.Tax.NetflowRate), taxRate1)
	s.Require().Equal(streamRecordsAfterCopy.User.NetflowRate.Sub(streamRecordsAfter.User.NetflowRate).BigInt().String(), userTotalRate1.Neg().BigInt().String())
}

// TestStorageBill_UpdateBucketQuota
func (s *PaymentTestSuite) TestStorageBill_UpdateBucketQuota() {
	var err error
	ctx := context.Background()
	sp := s.PickStorageProvider()
	user := s.GenAndChargeAccounts(1, 10)[0]

	gvg, found := sp.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	// case: create bucket with zero read quota
	bucketName := s.createBucket(sp, gvg, user, 0)

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

	readQuota := uint64(1024 * 1024 * 100)
	// case: update bucket read quota
	bucketInfo, err := s.updateBucket(user, bucketName, "", readQuota)
	s.Require().NoError(err)

	// check price and rate calculation
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: bucketInfo.CreateAt,
	})
	s.T().Logf("priceRes %s, err: %v", priceRes, err)
	s.Require().NoError(err)

	readPrice := priceRes.GlobalSpStorePrice.ReadPrice
	readChargeRate := readPrice.MulInt(sdk.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	taxRate := paymentParams.Params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
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

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	// case: update bucket read quota
	bucketInfo, err = s.updateBucket(user, bucketName, "", readQuota*2)
	s.Require().NoError(err)

	// check price and rate calculation
	priceRes, err = s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: bucketInfo.CreateAt,
	})
	s.T().Logf("priceRes %s, err: %v", priceRes, err)
	s.Require().NoError(err)

	readPrice = priceRes.GlobalSpStorePrice.ReadPrice
	readChargeRate = readPrice.MulInt(sdk.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	taxRate = paymentParams.Params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate = readChargeRate.Add(taxRate)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), readChargeRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)

	expectedOutFlows = []paymenttypes.OutFlow{
		{ToAddress: family.VirtualPaymentAddress, Rate: readChargeRate},
		{ToAddress: paymenttypes.ValidatorTaxPoolAddress.String(), Rate: taxRate},
	}
	userOutFlowsResponse, err = s.Client.OutFlows(ctx, &paymenttypes.QueryOutFlowsRequest{Account: user.GetAddr().String()})
	s.Require().NoError(err)
	sort.Slice(userOutFlowsResponse.OutFlows, func(i, j int) bool {
		return userOutFlowsResponse.OutFlows[i].ToAddress < userOutFlowsResponse.OutFlows[j].ToAddress
	})
	sort.Slice(expectedOutFlows, func(i, j int) bool {
		return expectedOutFlows[i].ToAddress < expectedOutFlows[j].ToAddress
	})
	s.Require().Equal(expectedOutFlows, userOutFlowsResponse.OutFlows)

	// set big read price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(1024*1024*1025), priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	chargedReadQuota := readQuota * 1024 * 1024
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &chargedReadQuota, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.reduceBNBBalance(user, s.Validator, sdkmath.NewIntWithDecimal(1, 16))
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "apply user flows list failed")
}

// TestStorageBill_UpdatePaymentAddress
func (s *PaymentTestSuite) TestStorageBill_UpdatePaymentAddress() {
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
	user := s.GenAndChargeAccounts(1, 100)[0]

	paymentAccountAddr := s.CreatePaymentAccount(user, 1, 17)
	paymentAcc := sdk.MustAccAddressFromHex(paymentAccountAddr)
	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	streamRecordsBefore := s.getStreamRecords(streamAddresses)

	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	// case: create bucket with zero read quota
	bucketName := s.createBucket(sp, gvg, user, 0)

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

	readQuota := uint64(1024 * 100)
	// case: update bucket read quota
	bucketInfo, err := s.updateBucket(user, bucketName, "", readQuota)
	s.Require().NoError(err)

	// check price and rate calculation
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: bucketInfo.CreateAt,
	})
	s.T().Logf("priceRes %s, err: %v", priceRes, err)
	s.Require().NoError(err)

	readPrice := priceRes.GlobalSpStorePrice.ReadPrice
	readChargeRate := readPrice.MulInt(sdk.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	taxRate := paymentParams.Params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate := readChargeRate.Add(taxRate)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), readChargeRate.Int64())
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

	// update new price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(100), priceRes.GlobalSpStorePrice.PrimaryStorePrice)
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	// case: update bucket paymentAccountAddr
	bucketInfo, err = s.updateBucket(user, bucketName, paymentAccountAddr, readQuota)
	s.Require().NoError(err)

	// check price and rate calculation
	priceRes, err = s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("priceRes %s, err: %v", priceRes, err)
	s.Require().NoError(err)

	readPrice = priceRes.GlobalSpStorePrice.ReadPrice
	readChargeRate = readPrice.MulInt(sdk.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	taxRate = paymentParams.Params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate = readChargeRate.Add(taxRate)

	// assertions
	streamAddresses[0] = paymentAccountAddr
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), readChargeRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)

	expectedOutFlows = []paymenttypes.OutFlow{
		{ToAddress: family.VirtualPaymentAddress, Rate: readChargeRate},
		{ToAddress: paymenttypes.ValidatorTaxPoolAddress.String(), Rate: taxRate},
	}
	userOutFlowsResponse, err = s.Client.OutFlows(ctx, &paymenttypes.QueryOutFlowsRequest{Account: paymentAccountAddr})
	s.Require().NoError(err)
	sort.Slice(userOutFlowsResponse.OutFlows, func(i, j int) bool {
		return userOutFlowsResponse.OutFlows[i].ToAddress < userOutFlowsResponse.OutFlows[j].ToAddress
	})
	sort.Slice(expectedOutFlows, func(i, j int) bool {
		return expectedOutFlows[i].ToAddress < expectedOutFlows[j].ToAddress
	})
	s.Require().Equal(expectedOutFlows, userOutFlowsResponse.OutFlows)

	// set big read price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(1024*1024*1024), priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	chargedReadQuota := readQuota * 1024 * 1024 * 1024 * 1024
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &chargedReadQuota, paymentAcc, storagetypes.VISIBILITY_TYPE_PRIVATE)
	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "apply user flows list failed")
	// new payment account balance not enough
	paymentAccountAddr = s.CreatePaymentAccount(user, 1, 13)
	paymentAcc = sdk.MustAccAddressFromHex(paymentAccountAddr)
	msgUpdateBucketInfo = storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &readQuota, paymentAcc, storagetypes.VISIBILITY_TYPE_PRIVATE)

	s.SendTxBlockWithExpectErrorString(msgUpdateBucketInfo, user, "apply user flows list failed")
}

func (s *PaymentTestSuite) TestStorageBill_MigrateBucket() {
	var err error
	ctx := context.Background()
	primarySP := s.PickStorageProvider()
	gvg, found := primarySP.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 10)[0]

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}

	streamAddresses0 := streamAddresses
	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	bucketName := s.createBucket(primarySP, gvg, user, 0)
	bucketInfo, err := s.Client.HeadBucket(context.Background(), &storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// case: seal object without price change
	s.sealObject(primarySP, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(bucketName, objectName, payloadSize, 0)
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)
	taxRate0 := taxRate

	dstPrimarySP := s.CreateNewStorageProvider()

	// update price
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(2), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	_, secondarySPIDs := s.GetSecondarySP(dstPrimarySP, primarySP)
	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(dstPrimarySP, 0, secondarySPIDs, 1)
	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &virtualgrouptypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	dstGVG := gvgResp.GlobalVirtualGroup
	s.Require().True(found)

	queryFamilyResponse, err = s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: dstGVG.FamilyId,
	})
	s.Require().NoError(err)
	family = queryFamilyResponse.GlobalVirtualGroupFamily
	streamAddresses = []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		dstGVG.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	fundAddress := primarySP.FundingKey.GetAddr()
	streamRecordsBefore = s.getStreamRecords(streamAddresses)

	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: fundAddress.String()}
	fundBalanceBefore, err := s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)

	// MigrationBucket
	msgMigrationBucket, msgCompleteMigrationBucket := s.NewMigrateBucket(primarySP, dstPrimarySP, user, bucketName, gvg.FamilyId, dstGVG.FamilyId, bucketInfo.BucketInfo.Id)
	s.SendTxBlock(user, msgMigrationBucket)
	s.Require().NoError(err)

	// complete MigrationBucket
	s.SendTxBlock(dstPrimarySP.OperatorKey, msgCompleteMigrationBucket)
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	fundBalanceAfter, err := s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("fundBalanceBefore: %v, fundBalanceAfter: %v, diff: %v", fundBalanceBefore, fundBalanceAfter, fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount))
	s.Require().True(fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount).GT(sdkmath.NewInt(0)), "migrate sp fund address need settle")
	gvgFamilyRate, gvgRate, taxRate, userTotalRate = s.calculateStorageRates(bucketName, objectName, payloadSize, time.Now().Unix())
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.T().Logf("NetflowRate: %v, userTotalRate: %v, actual taxRate diff: %v, expect taxRate diff: %v", streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Neg(), streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	// tax rate diff
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Abs())

	// set price
	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(120), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(5000))

	queryBalanceRequest.Address = dstPrimarySP.FundingKey.GetAddr().String()
	fundBalanceBefore, err = s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)
	streamRecordsBefore = s.getStreamRecords(streamAddresses0)
	// send msgMigrationBucket
	msgMigrationBucket, msgCompleteMigrationBucket = s.NewMigrateBucket(dstPrimarySP, primarySP, user, bucketName, dstGVG.FamilyId, gvg.FamilyId, bucketInfo.BucketInfo.Id)

	s.SendTxBlock(user, msgMigrationBucket)
	s.Require().NoError(err)
	s.reduceBNBBalance(user, s.Validator, sdkmath.NewIntWithDecimal(1, 1))

	s.SendTxBlockWithExpectErrorString(msgCompleteMigrationBucket, primarySP.OperatorKey, "apply stream record changes for user failed")

	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice.MulInt64(10), priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10))
	readPrice, primaryPrice, secondaryPrice := s.getPrices(time.Now().Unix())
	s.T().Logf("readPrice: %v, primaryPrice: %v,secondaryPrice: %v", readPrice, primaryPrice, secondaryPrice)

	s.transferBNB(s.Validator, user, sdkmath.NewIntWithDecimal(10000, 18))

	s.SendTxBlock(primarySP.OperatorKey, msgCompleteMigrationBucket)
	streamRecordsAfter = s.getStreamRecords(streamAddresses0)
	fundBalanceAfter, err = s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("fundBalanceBefore: %v, fundBalanceAfter: %v, diff: %v", fundBalanceBefore, fundBalanceAfter, fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount))
	s.Require().True(fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount).GT(sdkmath.NewInt(0)), "migrate sp fund address need settle")
	taxRate1 := taxRate
	gvgFamilyRate, gvgRate, taxRate, userTotalRate = s.calculateStorageRates(bucketName, objectName, payloadSize, time.Now().Unix())
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.T().Logf("NetflowRate: %v, userTotalRate: %v, actual taxRate diff: %v, expect taxRate diff: %v", streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Neg(), streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	// tax rate diff
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate1))
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Abs())
}

func (s *PaymentTestSuite) TestStorageBill_MigrateBucket_LockedFee_ThenDiscontinueBucket() {
	var err error
	ctx := context.Background()
	primarySP := s.PickStorageProvider()
	gvg, found := primarySP.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	family := queryFamilyResponse.GlobalVirtualGroupFamily
	user := s.GenAndChargeAccounts(1, 10)[0]

	streamAddresses := []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		gvg.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}

	paymentParams, err := s.Client.PaymentQueryClient.Params(ctx, &paymenttypes.QueryParamsRequest{})
	s.T().Logf("paymentParams %s, err: %v", paymentParams, err)
	s.Require().NoError(err)

	bucketName := s.createBucket(primarySP, gvg, user, 0)
	bucketInfo, err := s.Client.HeadBucket(context.Background(), &storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)

	// create object with none zero payload size
	streamRecordsBefore := s.getStreamRecords(streamAddresses)
	_, _, objectName, objectId, checksums, payloadSize := s.createObject(user, bucketName, false)

	// assertions
	streamRecordsAfter := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	lockFee := s.calculateLockFee(bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// case: seal object
	s.sealObject(primarySP, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(bucketName, objectName, payloadSize, 0)
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)
	taxRate0 := taxRate
	dstPrimarySP := s.CreateNewStorageProvider()

	// create a new object without seal
	s.createObject(user, bucketName, false)

	// update price after lock
	priceRes, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log("price", priceRes.GlobalSpStorePrice)

	s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice.MulInt64(10000))
	defer s.updateGlobalSpPrice(priceRes.GlobalSpStorePrice.ReadPrice, priceRes.GlobalSpStorePrice.PrimaryStorePrice)

	_, secondarySPIDs := s.GetSecondarySP(dstPrimarySP, primarySP)
	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(dstPrimarySP, 0, secondarySPIDs, 1)
	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &virtualgrouptypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	dstGVG := gvgResp.GlobalVirtualGroup
	s.Require().True(found)

	queryFamilyResponse, err = s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: dstGVG.FamilyId,
	})
	s.Require().NoError(err)
	family = queryFamilyResponse.GlobalVirtualGroupFamily
	streamAddresses = []string{
		user.GetAddr().String(),
		family.VirtualPaymentAddress,
		dstGVG.VirtualPaymentAddress,
		paymenttypes.ValidatorTaxPoolAddress.String(),
	}
	fundAddress := primarySP.FundingKey.GetAddr()
	streamRecordsBefore = s.getStreamRecords(streamAddresses)

	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: fundAddress.String()}
	fundBalanceBefore, err := s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)

	// MigrateBucket
	msgMigrateBucket, msgCompleteMigrateBucket := s.NewMigrateBucket(primarySP, dstPrimarySP, user, bucketName, gvg.FamilyId, dstGVG.FamilyId, bucketInfo.BucketInfo.Id)
	s.SendTxBlock(user, msgMigrateBucket)
	s.Require().NoError(err)

	// complete MigrateBucket
	s.SendTxBlock(dstPrimarySP.OperatorKey, msgCompleteMigrateBucket)
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	fundBalanceAfter, err := s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("fundBalanceBefore: %v, fundBalanceAfter: %v, diff: %v", fundBalanceBefore, fundBalanceAfter, fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount))
	s.Require().True(fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount).GT(sdkmath.NewInt(0)), "migrate sp fund address need settle")
	gvgFamilyRate, gvgRate, taxRate, userTotalRate = s.calculateStorageRates(bucketName, objectName, payloadSize, time.Now().Unix())
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.T().Logf("NetflowRate: %v, userTotalRate: %v, actual taxRate diff: %v, expect taxRate diff: %v", streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Neg(), streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	// tax rate diff
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Abs())

	// force delete bucket
	headBucketResp, _ := s.Client.HeadBucket(ctx, &storagetypes.QueryHeadBucketRequest{BucketName: bucketName})
	s.T().Log("headBucketResp", core.YamlString(headBucketResp))
	msgDiscontinueBucket := storagetypes.NewMsgDiscontinueBucket(dstPrimarySP.GcKey.GetAddr(), bucketName, "test")
	txRes := s.SendTxBlock(dstPrimarySP.GcKey, msgDiscontinueBucket)
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
}

func (s *PaymentTestSuite) TestStorageBill_MigrateBucket_FrozenAccount_NotAllowed() {
	var err error
	ctx := context.Background()
	primarySP := s.PickStorageProvider()
	gvg, found := primarySP.GetFirstGlobalVirtualGroup()
	s.Require().True(found)
	queryFamilyResponse, err := s.Client.GlobalVirtualGroupFamily(ctx, &virtualgrouptypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: gvg.FamilyId,
	})
	s.Require().NoError(err)
	s.T().Log("queryFamilyResponse", core.YamlString(queryFamilyResponse))
	user := s.GenAndChargeAccounts(1, 10)[0]

	params := s.queryParams()
	reserveTime := params.VersionedParams.ReserveTime
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.GlobalSpStorePrice.ReadPrice
	bucketChargedReadQuota := uint64(1000)
	readRate := readPrice.MulInt(sdkmath.NewIntFromUint64(bucketChargedReadQuota)).TruncateInt()
	readTaxRate := params.VersionedParams.ValidatorTaxRate.MulInt(readRate).TruncateInt()
	readTotalRate := readRate.Add(readTaxRate)
	paymentAccountBNBNeeded := readTotalRate.Mul(sdkmath.NewIntFromUint64(reserveTime))

	// create payment account and deposit
	msgCreatePaymentAccount := &paymenttypes.MsgCreatePaymentAccount{
		Creator: user.GetAddr().String(),
	}
	_ = s.SendTxBlock(user, msgCreatePaymentAccount)
	paymentAccountsReq := &paymenttypes.QueryPaymentAccountsByOwnerRequest{Owner: user.GetAddr().String()}
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

	bucketName := "ch" + storagetestutils.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, primarySP.OperatorKey.GetAddr(),
		sdk.MustAccAddressFromHex(paymentAddr), math.MaxUint, nil, bucketChargedReadQuota)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = primarySP.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateBucket)

	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	_, err = s.Client.HeadBucket(context.Background(), &queryHeadBucketRequest)
	s.Require().NoError(err)
	bucketInfo, err := s.Client.HeadBucket(context.Background(), &storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)

	// wait until settle time
	paymentAccountStreamRecord := s.getStreamRecord(paymentAddr)
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
	paymentAccountStreamRecordAfterAutoSettle := s.getStreamRecord(paymentAddr)
	s.T().Logf("paymentAccountStreamRecordAfterAutoSettle %s", core.YamlString(paymentAccountStreamRecordAfterAutoSettle))
	s.Require().Equal(paymentAccountStreamRecordAfterAutoSettle.Status, paymenttypes.STREAM_ACCOUNT_STATUS_FROZEN)
	s.Require().Equal(paymentAccountStreamRecordAfterAutoSettle.NetflowRate.Int64(), int64(0))
	s.Require().Equal(paymentAccountStreamRecordAfterAutoSettle.FrozenNetflowRate.Int64(), readTotalRate.Neg().Int64())

	dstPrimarySP := s.CreateNewStorageProvider()
	_, secondarySPIDs := s.GetSecondarySP(dstPrimarySP, primarySP)
	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(dstPrimarySP, 0, secondarySPIDs, 1)
	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &virtualgrouptypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	dstGVG := gvgResp.GlobalVirtualGroup
	s.Require().True(found)

	// MigrateBucket
	msgMigrateBucket, _ := s.NewMigrateBucket(primarySP, dstPrimarySP, user, bucketName, gvg.FamilyId, dstGVG.FamilyId, bucketInfo.BucketInfo.Id)
	s.SendTxBlockWithExpectErrorString(msgMigrateBucket, user, "frozen")
}

func (s *PaymentTestSuite) GetSecondarySP(sps ...*core.StorageProvider) ([]*core.StorageProvider, []uint32) {
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

func (s *PaymentTestSuite) NewMigrateBucket(srcSP, dstSP *core.StorageProvider, user keys.KeyManager, bucketName string, srcID, dstID uint32, bucketID sdkmath.Uint) (*storagetypes.MsgMigrateBucket, *storagetypes.MsgCompleteMigrateBucket) {
	secondarySPs, _ := s.GetSecondarySP(srcSP, dstSP)
	var gvgMappings []*storagetypes.GVGMapping
	gvgMappings = append(gvgMappings, &storagetypes.GVGMapping{SrcGlobalVirtualGroupId: srcID, DstGlobalVirtualGroupId: dstID})
	for _, gvgMapping := range gvgMappings {
		migrationBucketSignHash := storagetypes.NewSecondarySpMigrationBucketSignDoc(s.GetChainID(), bucketID, dstSP.Info.Id, gvgMapping.SrcGlobalVirtualGroupId, gvgMapping.DstGlobalVirtualGroupId).GetBlsSignHash()
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

	msgMigrationBucket := storagetypes.NewMsgMigrateBucket(user.GetAddr(), bucketName, dstSP.Info.Id)
	msgMigrationBucket.DstPrimarySpApproval.ExpiredHeight = math.MaxInt
	msgMigrationBucket.DstPrimarySpApproval.Sig, _ = dstSP.ApprovalKey.Sign(msgMigrationBucket.GetApprovalBytes())

	msgCompleteMigrationBucket := storagetypes.NewMsgCompleteMigrateBucket(dstSP.OperatorKey.GetAddr(), bucketName, dstID, gvgMappings)

	return msgMigrationBucket, msgCompleteMigrationBucket
}

// CreatePaymentAccount create new payment account and return latest payment account
func (s *PaymentTestSuite) CreatePaymentAccount(user keys.KeyManager, amount, decimal int64) string {
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
	paymentAccountAddr := paymentAccounts.PaymentAccounts[len(paymentAccounts.PaymentAccounts)-1]
	// charge payment account
	paymentAcc := sdk.MustAccAddressFromHex(paymentAccountAddr)
	msgDeposit := &paymenttypes.MsgDeposit{
		Creator: user.GetAddr().String(),
		To:      paymentAcc.String(),
		Amount:  types.NewIntFromInt64WithDecimal(amount, decimal), // deposit more than needed
	}
	s.SendTxBlock(user, msgDeposit)

	return paymentAccountAddr
}

func (s *PaymentTestSuite) copyObject(user keys.KeyManager, sp *core.StorageProvider, bucketName, objectName, dstBucketName, dstObjectName string) (*storagetypes.ObjectInfo, error) {
	msgCopyObject := storagetypes.NewMsgCopyObject(user.GetAddr(), bucketName, dstBucketName, objectName, dstObjectName, math.MaxUint, nil)
	msgCopyObject.DstPrimarySpApproval.Sig, _ = sp.ApprovalKey.Sign(msgCopyObject.GetApprovalBytes())

	s.SendTxBlock(user, msgCopyObject)
	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: dstBucketName,
		ObjectName: dstObjectName,
	}
	headObjectResponse, err := s.Client.HeadObject(context.Background(), &queryHeadObjectRequest)
	return headObjectResponse.ObjectInfo, err
}

func (s *PaymentTestSuite) updateBucket(user keys.KeyManager, bucketName, paymentAddress string, chargedReadQuota uint64) (*storagetypes.BucketInfo, error) {
	msgUpdateBucketInfo := storagetypes.NewMsgUpdateBucketInfo(
		user.GetAddr(), bucketName, &chargedReadQuota, user.GetAddr(), storagetypes.VISIBILITY_TYPE_PRIVATE)
	if paymentAddress != "" {
		msgUpdateBucketInfo.PaymentAddress = paymentAddress
	}
	s.SendTxBlock(user, msgUpdateBucketInfo)

	queryHeadObjectRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	headObjectResponse, err := s.Client.HeadBucket(context.Background(), &queryHeadObjectRequest)
	return headObjectResponse.BucketInfo, err
}

func (s *PaymentTestSuite) reduceBNBBalance(user, to keys.KeyManager, leftBalance sdkmath.Int) {
	queryBalanceRequest := banktypes.QueryBalanceRequest{Denom: s.Config.Denom, Address: user.GetAddr().String()}
	queryBalanceResponse, err := s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)

	msgSend := banktypes.NewMsgSend(user.GetAddr(), to.GetAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, queryBalanceResponse.Balance.Amount.SubRaw(5*types.DecimalGwei)),
	))

	simulateResponse := s.SimulateTx(msgSend, user)
	gasLimit := simulateResponse.GasInfo.GetGasUsed()
	gasPrice, err := sdk.ParseCoinNormalized(simulateResponse.GasInfo.GetMinGasPrice())
	s.Require().NoError(err)
	sendBNBAmount := queryBalanceResponse.Balance.Amount.Sub(gasPrice.Amount.Mul(sdk.NewInt(int64(gasLimit)))).Sub(leftBalance)
	msgSend.Amount = sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, sendBNBAmount),
	)
	s.SendTxBlock(user, msgSend)
	queryBalanceResponse, err = s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("balance: %v", queryBalanceResponse.Balance.Amount)
}

func (s *PaymentTestSuite) transferBNB(from, to keys.KeyManager, amount sdkmath.Int) {
	msgSend := banktypes.NewMsgSend(from.GetAddr(), to.GetAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, amount),
	))
	s.SendTxBlock(from, msgSend)
}
