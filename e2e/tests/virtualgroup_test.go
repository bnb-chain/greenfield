package tests

import (
	"bytes"
	"context"
	"fmt"
	"math"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	types2 "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	storagetestutil "github.com/bnb-chain/greenfield/testutil/storage"
	"github.com/bnb-chain/greenfield/types/common"
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

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroup(gvgID uint32) *virtualgroupmoduletypes.GlobalVirtualGroup {
	resp, err := s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroup
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroupByFamily(spID, familyID uint32) []*virtualgroupmoduletypes.GlobalVirtualGroup {
	resp, err := s.Client.GlobalVirtualGroupByFamilyID(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupByFamilyIDRequest{
			StorageProviderId:          spID,
			GlobalVirtualGroupFamilyId: familyID,
		})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroups
}

func (s *VirtualGroupTestSuite) queryGlobalVirtualGroupFamilies(spID uint32) []*virtualgroupmoduletypes.GlobalVirtualGroupFamily {
	resp, err := s.Client.GlobalVirtualGroupFamilies(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupFamiliesRequest{StorageProviderId: spID})
	s.Require().NoError(err)
	return resp.GlobalVirtualGroupFamilies
}

func (s *VirtualGroupTestSuite) TestBasic() {
	primarySP := s.StorageProviders[0]

	gvgFamilies := s.queryGlobalVirtualGroupFamilies(primarySP.Info.Id)
	s.Require().Greater(len(gvgFamilies), 0)

	family := gvgFamilies[0]
	s.T().Log(family.String())

	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != primarySP.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
	}
	s.BaseSuite.CreateGlobalVirtualGroup(&primarySP, family.Id, secondarySPIDs, 1)

	gvgs := s.queryGlobalVirtualGroupByFamily(primarySP.Info.Id, family.Id)
	s.Require().Equal(len(gvgs), len(family.GlobalVirtualGroupIds)+1)

	oldGVGIDs := make(map[uint32]bool)
	for _, id := range family.GlobalVirtualGroupIds {
		oldGVGIDs[id] = true
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
		FundingAddress:       primarySP.FundingKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
		Deposit:              sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(1, types.DecimalBNB)),
	}
	s.SendTxBlock(primarySP.FundingKey, &msgDeposit)

	gvgAfterDeposit := s.queryGlobalVirtualGroup(newGVG.Id)
	s.Require().Equal(gvgAfterDeposit.TotalDeposit.Int64(), int64(2000000000000000000))

	// test withdraw
	balance, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySP.FundingKey.GetAddr().String()})
	s.Require().NoError(err)

	msgWithdraw := virtualgroupmoduletypes.MsgWithdraw{
		FundingAddress:       primarySP.FundingKey.GetAddr().String(),
		Withdraw:             sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(1, types.DecimalBNB)),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.FundingKey, &msgWithdraw)
	balanceAfterWithdraw, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySP.FundingKey.GetAddr().String()})
	s.Require().NoError(err)

	s.T().Logf("balance: %s, after: %s", balance.String(), balanceAfterWithdraw.String())
	s.Require().Equal(balanceAfterWithdraw.Balance.Amount.Sub(balance.Balance.Amount).Int64(), int64(999994000000000000))

	// test delete gvg
	msgDeleteGVG := virtualgroupmoduletypes.MsgDeleteGlobalVirtualGroup{
		PrimarySpAddress:     primarySP.OperatorKey.GetAddr().String(),
		GlobalVirtualGroupId: newGVG.Id,
	}
	s.SendTxBlock(primarySP.OperatorKey, &msgDeleteGVG)

	newGVGs := s.queryGlobalVirtualGroupByFamily(primarySP.Info.Id, family.Id)

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
	_, _, primarySp, secondarySps, gvgFamilyId, gvgId := s.createObject()
	s.T().Log("gvg family", gvgFamilyId, "gvg", gvgId)

	queryFamilyResp, err := s.Client.GlobalVirtualGroupFamily(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupFamilyRequest{
			StorageProviderId: primarySp.Info.Id,
			FamilyId:          gvgFamilyId,
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

	var lastSecondaySp *core.StorageProvider
	secondarySpAddrs := make([]string, 0)
	for _, secondarySp := range secondarySps {
		if _, ok := secondarySpIds[secondarySp.Info.Id]; ok {
			secondarySpAddrs = append(secondarySpAddrs, secondarySp.FundingKey.GetAddr().String())
			lastSecondaySp = &secondarySp
		}
	}

	// sleep seconds
	time.Sleep(3 * time.Second)

	primaryBalance, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySp.FundingKey.GetAddr().String()})
	s.Require().NoError(err)
	secondaryBalances := make([]sdkmath.Int, 0)
	for _, addr := range secondarySpAddrs {
		tempResp, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
			Denom: s.Config.Denom, Address: addr})
		s.Require().NoError(err)
		secondaryBalances = append(secondaryBalances, tempResp.Balance.Amount)
	}

	// settle gvg family
	msgSettle := virtualgroupmoduletypes.MsgSettle{
		FundingAddress:             primarySp.FundingKey.GetAddr().String(),
		GlobalVirtualGroupFamilyId: gvgFamily.Id,
	}
	s.SendTxBlock(primarySp.FundingKey, &msgSettle)

	primaryBalanceAfter, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
		Denom: s.Config.Denom, Address: primarySp.FundingKey.GetAddr().String()})
	s.Require().NoError(err)

	s.T().Logf("primaryBalance: %s, after: %s", primaryBalance.String(), primaryBalanceAfter.String())
	s.Require().True(primaryBalanceAfter.Balance.Amount.GT(primaryBalance.Balance.Amount))

	// settle gvg
	msgSettle = virtualgroupmoduletypes.MsgSettle{
		FundingAddress:             lastSecondaySp.FundingKey.GetAddr().String(),
		GlobalVirtualGroupFamilyId: 0,
		GlobalVirtualGroupIds:      []uint32{gvgId},
	}
	s.SendTxBlock(lastSecondaySp.FundingKey, &msgSettle)

	secondaryBalancesAfter := make([]sdkmath.Int, 0, len(secondaryBalances))
	for _, addr := range secondarySpAddrs {
		tempResp, err := s.Client.BankQueryClient.Balance(context.Background(), &types2.QueryBalanceRequest{
			Denom: s.Config.Denom, Address: addr})
		s.Require().NoError(err)
		secondaryBalancesAfter = append(secondaryBalancesAfter, tempResp.Balance.Amount)
	}

	for i := range secondaryBalances {
		s.T().Logf("secondaryBalance: %s, after: %s", secondaryBalances[i].String(), secondaryBalancesAfter[i].String())
		s.Require().True(secondaryBalancesAfter[i].GT(secondaryBalances[i]))
	}
}

func (s *VirtualGroupTestSuite) createObject() (string, string, core.StorageProvider, []core.StorageProvider, uint32, uint32) {
	var err error
	sp := s.StorageProviders[0]
	secondarySps := make([]core.StorageProvider, 0)
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
	signBz := storagetypes.NewSecondarySpSealObjectSignDoc(queryHeadObjectResponse.ObjectInfo.Id, gvgId, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetSignBytes()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := core.BlsSignAndVerify(s.StorageProviders[i], signBz)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
		if s.StorageProviders[i].Info.Id != sp.Info.Id {
			secondarySps = append(secondarySps, s.StorageProviders[i])
		}
	}
	aggBlsSig, err := core.BlsAggregateAndVerify(secondarySPBlsPubKeys, signBz, secondarySigs)
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

func (s *VirtualGroupTestSuite) TestSPExit() {
	// 1, create a new storage provider
	sp := s.BaseSuite.CreateNewStorageProvider()

	successor_sp := s.StorageProviders[0]

	// 2, create a new gvg group for this storage provider
	var secondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != successor_sp.Info.Id {
			secondarySPIDs = append(secondarySPIDs, ssp.Info.Id)
		}
		if len(secondarySPIDs) == 6 {
			break
		}
	}

	gvgID, familyID := s.BaseSuite.CreateGlobalVirtualGroup(sp, 0, secondarySPIDs, 1)

	// 3. create object
	s.BaseSuite.CreateObject(sp, gvgID, storagetestutil.GenRandomBucketName(), storagetestutil.GenRandomObjectName())

	// 4. Create another gvg contains this new sp
	anotherSP := s.StorageProviders[1]
	var anotherSecondarySPIDs []uint32
	for _, ssp := range s.StorageProviders {
		if ssp.Info.Id != successor_sp.Info.Id {
			anotherSecondarySPIDs = append(anotherSecondarySPIDs, ssp.Info.Id)
		}
		if ssp.Info.Id != anotherSP.Info.Id {
			anotherSecondarySPIDs = append(anotherSecondarySPIDs, ssp.Info.Id)
		}
		if len(secondarySPIDs) == 5 {
			break
		}
	}
	anotherSecondarySPIDs = append(anotherSecondarySPIDs, sp.Info.Id)

	anotherSPsFamilies := s.queryGlobalVirtualGroupFamilies(anotherSP.Info.Id)
	s.Require().Greater(len(anotherSPsFamilies), 0)
	anotherGVGID, _ := s.BaseSuite.CreateGlobalVirtualGroup(&anotherSP, anotherSPsFamilies[0].Id, anotherSecondarySPIDs, 1)

	// 5. sp exit
	s.SendTxBlock(sp.OperatorKey, &virtualgroupmoduletypes.MsgStorageProviderExit{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
	})

	resp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(resp.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 6. sp complete exit failed
	s.SendTxBlockWithExpectErrorString(
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{OperatorAddress: sp.OperatorKey.GetAddr().String()},
		sp.OperatorKey,
		"not swap out from all the family")

	// 7. swapout, as primary sp
	msgSwapOut := virtualgroupmoduletypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), familyID, nil, successor_sp.Info.Id)
	msgSwapOut.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
	msgSwapOut.SuccessorSpApproval.Sig, err = successor_sp.ApprovalKey.Sign(msgSwapOut.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(sp.OperatorKey, msgSwapOut)

	// 8. exist failed
	s.SendTxBlockWithExpectErrorString(
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{OperatorAddress: sp.OperatorKey.GetAddr().String()},
		sp.OperatorKey,
		"not swap out from all the gvgs")

	// 7. swapout, as secondary sp
	msgSwapOut2 := virtualgroupmoduletypes.NewMsgSwapOut(sp.OperatorKey.GetAddr(), 0, []uint32{anotherGVGID}, successor_sp.Info.Id)
	msgSwapOut2.SuccessorSpApproval = &common.Approval{ExpiredHeight: math.MaxUint}
	msgSwapOut2.SuccessorSpApproval.Sig, err = successor_sp.ApprovalKey.Sign(msgSwapOut2.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(sp.OperatorKey, msgSwapOut2)

	// 8. sp complete exit success
	s.SendTxBlock(
		sp.OperatorKey,
		&virtualgroupmoduletypes.MsgCompleteStorageProviderExit{OperatorAddress: sp.OperatorKey.GetAddr().String()},
	)

}
