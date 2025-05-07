package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/v5/crypto/bls"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	storagetestutil "github.com/bnb-chain/greenfield/testutil/storage"
	types3 "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
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

func (s *VirtualGroupTestSuite) getSecondarySPIDs(primarySPID uint32, excludeSecondarySP *uint32) []uint32 {
	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySPID {
			if excludeSecondarySP != nil && ssp.Info.Id == *excludeSecondarySP {
				continue
			}
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
		if len(secondarySPIDs) == 6 {
			break
		}
	}
	return secondarySPIDs
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroup(gvgID uint32) *virtualgroupmoduletypes.GlobalVirtualGroup {
	resp, err := s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroup
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroupsByFamily(familyID uint32) []*virtualgroupmoduletypes.GlobalVirtualGroup {
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

func (s *VirtualGroupTestSuite) querySpAvailableGlobalVirtualGroupFamilies(spId uint32) []uint32 {
	resp, err := s.Client.QuerySpAvailableGlobalVirtualGroupFamilies(
		context.Background(),
		&virtualgroupmoduletypes.QuerySPAvailableGlobalVirtualGroupFamiliesRequest{
			SpId: spId,
		})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroupFamilyIds
}

func (s *VirtualGroupTestSuite) querySpOptimalGlobalVirtualGroupFamily(spId uint32, strategy virtualgroupmoduletypes.PickVGFStrategy) uint32 {
	resp, err := s.Client.QuerySpOptimalGlobalVirtualGroupFamily(
		context.Background(),
		&virtualgroupmoduletypes.QuerySpOptimalGlobalVirtualGroupFamilyRequest{
			SpId:            spId,
			PickVgfStrategy: strategy,
		})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroupFamilyId
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
	spAvailableGvgFamilyIds := s.querySpAvailableGlobalVirtualGroupFamilies(primarySP.Info.Id)
	s.Require().Contains(spAvailableGvgFamilyIds, gvg.FamilyId)

	spOptimalGvgFamilyId := s.querySpOptimalGlobalVirtualGroupFamily(primarySP.Info.Id, virtualgroupmoduletypes.Strategy_Maximize_Free_Store_Size)
	s.Require().Contains(spAvailableGvgFamilyIds, spOptimalGvgFamilyId)

	srcGVGs := s.queryGlobalVirtualGroupsByFamily(gvg.FamilyId)

	secondarySPIDs := s.getSecondarySPIDs(primarySP.Info.Id, nil)
	s.BaseSuite.CreateGlobalVirtualGroup(primarySP, gvg.FamilyId, secondarySPIDs, 1)

	gvgs = s.queryGlobalVirtualGroupsByFamily(gvg.FamilyId)

	if len(srcGVGs) == len(gvgs) {
		secondarySPIDs = s.getSecondarySPIDs(primarySP.Info.Id, &secondarySPIDs[0])
		s.BaseSuite.CreateGlobalVirtualGroup(primarySP, gvg.FamilyId, secondarySPIDs, 1)
	}

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

	newGVGs := s.queryGlobalVirtualGroupsByFamily(newGVG.FamilyId)

	for _, gvg := range newGVGs {
		if gvg.Id == newGVG.Id {
			s.Assert().True(false)
		}
	}
	_, err = s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: newGVG.Id})
	s.Require().Error(err)

	// test number of secondary SP doest not match onchain requirement
	secondarySPIDs = append(secondarySPIDs, secondarySPIDs[0])
	msgCreateGVG := virtualgroupmoduletypes.MsgCreateGlobalVirtualGroup{
		StorageProvider: primarySP.OperatorKey.GetAddr().String(),
		FamilyId:        virtualgroupmoduletypes.NoSpecifiedFamilyId,
		SecondarySpIds:  secondarySPIDs,
		Deposit: sdk.Coin{
			Denom:  s.Config.Denom,
			Amount: types.NewIntFromInt64WithDecimal(1, types.DecimalBNB),
		},
	}
	s.SendTxBlockWithExpectErrorString(&msgCreateGVG, primarySP.OperatorKey, virtualgroupmoduletypes.ErrInvalidSecondarySPCount.Error())

	// test GVG has duplicated secondary Sp
	secondarySPIDs = make([]uint32, 0)
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySP.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
	}
	secondarySPIDs[len(secondarySPIDs)-1] = secondarySPIDs[0]
	msgCreateGVG = virtualgroupmoduletypes.MsgCreateGlobalVirtualGroup{
		StorageProvider: primarySP.OperatorKey.GetAddr().String(),
		FamilyId:        virtualgroupmoduletypes.NoSpecifiedFamilyId,
		SecondarySpIds:  secondarySPIDs,
		Deposit: sdk.Coin{
			Denom:  s.Config.Denom,
			Amount: types.NewIntFromInt64WithDecimal(1, types.DecimalBNB),
		},
	}
	s.SendTxBlockWithExpectErrorString(&msgCreateGVG, primarySP.OperatorKey, virtualgroupmoduletypes.ErrDuplicateSecondarySP.Error())

	// test create a duplicated GVG in a family
	secondarySPIDs = s.getSecondarySPIDs(primarySP.Info.Id, nil)
	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(primarySP, 0, secondarySPIDs, 1)
	gvgResp, err := s.Client.VirtualGroupQueryClient.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	s.Require().Equal(secondarySPIDs, gvgResp.GlobalVirtualGroup.SecondarySpIds)
	s.Require().Equal(familyID, gvgResp.GlobalVirtualGroup.FamilyId)

	msgCreateGVG = virtualgroupmoduletypes.MsgCreateGlobalVirtualGroup{
		StorageProvider: primarySP.OperatorKey.GetAddr().String(),
		FamilyId:        familyID,
		SecondarySpIds:  secondarySPIDs,
		Deposit: sdk.Coin{
			Denom:  s.Config.Denom,
			Amount: types.NewIntFromInt64WithDecimal(1, types.DecimalBNB),
		},
	}
	s.SendTxBlockWithExpectErrorString(&msgCreateGVG, primarySP.OperatorKey, virtualgroupmoduletypes.ErrDuplicateGVG.Error())
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
	txResp1 := s.SendTxBlock(user, &msgSettle)

	primaryBalanceAfter, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySp.FundingKey.GetAddr().String(),
	})
	s.Require().NoError(err)

	s.T().Logf("primaryBalance: %s, after: %s", primaryBalance.String(), primaryBalanceAfter.String())
	s.Require().True(primaryBalanceAfter.Balance.Amount.GT(primaryBalance.Balance.Amount))

	settleGVGFamilyEvent := filterSettleGVGFamilyEventFromTx(txResp1)
	s.Require().True(settleGVGFamilyEvent.Id == gvgFamilyId)
	s.Require().True(settleGVGFamilyEvent.SpId == gvgFamily.PrimarySpId)
	s.Require().True(settleGVGFamilyEvent.Amount.Equal(primaryBalanceAfter.Balance.Amount.Sub(primaryBalance.Balance.Amount)))

	// settle gvg
	msgSettle = virtualgroupmoduletypes.MsgSettle{
		StorageProvider:            user.GetAddr().String(),
		GlobalVirtualGroupFamilyId: 0,
		GlobalVirtualGroupIds:      []uint32{gvgId},
	}
	txResp2 := s.SendTxBlock(user, &msgSettle)

	secondaryBalancesAfter := make([]sdkmath.Int, 0, len(secondaryBalances))
	for _, addr := range secondarySpAddrs {
		tempResp, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
			Denom: s.Config.Denom, Address: addr,
		})
		s.Require().NoError(err)
		secondaryBalancesAfter = append(secondaryBalancesAfter, tempResp.Balance.Amount)
	}

	settleGVGEvent := filterSettleGVGEventFromTx(txResp2)
	s.Require().True(settleGVGEvent.Id == gvgId)
	for i := range secondaryBalances {
		s.T().Logf("secondaryBalance: %s, after: %s", secondaryBalances[i].String(), secondaryBalancesAfter[i].String())
		s.Require().True(secondaryBalancesAfter[i].GT(secondaryBalances[i]))
		s.Require().True(settleGVGEvent.Amount.Equal(secondaryBalancesAfter[i].Sub(secondaryBalances[i])))
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

func (s *VirtualGroupTestSuite) TestEmptyGlobalVirtualGroupFamily() {
	primarySP := s.BaseSuite.PickStorageProvider()
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	secondarySPIDs := s.getSecondarySPIDs(primarySP.Info.Id, nil)

	// The Sp creates a family which has 1 GVG.
	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(primarySP, 0, secondarySPIDs, 1)
	gvgs := s.queryGlobalVirtualGroupsByFamily(familyID)
	s.Require().Equal(1, len(gvgs))

	// a User creates an object served by this GVG
	bucketName := storagetestutil.GenRandomBucketName()
	objectName := storagetestutil.GenRandomObjectName()
	s.BaseSuite.CreateObject(user, primarySP, gvgID, bucketName, objectName)

	// The User deletes the object
	s.SendTxBlock(user, storagetypes.NewMsgDeleteObject(user.GetAddr(), bucketName, objectName))

	// object isn't found onchain
	_, err := s.Client.HeadObject(context.Background(), &storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName,
		ObjectName: objectName,
	})
	s.Require().Error(err)

	// The SP deletes the GVG
	msgDeleteGVG := virtualgroupmoduletypes.MsgDeleteGlobalVirtualGroup{
		StorageProvider:      primarySP.OperatorKey.GetAddr().String(),
		GlobalVirtualGroupId: gvgID,
	}
	s.SendTxBlock(primarySP.OperatorKey, &msgDeleteGVG)
	_, err = s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().Error(err)

	// The bucket onchain still shows the family info, and the family is indeed exist
	bucket, err := s.Client.HeadBucket(context.Background(), &storagetypes.QueryHeadBucketRequest{
		BucketName: bucketName,
	})
	s.Require().NoError(err)
	s.Require().Equal(familyID, bucket.BucketInfo.GlobalVirtualGroupFamilyId)

	family, err := s.Client.GlobalVirtualGroupFamily(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{
		FamilyId: bucket.BucketInfo.GlobalVirtualGroupFamilyId,
	})
	s.Require().NoError(err)
	s.Require().Equal(0, len(family.GlobalVirtualGroupFamily.GlobalVirtualGroupIds))

	//the SP can create new GVG on this empty family
	newGVGID, _ := s.BaseSuite.CreateGlobalVirtualGroup(primarySP, familyID, secondarySPIDs, 1)
	gvgs = s.queryGlobalVirtualGroupsByFamily(familyID)
	s.Require().Equal(1, len(gvgs))
	s.Require().Equal(gvgs[0].Id, newGVGID)
}

func (s *VirtualGroupTestSuite) TestSPExit() {
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	// 1. create an SP-x that wants to exit
	spx := s.BaseSuite.CreateNewStorageProvider()
	s.T().Logf("new SP(successor) Info: %s", spx.Info.String())

	// 2. create a successor SP-y
	spy := s.BaseSuite.CreateNewStorageProvider()
	s.T().Logf("new SP(successor) Info: %s", spy.Info.String())

	// 3, SP-x create a new family with a GVG. Family {GVG: [x|2, 3, 4, 5, 6, 7]}
	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(spx, 0, []uint32{2, 3, 4, 5, 6, 7}, 1)

	// 4. create object
	s.BaseSuite.CreateObject(user, spx, gvgID, storagetestutil.GenRandomBucketName(), storagetestutil.GenRandomObjectName())

	// 5. SP-2 creates gvg contains SP-x  [2|x,3,4,5,6,7]
	sp2 := s.BaseSuite.PickStorageProviderByID(2)
	s.T().Logf("SP 2 Info: %s", spx.Info.String())
	gvgID2, familyID2 := s.BaseSuite.CreateGlobalVirtualGroup(sp2, 0, []uint32{spx.Info.Id, 3, 4, 5, 6, 7}, 1)

	// 6. SP-x declare to exit
	s.SendTxBlock(spx.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spx.OperatorKey.GetAddr().String(),
	})
	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 7. SP-x complete exit, it would fail due to there are family and GVG binded to it.
	s.SendTxBlockWithExpectErrorString(
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
			StorageProvider: spx.OperatorKey.GetAddr().String(),
			Operator:        spx.OperatorKey.GetAddr().String()},
		spx.OperatorKey,
		"not swap out from all the family")

	// 8.The SP-y will reserve the swapIn as primary sp for the SP-x's family
	msgReserveSwapIn := virtualgroupmoduletypes.NewMsgReserveSwapIn(spy.OperatorKey.GetAddr(), spx.Info.Id, familyID, 0)
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)

	// 9 query the swapInInfo onchain, show reservation is recorded onchain
	swapInInfo, err := s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().NoError(err)
	s.Require().Equal(swapInInfo.SwapInInfo.SuccessorSpId, spy.Info.Id)
	s.Require().Equal(swapInInfo.SwapInInfo.TargetSpId, spx.Info.Id)

	// 10. SP-y cancel the swapIn
	msgCancelSwapIn := virtualgroupmoduletypes.NewMsgCancelSwapIn(spy.OperatorKey.GetAddr(), familyID, 0)
	s.Require().NoError(err)
	s.SendTxBlock(spy.OperatorKey, msgCancelSwapIn)

	// 11. SP-y wants to complete swap in, as primary sp, failure is expected.
	msgCompleteSwapIn := virtualgroupmoduletypes.NewMsgCompleteSwapIn(spy.OperatorKey.GetAddr(), familyID, 0)
	s.Require().NoError(err)
	s.SendTxBlockWithExpectErrorString(msgCompleteSwapIn, spy.OperatorKey, "The swap info not found in blockchain")

	// 12 SP-y reserves swapIn again, and complete swapIn
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)

	swapInInfo, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().NoError(err)
	s.Require().Equal(swapInInfo.SwapInInfo.SuccessorSpId, spy.Info.Id)
	s.Require().Equal(swapInInfo.SwapInInfo.TargetSpId, spx.Info.Id)

	s.SendTxBlock(spy.OperatorKey, msgCompleteSwapIn)

	// 13. query the swapInInfo should be not found onChain.
	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().Error(err)

	// 14. The SP-y has replaced Sp-x in the family successfully, and it becomes the primary SP.  Family {GVG: [y|2, 3, 4, 5, 6, 7]}
	familyAfterSwapIn, err := s.Client.GlobalVirtualGroupFamily(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{FamilyId: familyID})
	s.Require().NoError(err)
	s.Require().Equal(spy.Info.Id, familyAfterSwapIn.GlobalVirtualGroupFamily.PrimarySpId)
	s.Require().Equal(gvgID, familyAfterSwapIn.GlobalVirtualGroupFamily.GlobalVirtualGroupIds[0])

	gvgAfterSwapIn, err := s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	s.Require().Equal(spy.Info.Id, gvgAfterSwapIn.GlobalVirtualGroup.PrimarySpId)
	s.Require().Equal(familyID, gvgAfterSwapIn.GlobalVirtualGroup.FamilyId)
	s.Require().Equal([]uint32{2, 3, 4, 5, 6, 7}, gvgAfterSwapIn.GlobalVirtualGroup.SecondarySpIds)

	// 15. SP-x tries to complete exit, but would fail, since SP-2 has a GVG that includes SP-x [2|x,3,4,5,6,7]
	s.SendTxBlockWithExpectErrorString(
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
			StorageProvider: spx.OperatorKey.GetAddr().String(),
			Operator:        spx.OperatorKey.GetAddr().String()},
		spx.OperatorKey,
		"not swap out from all the gvgs")

	// 16. SP-y reserves the swapIn, as secondary sp
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spy.OperatorKey.GetAddr(), spx.Info.Id, 0, gvgID2)
	s.Require().NoError(err)
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)

	// query the swapInInfo
	swapInInfo, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupId: gvgID2,
	})
	s.Require().NoError(err)
	s.Require().Equal(spy.Info.Id, swapInInfo.SwapInInfo.SuccessorSpId)
	s.Require().Equal(spx.Info.Id, swapInInfo.SwapInInfo.TargetSpId)

	// 17 SP-y cancels swap in as secondary sp
	msgCancelSwapIn = virtualgroupmoduletypes.NewMsgCancelSwapIn(spy.OperatorKey.GetAddr(), 0, gvgID2)
	s.Require().NoError(err)
	s.SendTxBlock(spy.OperatorKey, msgCancelSwapIn)

	// 18 query the swapInInfo not found
	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupId: gvgID2,
	})
	s.Require().Error(err)

	// 19. SP-y tries to complete swapIn, failure expected.
	msgCompleteSwapIn = virtualgroupmoduletypes.NewMsgCompleteSwapIn(spy.OperatorKey.GetAddr(), 0, gvgID2)
	s.SendTxBlockWithExpectErrorString(msgCompleteSwapIn, spy.OperatorKey, "The swap info not found in blockchain")

	// 20. SP-y reserves swapIn again, as secondary sp
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spy.OperatorKey.GetAddr(), spx.Info.Id, 0, gvgID2)
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)

	// 21 SP-y complete the swap in
	msgCompleteSwapIn = virtualgroupmoduletypes.NewMsgCompleteSwapIn(spy.OperatorKey.GetAddr(), 0, gvgID2)
	s.SendTxBlock(spy.OperatorKey, msgCompleteSwapIn)

	// 22 SP-y has replaced Sp-x in Sp-2's GVG, the GVG becomes [2|y, 3, 4, 5, 6, 7]}
	gvgAfterSwapIn, err = s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID2})
	s.Require().NoError(err)
	s.Require().Equal(sp2.Info.Id, gvgAfterSwapIn.GlobalVirtualGroup.PrimarySpId)
	s.Require().Equal(familyID2, gvgAfterSwapIn.GlobalVirtualGroup.FamilyId)
	s.Require().Equal([]uint32{spy.Info.Id, 3, 4, 5, 6, 7}, gvgAfterSwapIn.GlobalVirtualGroup.SecondarySpIds)

	// 23. SP-x complete exit success
	s.SendTxBlock(
		spx.OperatorKey,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: spx.OperatorKey.GetAddr().String(), Operator: spx.OperatorKey.GetAddr().String()},
	)

	// 24 SP-x no longer found on chain
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().Error(err)
}

func (s *VirtualGroupTestSuite) TestSPExit2() {
	// 1. Create SP-x, y, z, SP-x wants to exit, SP-y, SP-z will SwapIn SP-x family and GVG
	spx := s.BaseSuite.CreateNewStorageProvider()
	spy := s.BaseSuite.CreateNewStorageProvider()
	spz := s.BaseSuite.CreateNewStorageProvider()

	// 2 SP-x creates a new family with a GVG. Family: {GVG: [x|y, 3, 4, 5, 6, 7]}, SP-y is a secondary on this GVG.
	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(spx, 0, []uint32{spy.Info.Id, 3, 4, 5, 6, 7}, 1)

	// 3. SP-x announces to exit
	s.SendTxBlock(spx.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spx.OperatorKey.GetAddr().String(),
	})
	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 4. SP-y also announces to exit, multiple SP exit concurrently not allowed.
	s.SendTxBlockWithExpectErrorString(&virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spy.OperatorKey.GetAddr().String()}, spy.OperatorKey, "")
	resp, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spy.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_IN_SERVICE)

	// 5. SP-y reserves the swapIn and complete it for SP-x family {GVG: [x|y, 3, 4, 5, 6, 7]}, then have {[y|y,3,4,5,6,7]}
	// Sp-y will act both role in this GVG, primary and one of the secondary.
	msgReserveSwapIn := virtualgroupmoduletypes.NewMsgReserveSwapIn(spy.OperatorKey.GetAddr(), spx.Info.Id, familyID, 0)
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)
	msgCompleteSwapIn := virtualgroupmoduletypes.NewMsgCompleteSwapIn(spy.OperatorKey.GetAddr(), familyID, 0)
	s.SendTxBlock(spy.OperatorKey, msgCompleteSwapIn)

	gvgAfterSwapIn, err := s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	s.Require().Equal(spy.Info.Id, gvgAfterSwapIn.GlobalVirtualGroup.PrimarySpId)
	s.Require().Equal(familyID, gvgAfterSwapIn.GlobalVirtualGroup.FamilyId)
	s.Require().Equal([]uint32{spy.Info.Id, 3, 4, 5, 6, 7}, gvgAfterSwapIn.GlobalVirtualGroup.SecondarySpIds)

	// 6. SP-x now can complete the exit, since it has only 1 family and it has been taken by SP-y
	s.SendTxBlock(
		spx.OperatorKey,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
			StorageProvider: spx.OperatorKey.GetAddr().String(),
			Operator:        spx.OperatorKey.GetAddr().String()},
	)
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().Error(err)

	// 7. SP-y try to exit as well, not allowed and rejected by chain due to it has GVG [y|y,3,4,5,6,7] that break the redundancy requirement
	s.SendTxBlockWithExpectErrorString(&virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spy.OperatorKey.GetAddr().String()}, spy.OperatorKey, "break the redundancy requirement")
	resp, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spy.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_IN_SERVICE)

	// 8 SP-z reserves and completes the swap, GVG becomes  [y|z,3,4,5,6,7],
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spz.OperatorKey.GetAddr(), spy.Info.Id, 0, gvgID)
	s.SendTxBlock(spz.OperatorKey, msgReserveSwapIn)
	msgCompleteSwapIn = virtualgroupmoduletypes.NewMsgCompleteSwapIn(spz.OperatorKey.GetAddr(), 0, gvgID)
	s.SendTxBlock(spz.OperatorKey, msgCompleteSwapIn)

	gvgAfterSwapIn, err = s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	s.Require().Equal(spy.Info.Id, gvgAfterSwapIn.GlobalVirtualGroup.PrimarySpId)
	s.Require().Equal(familyID, gvgAfterSwapIn.GlobalVirtualGroup.FamilyId)
	s.Require().Equal([]uint32{spz.Info.Id, 3, 4, 5, 6, 7}, gvgAfterSwapIn.GlobalVirtualGroup.SecondarySpIds)

	// 9 SP-y can declare to exit
	s.SendTxBlock(spy.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spy.OperatorKey.GetAddr().String()})
	resp, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spy.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 10 SP-z reserves and complete swapIn for family {[y|z,3,4,5,6,7]} and becomes {[z|z,3,4,5,6,7]}
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spz.OperatorKey.GetAddr(), spy.Info.Id, familyID, 0)
	s.SendTxBlock(spz.OperatorKey, msgReserveSwapIn)
	msgCompleteSwapIn = virtualgroupmoduletypes.NewMsgCompleteSwapIn(spz.OperatorKey.GetAddr(), familyID, 0)
	s.SendTxBlock(spz.OperatorKey, msgCompleteSwapIn)

	gvgAfterSwapIn, err = s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	s.Require().Equal(spz.Info.Id, gvgAfterSwapIn.GlobalVirtualGroup.PrimarySpId)
	s.Require().Equal(familyID, gvgAfterSwapIn.GlobalVirtualGroup.FamilyId)
	s.Require().Equal([]uint32{spz.Info.Id, 3, 4, 5, 6, 7}, gvgAfterSwapIn.GlobalVirtualGroup.SecondarySpIds)

	// 11 complete SPy's exit by sp-z
	s.SendTxBlock(
		spz.OperatorKey,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
			StorageProvider: spy.OperatorKey.GetAddr().String(),
			Operator:        spz.OperatorKey.GetAddr().String()},
	)
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spy.Info.Id})
	s.Require().Error(err)
}

func (s *VirtualGroupTestSuite) TestSPForcedExit() {

	ctx := context.Background()
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	// 1. create SPs
	spx := s.BaseSuite.CreateNewStorageProvider()
	spy := s.BaseSuite.CreateNewStorageProvider()

	// get the dynamic balance of gov address account i payment module
	govAddrInPaymentBalance, err := s.Client.DynamicBalance(context.Background(), &types3.QueryDynamicBalanceRequest{
		Account: types3.GovernanceAddress.String(),
	})
	s.Require().NoError(err)
	s.T().Logf("payment module gov stream record balance is %s", core.YamlString(govAddrInPaymentBalance))

	// 2. SP-x creates a new family with a gvg: {[x|2,3,4,5,6,7]}
	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(spx, 0, []uint32{2, 3, 4, 5, 6, 7}, 1)

	//  User creates an object and sealed by the SP-x's GVG
	bucketName := storagetestutil.GenRandomBucketName()
	objectName := storagetestutil.GenRandomBucketName()
	s.BaseSuite.CreateObject(user, spx, gvgID, bucketName, objectName)
	objectResp, err := s.Client.HeadObject(context.Background(), &storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName, ObjectName: objectName,
	})
	s.Require().NoError(err)

	// 3. create a proposal that puts SP-x to FORCE_EXIT
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	msgForcedExit := virtualgroupmoduletypes.NewMsgStorageProviderForcedExit(govAddr, spx.OperatorKey.GetAddr())

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgForcedExit}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "put SP to force exit status", "put SP to force exit status")
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
	s.Require().True(proposalID != 0)
	queryProposal := &v1.QueryProposalRequest{ProposalId: uint64(proposalID)}
	_, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote := v1.NewMsgVote(s.Validator.GetAddr(), uint64(proposalID), v1.OptionYes, "test")
	txRes := s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := v1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod + time.Second)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(v1.ProposalStatus_PROPOSAL_STATUS_PASSED, proposalRes.Proposal.Status)

	// 6. SP-x status will be FORCE_EXITING
	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_FORCED_EXITING, resp.StorageProvider.Status)

	// 7. SP-x successor SP try swapIn family
	msgReserveSwapIn := virtualgroupmoduletypes.NewMsgReserveSwapIn(spy.OperatorKey.GetAddr(), spx.Info.Id, familyID, 0)
	s.Require().NoError(err)
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)

	// query the swapInInfo
	swapInInfo, err := s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().NoError(err)
	s.Require().Equal(swapInInfo.SwapInInfo.SuccessorSpId, spy.Info.Id)
	s.Require().Equal(swapInInfo.SwapInInfo.TargetSpId, spx.Info.Id)

	// swapin info not found
	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().Error(err)

	// SP-y is able to discontinue the object as a successor Primary SP
	msgDiscontinueObject := &storagetypes.MsgDiscontinueObject{
		Operator:   spy.GcKey.GetAddr().String(),
		BucketName: bucketName,
		ObjectIds:  []sdkmath.Uint{objectResp.ObjectInfo.Id},
	}
	s.SendTxBlock(spy.GcKey, msgDiscontinueObject)
	time.Sleep(5 * time.Second)
	_, err = s.Client.HeadObject(context.Background(), &storagetypes.QueryHeadObjectRequest{
		BucketName: bucketName, ObjectName: objectName,
	})
	s.Require().Error(err)

	// 8. SP-y complete SwapIn
	msgCompleteSwapIn := virtualgroupmoduletypes.NewMsgCompleteSwapIn(spy.OperatorKey.GetAddr(), familyID, 0)
	s.SendTxBlock(spy.OperatorKey, msgCompleteSwapIn)

	// 9. query the swapInInfo should be not found
	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().Error(err)

	gvgAfterSwapIn, err := s.Client.GlobalVirtualGroup(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	s.Require().Equal(spy.Info.Id, gvgAfterSwapIn.GlobalVirtualGroup.PrimarySpId)
	s.Require().Equal(familyID, gvgAfterSwapIn.GlobalVirtualGroup.FamilyId)
	s.Require().Equal([]uint32{2, 3, 4, 5, 6, 7}, gvgAfterSwapIn.GlobalVirtualGroup.SecondarySpIds)

	// 11 User help complete the exit
	s.SendTxBlock(
		user,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: spx.OperatorKey.GetAddr().String(),
			Operator: user.GetAddr().String()},
	)
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().Error(err)

	govAddrInPaymentBalanceAfter, err := s.Client.DynamicBalance(context.Background(), &types3.QueryDynamicBalanceRequest{
		Account: types3.GovernanceAddress.String(),
	})
	s.Require().NoError(err)
	s.T().Logf("payment module gov stream record balance is %s", core.YamlString(govAddrInPaymentBalanceAfter))
	s.Require().Equal(govAddrInPaymentBalance.BankBalance.Add(resp.StorageProvider.TotalDeposit), govAddrInPaymentBalanceAfter.BankBalance)
}

func (s *VirtualGroupTestSuite) updateParams(params virtualgroupmoduletypes.Params) {
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	msgUpdateParams := &virtualgroupmoduletypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    params,
	}
	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update virtual group params", "update virtual group params")
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
	s.Require().True(proposalID != 0)
	queryProposal := &v1.QueryProposalRequest{ProposalId: uint64(proposalID)}
	_, err = s.Client.GovQueryClientV1.Proposal(context.Background(), queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote := v1.NewMsgVote(s.Validator.GetAddr(), uint64(proposalID), v1.OptionYes, "test")
	txRes := s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := v1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(context.Background(), &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod + time.Second)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(context.Background(), queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, v1.ProposalStatus_PROPOSAL_STATUS_PASSED)
}

func (s *VirtualGroupTestSuite) TestSPExit_SwapInfo_Expired() {

	// update the param, swapInInfo validity period is 10s
	queryParamsResp, err := s.Client.VirtualGroupQueryClient.Params(context.Background(), &virtualgroupmoduletypes.QueryParamsRequest{})
	s.Require().NoError(err)
	updatedParams := queryParamsResp.Params

	swapInValidityPeriod := sdk.NewInt(10)
	updatedParams.SwapInValidityPeriod = &swapInValidityPeriod // the swapInInfo will expire in 10 seconds
	s.updateParams(updatedParams)

	queryParamsResp, err = s.Client.VirtualGroupQueryClient.Params(context.Background(), &virtualgroupmoduletypes.QueryParamsRequest{})
	s.Require().NoError(err)
	s.Require().Equal(uint64(10), queryParamsResp.Params.SwapInValidityPeriod.Uint64())

	// 1. create an SP-x that wants to exit
	spx := s.BaseSuite.CreateNewStorageProvider()
	s.T().Logf("new SPx(successor) Info: %s", spx.Info.String())

	// 2. create a successor SP-y,  successor SP-z
	spy := s.BaseSuite.CreateNewStorageProvider()
	s.T().Logf("new SPy(successor) Info: %s", spy.Info.String())
	spz := s.BaseSuite.CreateNewStorageProvider()
	s.T().Logf("new SPz(successor) Info: %s", spz.Info.String())

	// 3 SP-x create a new family with a GVG. Family {GVG: [x|2, 3, 4, 5, 6, 7]}
	_, familyID := s.BaseSuite.CreateGlobalVirtualGroup(spx, 0, []uint32{2, 3, 4, 5, 6, 7}, 1)

	// 4. SP-x declare to exit
	s.SendTxBlock(spx.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spx.OperatorKey.GetAddr().String(),
	})
	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 5 SP-y reserves the swapIn family
	msgReserveSwapIn := virtualgroupmoduletypes.NewMsgReserveSwapIn(spy.OperatorKey.GetAddr(), spx.Info.Id, familyID, 0)
	s.SendTxBlock(spy.OperatorKey, msgReserveSwapIn)

	// 6 query the swapInInfo onchain, show reservation is recorded onchain
	swapInInfo, err := s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().NoError(err)
	s.Require().Equal(swapInInfo.SwapInInfo.SuccessorSpId, spy.Info.Id)
	s.Require().Equal(swapInInfo.SwapInInfo.TargetSpId, spx.Info.Id)

	// 7. SP-z try to swapIn family, it would fail
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spz.OperatorKey.GetAddr(), spx.Info.Id, familyID, 0)
	s.SendTxBlockWithExpectErrorString(msgReserveSwapIn, spz.OperatorKey, "already exist SP")

	// 8 waits for swapIno is expired, it can still be queried but with expired info
	time.Sleep(11 * time.Second)
	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().NoError(err)
	s.Require().True(swapInInfo.SwapInInfo.ExpirationTime < uint64(time.Now().Unix()))

	// 9 SP-y try to complete swapIn family, it is allowed, since no other SP apply swapIn, and swapIn info will be gone after completion
	msgCompleteSwapIn := virtualgroupmoduletypes.NewMsgCompleteSwapIn(spy.OperatorKey.GetAddr(), familyID, 0)
	s.SendTxBlock(spy.OperatorKey, msgCompleteSwapIn)

	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().Error(err)

	// 10 sp-x completes the exit
	s.SendTxBlock(
		spx.OperatorKey,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: spx.OperatorKey.GetAddr().String(),
			Operator: spx.OperatorKey.GetAddr().String()},
	)
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spx.Info.Id})
	s.Require().Error(err)

	// 11. SP-y declare to exit
	s.SendTxBlock(spy.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: spy.OperatorKey.GetAddr().String(),
	})
	resp, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spy.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 12 SP-z reserves the swapIn family but not complete it
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spz.OperatorKey.GetAddr(), spy.Info.Id, familyID, 0)
	s.SendTxBlock(spz.OperatorKey, msgReserveSwapIn)

	// 13 query the swapInInfo onchain, show reservation is recorded onchain
	_, err = s.Client.SwapInInfo(context.Background(), &virtualgroupmoduletypes.QuerySwapInInfoRequest{
		GlobalVirtualGroupFamilyId: familyID,
	})
	s.Require().NoError(err)
	// 14. wait for swapIn info expired
	time.Sleep(11 * time.Second)

	// 15 a new SP can reserve since the prev one is expired
	spn := s.BaseSuite.CreateNewStorageProvider()
	msgReserveSwapIn = virtualgroupmoduletypes.NewMsgReserveSwapIn(spn.OperatorKey.GetAddr(), spy.Info.Id, familyID, 0)
	s.SendTxBlock(spn.OperatorKey, msgReserveSwapIn)

	msgCompleteSwapIn = virtualgroupmoduletypes.NewMsgCompleteSwapIn(spn.OperatorKey.GetAddr(), familyID, 0)
	s.SendTxBlock(spn.OperatorKey, msgCompleteSwapIn)

	// 16 spy complete exit
	s.SendTxBlock(
		spy.OperatorKey,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{StorageProvider: spy.OperatorKey.GetAddr().String(),
			Operator: spy.OperatorKey.GetAddr().String()},
	)
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: spy.Info.Id})
	s.Require().Error(err)
}

func filterSettleGVGEventFromTx(txRes *sdk.TxResponse) virtualgroupmoduletypes.EventSettleGlobalVirtualGroup {
	idStr, amountStr := "", ""
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "greenfield.virtualgroup.EventSettleGlobalVirtualGroup" {
			for _, attr := range event.Attributes {
				if attr.Key == "id" {
					idStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "amount" {
					amountStr = strings.Trim(attr.Value, `"`)
				}
			}
		}
	}
	id, _ := strconv.ParseInt(idStr, 10, 32)
	amount := sdkmath.NewUintFromString(amountStr)
	return virtualgroupmoduletypes.EventSettleGlobalVirtualGroup{
		Id:     uint32(id),
		Amount: sdkmath.NewInt(int64(amount.Uint64())),
	}
}

func filterSettleGVGFamilyEventFromTx(txRes *sdk.TxResponse) virtualgroupmoduletypes.EventSettleGlobalVirtualGroupFamily {
	idStr, spIdStr, amountStr := "", "", ""
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "greenfield.virtualgroup.EventSettleGlobalVirtualGroupFamily" {
			for _, attr := range event.Attributes {
				if attr.Key == "id" {
					idStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "sp_id" {
					spIdStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "amount" {
					amountStr = strings.Trim(attr.Value, `"`)
				}
			}
		}
	}
	id, _ := strconv.ParseInt(idStr, 10, 32)
	spId, _ := strconv.ParseInt(spIdStr, 10, 32)
	amount := sdkmath.NewUintFromString(amountStr)
	return virtualgroupmoduletypes.EventSettleGlobalVirtualGroupFamily{
		Id:     uint32(id),
		SpId:   uint32(spId),
		Amount: sdkmath.NewInt(int64(amount.Uint64())),
	}
}
