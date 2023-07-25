package tests

import (
	"context"
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

	distBucketName := bucketName
	distObjectName := storagetestutils.GenRandomObjectName()

	objectIfo, err := s.copyObject(user, sp, bucketName, objectName, distBucketName, distObjectName)
	s.Require().NoError(err)
	s.sealObject(sp, gvg, distBucketName, distObjectName, objectIfo.Id, objectIfo.Checksums)
	// assertions
	streamRecordsAfterCopy := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfterCopy.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfterCopy.User.LockBalance, sdkmath.ZeroInt())
	//gvgFamilyRate1, gvgRate1, taxRate1, userTotalRate1 := s.calculateStorageRates(sp,distBucketName, distObjectName, payloadSize)
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

	priceRes, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: 0,
	})
	s.Require().NoError(err)
	// update new price
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     priceRes.SpStoragePrice.ReadPrice,
		FreeReadQuota: priceRes.SpStoragePrice.FreeReadQuota,
		StorePrice:    priceRes.SpStoragePrice.StorePrice.MulInt64(1000),
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	distBucketName := s.createBucket(sp, user, 0)
	distObjectName := storagetestutils.GenRandomObjectName()
	objectIfo, err := s.copyObject(user, sp, bucketName, objectName, distBucketName, distObjectName)
	s.Require().NoError(err)
	s.sealObject(sp, gvg, distBucketName, distObjectName, objectIfo.Id, objectIfo.Checksums)
	// assertions
	streamRecordsAfterCopy := s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfterCopy.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfterCopy.User.LockBalance, sdkmath.ZeroInt())
	gvgFamilyRate1, gvgRate1, taxRate1, userTotalRate1 := s.calculateStorageRates(sp, distBucketName, distObjectName, payloadSize, 0)
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
	// recover price
	defer s.SetSPPrice(sp, "12.34", "0")
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

	readQuota := uint64(1024 * 1024 * 100)
	// case: update bucket read quota
	bucketInfo, err := s.updateBucket(user, bucketName, "", readQuota)
	s.Require().NoError(err)

	// check price and rate calculation
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: time.Now().Unix(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
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
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice.MulInt64(100),
		FreeReadQuota: queryGetSpStoragePriceByTimeResp.SpStoragePrice.FreeReadQuota,
		StorePrice:    queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// case: update bucket read quota
	bucketInfo, err = s.updateBucket(user, bucketName, "", readQuota*2)
	s.Require().NoError(err)

	// check price and rate calculation
	queryGetSpStoragePriceByTimeResp, err = s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: time.Now().Unix(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice = queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
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
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice.MulInt64(1024 * 1024 * 1024),
		FreeReadQuota: queryGetSpStoragePriceByTimeResp.SpStoragePrice.FreeReadQuota,
		StorePrice:    queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

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
	defer s.SetSPPrice(sp, "12.34", "0")
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

	readQuota := uint64(1024 * 100)
	// case: update bucket read quota
	bucketInfo, err := s.updateBucket(user, bucketName, "", readQuota)
	s.Require().NoError(err)

	// check price and rate calculation
	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: time.Now().Unix(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
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
	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice.MulInt64(100),
		FreeReadQuota: queryGetSpStoragePriceByTimeResp.SpStoragePrice.FreeReadQuota,
		StorePrice:    queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

	// case: update bucket paymentAccountAddr
	bucketInfo, err = s.updateBucket(user, bucketName, paymentAccountAddr, readQuota)
	s.Require().NoError(err)

	// check price and rate calculation
	queryGetSpStoragePriceByTimeResp, err = s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: time.Now().Unix(),
	})
	s.T().Logf("queryGetSpStoragePriceByTimeResp %s, err: %v", queryGetSpStoragePriceByTimeResp, err)
	s.Require().NoError(err)

	readPrice = queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice
	readChargeRate = readPrice.MulInt(sdk.NewIntFromUint64(bucketInfo.ChargedReadQuota)).TruncateInt()
	s.T().Logf("readPrice: %s, readChargeRate: %s", readPrice, readChargeRate)
	taxRate = paymentParams.Params.VersionedParams.ValidatorTaxRate.MulInt(readChargeRate).TruncateInt()
	userTotalRate = readChargeRate.Add(taxRate)

	// assertions
	streamAddresses[0] = paymentAccountAddr
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
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
	msgUpdatePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice.MulInt64(1024 * 1024 * 1024),
		FreeReadQuota: queryGetSpStoragePriceByTimeResp.SpStoragePrice.FreeReadQuota,
		StorePrice:    queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)

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

func (s *PaymentTestSuite) TestStorageBill_MigrationBucket() {
	var err error
	ctx := context.Background()
	primarySP := s.PickStorageProvider()
	s.SetSPPrice(primarySP, "1", "1.15")

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

	bucketName := s.createBucket(primarySP, user, 0)
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
	lockFee := s.calculateLockFee(primarySP, bucketName, objectName, payloadSize)
	s.Require().Equal(streamRecordsAfter.User.LockBalance.Sub(streamRecordsBefore.User.LockBalance), lockFee)
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate).Int64(), int64(0))
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate).Int64(), int64(0))

	// case: seal object without price change
	s.sealObject(primarySP, gvg, bucketName, objectName, objectId, checksums)

	// assertions
	streamRecordsAfter = s.getStreamRecords(streamAddresses)
	gvgFamilyRate, gvgRate, taxRate, userTotalRate := s.calculateStorageRates(primarySP, bucketName, objectName, payloadSize, 0)
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Sub(streamRecordsBefore.User.NetflowRate), userTotalRate.Neg())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate)
	taxRate0 := taxRate
	dstPrimarySP := s.CreateNewStorageProvider()

	s.SetSPPrice(dstPrimarySP, "2", "1.45")
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
	gvgFamilyRate, gvgRate, taxRate, userTotalRate = s.calculateStorageRates(dstPrimarySP, bucketName, objectName, payloadSize, time.Now().Unix())
	s.T().Logf("gvgFamilyRate: %v, gvgRate: %v, taxRate: %v, userTotalRate: %v", gvgFamilyRate, gvgRate, taxRate, userTotalRate)
	s.T().Logf("NetflowRate: %v, userTotalRate: %v, actual taxRate diff: %v, expect taxRate diff: %v", streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Neg(), streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))

	s.Require().Equal(streamRecordsAfter.User.StaticBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.User.LockBalance, sdkmath.ZeroInt())
	s.Require().Equal(streamRecordsAfter.GVGFamily.NetflowRate.Sub(streamRecordsBefore.GVGFamily.NetflowRate), gvgFamilyRate)
	s.Require().Equal(streamRecordsAfter.GVG.NetflowRate.Sub(streamRecordsBefore.GVG.NetflowRate), gvgRate)
	// tax rate diff
	s.Require().Equal(streamRecordsAfter.Tax.NetflowRate.Sub(streamRecordsBefore.Tax.NetflowRate), taxRate.Sub(taxRate0))
	s.Require().Equal(streamRecordsAfter.User.NetflowRate.Neg(), userTotalRate.Abs())

	s.SetSPPrice(primarySP, "12.3", "100")

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

	s.SetSPPrice(primarySP, "12.3", "13")
	readPrice, primaryPrice, secondaryPrice := s.getPrices(primarySP, time.Now().Unix())
	s.T().Logf("readPrice: %v, primaryPrice: %v,secondaryPrice: %v", readPrice, primaryPrice, secondaryPrice)

	s.transferBNB(s.Validator, user, sdkmath.NewIntWithDecimal(10000, 18))

	s.SendTxBlock(primarySP.OperatorKey, msgCompleteMigrationBucket)
	streamRecordsAfter = s.getStreamRecords(streamAddresses0)
	fundBalanceAfter, err = s.Client.BankQueryClient.Balance(context.Background(), &queryBalanceRequest)
	s.Require().NoError(err)
	s.T().Logf("fundBalanceBefore: %v, fundBalanceAfter: %v, diff: %v", fundBalanceBefore, fundBalanceAfter, fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount))
	s.Require().True(fundBalanceAfter.Balance.Amount.Sub(fundBalanceBefore.Balance.Amount).GT(sdkmath.NewInt(0)), "migrate sp fund address need settle")
	taxRate1 := taxRate
	gvgFamilyRate, gvgRate, taxRate, userTotalRate = s.calculateStorageRates(primarySP, bucketName, objectName, payloadSize, time.Now().Unix())
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
func (s *PaymentTestSuite) SetSPPrice(sp *core.StorageProvider, readPrice, storePrice string) {
	ctx := context.Background()

	queryGetSpStoragePriceByTimeResp, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: time.Now().Unix(),
	})
	s.Require().NoError(err)
	ReadPrice, err := sdk.NewDecFromStr(readPrice)
	s.Require().NoError(err)
	StorePrice := queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice
	if storePrice != "0" {
		StorePrice, err = sdk.NewDecFromStr(storePrice)
		s.Require().NoError(err)
	}

	msgUpdatePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     sp.OperatorKey.GetAddr().String(),
		ReadPrice:     ReadPrice,
		FreeReadQuota: queryGetSpStoragePriceByTimeResp.SpStoragePrice.FreeReadQuota,
		StorePrice:    StorePrice,
	}
	s.SendTxBlock(sp.OperatorKey, msgUpdatePrice)
	queryGetSpStoragePriceByTimeResp, err = s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    sp.OperatorKey.GetAddr().String(),
		Timestamp: time.Now().Unix(),
	})
	s.Require().NoError(err)
	s.T().Logf("queryGetSpStoragePriceByTimeResp read price: %s, store price: %v",
		queryGetSpStoragePriceByTimeResp.SpStoragePrice.ReadPrice, queryGetSpStoragePriceByTimeResp.SpStoragePrice.StorePrice)

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
	queryGetPaymentAccountsByOwnerRequest := paymenttypes.QueryGetPaymentAccountsByOwnerRequest{
		Owner: user.GetAddr().String(),
	}
	paymentAccounts, err := s.Client.GetPaymentAccountsByOwner(ctx, &queryGetPaymentAccountsByOwnerRequest)
	s.Require().NoError(err)
	s.T().Log(paymentAccounts)
	paymentAccountAddr := paymentAccounts.PaymentAccounts[len(paymentAccounts.PaymentAccounts)-1]
	// charge payment account
	paymentAcc := sdk.MustAccAddressFromHex(paymentAccountAddr)
	msgSend := banktypes.NewMsgSend(user.GetAddr(), paymentAcc, []sdk.Coin{{Denom: "BNB", Amount: types.NewIntFromInt64WithDecimal(amount, decimal)}})
	s.SendTxBlock(user, msgSend)

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

func (s *PaymentTestSuite) updateBucket(user keys.KeyManager, bucketName string, paymentAddress string, chargedReadQuota uint64) (*storagetypes.BucketInfo, error) {

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

func (s *PaymentTestSuite) transferBNB(user, to keys.KeyManager, amount sdkmath.Int) {

	msgSend := banktypes.NewMsgSend(user.GetAddr(), to.GetAddr(), sdk.NewCoins(
		sdk.NewCoin(s.Config.Denom, amount),
	))
	s.SendTxBlock(user, msgSend)

}
