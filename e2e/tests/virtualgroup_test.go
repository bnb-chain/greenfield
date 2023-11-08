package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	storagetestutil "github.com/bnb-chain/greenfield/testutil/storage"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type VirtualGroupTestSuite struct {
	core.BaseSuite
}

func (s *VirtualGroupTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *VirtualGroupTestSuite) SetupTest() {
}

func TestVirtualGroupTestSuite(t *testing.T) {
	suite.Run(t, new(VirtualGroupTestSuite))
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroup(gvgID uint32) *virtualgroupmoduletypes.GlobalVirtualGroup {
	resp, err := s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroup
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroupByFamily(familyID uint32) []*virtualgroupmoduletypes.GlobalVirtualGroup {
	s.T().Logf("familyID: %d", familyID)
	resp, err := s.Client.GlobalVirtualGroupByFamilyID(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupByFamilyIDRequest{
			GlobalVirtualGroupFamilyId: familyID,
		})
	s.Require().NoError(err)
	_, err = s.Client.VirtualGroupQueryClient.Params(context.Background(), &virtualgroupmoduletypes.QueryParamsRequest{})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroups
}

func (s *VirtualGroupTestSuite) queryAvailableGlobalVirtualGroupFamilies(familyIds []uint32) []uint32 {
	resp, err := s.Client.AvailableGlobalVirtualGroupFamilies(
		context.Background(),
		&virtualgroupmoduletypes.AvailableGlobalVirtualGroupFamiliesRequest{
			GlobalVirtualGroupFamilyIds: familyIds,
		})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroupFamilyIds
}

func (s *VirtualGroupTestSuite) TestBasic() {
	primarySP := s.BaseSuite.PickStorageProvider()

	s.BaseSuite.RefreshGVGFamilies()
	s.Require().Greater(len(primarySP.GlobalVirtualGroupFamilies), 0)

	var gvgs []*virtualgroupmoduletypes.GlobalVirtualGroup
	var gvg *virtualgroupmoduletypes.GlobalVirtualGroup
	for _, g := range primarySP.GlobalVirtualGroupFamilies {
		gvgs = g
	}
	s.Require().Greater(len(gvgs), 0)

	for _, g := range gvgs {
		gvg = g
	}

	availableGvgFamilyIds := s.queryAvailableGlobalVirtualGroupFamilies([]uint32{gvg.FamilyId})
	s.Require().Equal(availableGvgFamilyIds[0], gvg.FamilyId)

	srcGVGs := s.queryGlobalVirtualGroupByFamily(gvg.FamilyId)

	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySP.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
	}
	s.BaseSuite.CreateGlobalVirtualGroup(primarySP, gvg.FamilyId, secondarySPIDs, 1)

	gvgs = s.queryGlobalVirtualGroupByFamily(gvg.FamilyId)
	s.Require().Equal(len(gvgs), len(srcGVGs)+1)

	oldGVGIDs := make(map[uint32]bool)
	for _, gvg := range srcGVGs {
		oldGVGIDs[gvg.Id] = true
	}
	var newGVG *virtualgroupmoduletypes.GlobalVirtualGroup

	for _, gvg := range gvgs {
		if !oldGVGIDs[gvg.Id] {
			newGVG = gvg
			break
		}
	}

	s.Require().Equal(newGVG.TotalDeposit.Int64(), int64(1000000000000000000))

	// test deposit
	msgDeposit := virtualgroupmoduletypes.MsgDeposit{
		StorageProvider:      primarySP.FundingKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
		Deposit:              sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(1, types.DecimalBNB)),
	}
	s.SendTxBlock(primarySP.FundingKey, &msgDeposit)

	gvgAfterDeposit := s.queryGlobalVirtualGroup(newGVG.Id)
	s.Require().Equal(gvgAfterDeposit.TotalDeposit.Int64(), int64(2000000000000000000))

	// test withdraw
	balance, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySP.FundingKey.GetAddr().String(),
	})
	s.Require().NoError(err)

	msgWithdraw := virtualgroupmoduletypes.MsgWithdraw{
		StorageProvider:      primarySP.FundingKey.GetAddr().String(),
		Withdraw:             sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(1, types.DecimalBNB)),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.FundingKey, &msgWithdraw)
	balanceAfterWithdraw, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySP.FundingKey.GetAddr().String(),
	})
	s.Require().NoError(err)

	s.T().Logf("balance: %s, after: %s", balance.String(), balanceAfterWithdraw.String())
	s.Require().Equal(balanceAfterWithdraw.Balance.Amount.Sub(balance.Balance.Amount).Int64(), int64(999994000000000000))

	// test delete gvg
	msgDeleteGVG := virtualgroupmoduletypes.MsgDeleteGlobalVirtualGroup{
		StorageProvider:      primarySP.OperatorKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.OperatorKey, &msgDeleteGVG)

	newGVGs := s.queryGlobalVirtualGroupByFamily(newGVG.FamilyId)

	for _, gvg := range newGVGs {
		if gvg.Id == newGVG.Id {
			s.Assert().True(false)
		}
	}
	_, err = s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: newGVG.Id})
	s.Require().Error(err)
}

func (s *VirtualGroupTestSuite) TestSettle() {
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	_, _, primarySp, secondarySps, gvgFamilyId, gvgId := s.createObject()
	s.T().Log("gvg family", gvgFamilyId, "gvg", gvgId)

	queryFamilyResp, err := s.Client.GlobalVirtualGroupFamily(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{
			FamilyId: gvgFamilyId,
		})
	s.Require().NoError(err)
	gvgFamily := queryFamilyResp.GlobalVirtualGroupFamily

	queryGVGResp, err := s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{
			GlobalVirtualGroupId: gvgId,
		})
	s.Require().NoError(err)
	secondarySpIds := make(map[uint32]struct{})
	for _, id := range queryGVGResp.GlobalVirtualGroup.SecondarySpIds {
		secondarySpIds[id] = struct{}{}
	}

	secondarySpAddrs := make([]string, 0)
	for _, secondarySp := range secondarySps {
		if _, ok := secondarySpIds[secondarySp.Info.Id]; ok {
			secondarySpAddrs = append(secondarySpAddrs, secondarySp.FundingKey.GetAddr().String())
		}
	}

	// sleep seconds
	time.Sleep(3 * time.Second)

	primaryBalance, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySp.FundingKey.GetAddr().String(),
	})
	s.Require().NoError(err)
	secondaryBalances := make([]sdkmath.Int, 0)
	for _, addr := range secondarySpAddrs {
		tempResp, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
			Denom: s.Config.Denom, Address: addr,
		})
		s.Require().NoError(err)
		secondaryBalances = append(secondaryBalances, tempResp.Balance.Amount)
	}

	// settle gvg family
	msgSettle := virtualgroupmoduletypes.MsgSettle{
		StorageProvider:            user.GetAddr().String(),
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
	}
	s.SendTxBlock(user, &msgSettle)

	primaryBalanceAfter, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySp.FundingKey.GetAddr().String(),
	})
	s.Require().NoError(err)

	s.T().Logf("primaryBalance: %s, after: %s", primaryBalance.String(), primaryBalanceAfter.String())
	s.Require().True(primaryBalanceAfter.Balance.Amount.GT(primaryBalance.Balance.Amount))

	// settle gvg
	msgSettle = virtualgroupmoduletypes.MsgSettle{
		StorageProvider:            user.GetAddr().String(),
		GlobalVirtualGroupFamilyId: 0,
		GlobalVirtualGroupIds:      []uint32{gvgId},
	}
	s.SendTxBlock(user, &msgSettle)

	secondaryBalancesAfter := make([]sdkmath.Int, 0, len(secondaryBalances))
	for _, addr := range secondarySpAddrs {
		tempResp, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
			Denom: s.Config.Denom, Address: addr,
		})
		s.Require().NoError(err)
		secondaryBalancesAfter = append(secondaryBalancesAfter, tempResp.Balance.Amount)
	}

	for i := range secondaryBalances {
		s.T().Logf("secondaryBalance: %s, after: %s", secondaryBalances[i].String(), secondaryBalancesAfter[i].String())
		s.Require().True(secondaryBalancesAfter[i].GT(secondaryBalances[i]))
	}
}

func (s *VirtualGroupTestSuite) createObject() (string, string, *core.StorageProvider, []*core.StorageProvider, uint32, uint32) {
	var err error
	sp := s.BaseSuite.PickStorageProvider()
	secondarySps := make([]*core.StorageProvider, 0)
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

	// HeadBucket
	ctx := context.Background()
	queryHeadBucketRequest := storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	}
	queryHeadBucketResponse, err := s.Client.HeadBucket(ctx, &queryHeadBucketRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadBucketResponse.BucketInfo.BucketName, bucketName)

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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), storagetypes.VISIBILITY_TYPE_PRIVATE, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(user, msgCreateObject)

	// HeadObject
	queryHeadObjectRequest := storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	}
	queryHeadObjectResponse, err := s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)

	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)

	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for _, spID := range gvg.SecondarySpIds {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[spID], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[spID].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
		if s.StorageProviders[spID].Info.Id != sp.Info.Id {
			secondarySps = append(secondarySps, s.StorageProviders[spID])
		}
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.SendTxBlock(sp.SealKey, msgSealObject)

	queryHeadObjectResponse, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	return bucketName, objectName, sp, secondarySps, gvg.FamilyId, gvg.Id
}

//func (s *VirtualGroupTestSuite) TestSPExit() {
//	user := s.GenAndChargeAccounts(1, 1000000)[0]
//	// 1, create a new storage provider
//	sp := s.BaseSuite.CreateNewStorageProvider()
//	s.T().Logf("new SP Info: %s", sp.Info.String())
//
//	successorSp := s.BaseSuite.PickStorageProvider()
//
//	// 2, create a new gvg group for this storage provider
//	var secondarySPIDs []uint32
//	for _, ssp := range s.StorageProviders {
//		if ssp.Info.Id != successorSp.Info.Id {
//			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
//		}
//		if len(secondarySPIDs) == 6 {
//			break
//		}
//	}
//
//	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(sp, 0, secondarySPIDs, 1)
//
//	// 3. create object
//	s.BaseSuite.CreateObject(user, sp, gvgID, storagetestutil.GenRandomBucketName(), storagetestutil.GenRandomObjectName())
//
//	// 4. Create another gvg contains this new sp
//	var anotherSP *core.StorageProvider
//	for _, tsp := range s.StorageProviders {
//		if tsp.Info.Id != sp.Info.Id && tsp.Info.Id != successorSp.Info.Id {
//			anotherSP = tsp
//			break
//		}
//	}
//	var anotherSecondarySPIDs []uint32
//	for _, ssp := range s.StorageProviders {
//		if ssp.Info.Id != successorSp.Info.Id && ssp.Info.Id != anotherSP.Info.Id {
//			anotherSecondarySPIDs = append(anotherSecondarySPIDs, ssp.Info.Id)
//		}
//		if len(anotherSecondarySPIDs) == 5 {
//			break
//		}
//	}
//	anotherSecondarySPIDs = append(anotherSecondarySPIDs, sp.Info.Id)
//
//	anotherGVGID, _ := s.BaseSuite.CreateGlobalVirtualGroup(anotherSP, 0, anotherSecondarySPIDs, 1)
//
//	// 5. sp exit
//	s.SendTxBlock(sp.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
//		StorageProvider: sp.OperatorKey.GetAddr().String(),
//	})
//
//	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
//	s.Require().NoError(err)
//	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)
//
//	// 6. sp complete exit failed
//	s.SendTxBlockWithExpectErrorString(
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//		sp.OperatorKey,
//		"not swap out from all the family")
//
//	// 7. swap out, as primary sp
//	msgSwapOut := virtualgroupmoduletypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), familyID, nil, successorSp.Info.Id)
//	msgSwapOut.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
//	msgSwapOut.SuccessorSpApproval.Sig, err = successorSp.ApprovalKey.Sign(msgSwapOut.GetApprovalBytes())
//	s.Require().NoError(err)
//	s.SendTxBlock(sp.OperatorKey, msgSwapOut)
//
//	// 9. cancel swap out
//	msgCancelSwapOut := virtualgroupmoduletypes.NewMsgCancelSwapOut(sp.OperatorKey.GetAddr(), familyID, nil)
//	s.Require().NoError(err)
//	s.SendTxBlock(sp.OperatorKey, msgCancelSwapOut)
//
//	// 10. complete swap out, as primary sp
//	msgCompleteSwapOut := virtualgroupmoduletypes.NewMsgCompleteSwapOut(successorSp.OperatorKey.GetAddr(), familyID, nil)
//	s.Require().NoError(err)
//	s.SendTxBlockWithExpectErrorString(msgCompleteSwapOut, successorSp.OperatorKey, "The swap info not found in blockchain")
//
//	// 11 swap again
//	msgSwapOut = virtualgroupmoduletypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), familyID, nil, successorSp.Info.Id)
//	msgSwapOut.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
//	msgSwapOut.SuccessorSpApproval.Sig, err = successorSp.ApprovalKey.Sign(msgSwapOut.GetApprovalBytes())
//	s.Require().NoError(err)
//	s.SendTxBlock(sp.OperatorKey, msgSwapOut)
//
//	// 12. sp complete exit failed
//	s.SendTxBlockWithExpectErrorString(
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//		sp.OperatorKey,
//		"not swap out from all the family")
//
//	// 13. complete swap out, as primary sp
//	msgCompleteSwapOut = virtualgroupmoduletypes.NewMsgCompleteSwapOut(successorSp.OperatorKey.GetAddr(), familyID, nil)
//	s.Require().NoError(err)
//	s.SendTxBlock(successorSp.OperatorKey, msgCompleteSwapOut)
//
//	// 14. exist failed
//	s.SendTxBlockWithExpectErrorString(
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//		sp.OperatorKey,
//		"not swap out from all the gvgs")
//
//	// 15. swap out, as secondary sp
//	msgSwapOut2 := virtualgroupmoduletypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID}, successorSp.Info.Id)
//	msgSwapOut2.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
//	msgSwapOut2.SuccessorSpApproval.Sig, err = successorSp.ApprovalKey.Sign(msgSwapOut2.GetApprovalBytes())
//	s.Require().NoError(err)
//	s.SendTxBlock(sp.OperatorKey, msgSwapOut2)
//
//	// 16. exist failed
//	s.SendTxBlockWithExpectErrorString(
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//		sp.OperatorKey,
//		"not swap out from all the gvgs")
//
//	// 17 cancel swap out as secondary sp
//	msgCancelSwapOut = virtualgroupmoduletypes.NewMsgCancelSwapOut(sp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID})
//	s.Require().NoError(err)
//	s.SendTxBlock(sp.OperatorKey, msgCancelSwapOut)
//
//	// 18. swap
//	msgCompleteSwapOut2 := virtualgroupmoduletypes.NewMsgCompleteSwapOut(successorSp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID})
//	s.Require().NoError(err)
//	s.SendTxBlockWithExpectErrorString(msgCompleteSwapOut2, successorSp.OperatorKey, "The swap info not found in blockchain")
//
//	// 19. swap out again, as secondary sp
//	msgSwapOut2 = virtualgroupmoduletypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID}, successorSp.Info.Id)
//	msgSwapOut2.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
//	msgSwapOut2.SuccessorSpApproval.Sig, err = successorSp.ApprovalKey.Sign(msgSwapOut2.GetApprovalBytes())
//	s.Require().NoError(err)
//	s.SendTxBlock(sp.OperatorKey, msgSwapOut2)
//
//	// 20 complete swap out
//	msgCompleteSwapOut2 = virtualgroupmoduletypes.NewMsgCompleteSwapOut(successorSp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID})
//	s.Require().NoError(err)
//	s.SendTxBlock(successorSp.OperatorKey, msgCompleteSwapOut2)
//
//	// 18. sp complete exit success
//	s.SendTxBlock(
//		sp.OperatorKey,
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//	)
//}

//func (s *VirtualGroupTestSuite) TestSPExit_CreateAndDeleteBucket() {
//	user := s.GenAndChargeAccounts(1, 1000000)[0]
//	bucketName := storagetestutil.GenRandomBucketName()
//	objectName := storagetestutil.GenRandomObjectName()
//	// 1, create a new storage provider
//	sp := s.BaseSuite.CreateNewStorageProvider()
//	s.T().Logf("new SP Info: %s", sp.Info.String())
//
//	successorSp := s.BaseSuite.PickStorageProvider()
//
//	// 2, create a new gvg group for this storage provider
//	var secondarySPIDs []uint32
//	for _, ssp := range s.StorageProviders {
//		if ssp.Info.Id != successorSp.Info.Id {
//			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
//		}
//		if len(secondarySPIDs) == 6 {
//			break
//		}
//	}
//
//	gvgID, _ := s.BaseSuite.CreateGlobalVirtualGroup(sp, 0, secondarySPIDs, 1)
//
//	// 3. create object
//	s.BaseSuite.CreateObject(user, sp, gvgID, bucketName, objectName)
//
//	// 4. sp apply exit
//	s.SendTxBlock(sp.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
//		StorageProvider: sp.OperatorKey.GetAddr().String(),
//	})
//
//	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
//	s.Require().NoError(err)
//	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)
//
//	// 5. sp complete exit failed
//	s.SendTxBlockWithExpectErrorString(
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//		sp.OperatorKey,
//		"not swap out from all the family")
//
//	// 6. delete object
//	s.SendTxBlock(user, storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName))
//
//	// 7. delete bucket
//	s.SendTxBlock(user, storagetypes.NewMsgDeleteBucket(user.GetAddr(), bucketName))
//
//	// 8. delete gvg
//	s.SendTxBlock(sp.OperatorKey, virtualgroupmoduletypes.NewMsgDeleteGlobalVirtualGroup(sp.OperatorKey.GetAddr(), gvgID))
//	// 8. sp complete exit success
//	s.SendTxBlock(
//		sp.OperatorKey,
//		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: sp.OperatorKey.GetAddr().String()},
//	)
//}

func (s *VirtualGroupTestSuite) TestUpdateVirtualGroupParams() {
	// 1. create proposal
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	queryParamsResp, err := s.Client.VirtualGroupQueryClient.Params(context.Background(), &virtualgroupmoduletypes.QueryParamsRequest{})
	s.Require().NoError(err)
	updatedParams := queryParamsResp.Params
	updatedParams.MaxLocalVirtualGroupNumPerBucket = 1000

	msgUpdateParams := &virtualgroupmoduletypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    updatedParams,
	}

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update virtual group params", "Test update virtual group params")
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

	updatedQueryParamsResp, err := s.Client.VirtualGroupQueryClient.Params(context.Background(), &virtualgroupmoduletypes.QueryParamsRequest{})
	s.Require().NoError(err)
	if reflect.DeepEqual(updatedQueryParamsResp.Params, updatedParams) {
		s.T().Logf("update params success")
	} else {
		s.T().Errorf("update params failed")
	}
}
