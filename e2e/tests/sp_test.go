package tests

import (
	"context"
	"encoding/hex"
	"math/big"
	"reflect"
	"sort"
	"strconv"
	"testing"
	"time"

	"github.com/cometbft/cometbft/crypto/tmhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	"github.com/bnb-chain/greenfield/testutil/sample"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
)

type StorageProviderTestSuite struct {
	core.BaseSuite
	defaultParams sptypes.Params
}

func (s *StorageProviderTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
	s.defaultParams = s.queryParams()
}

func (s *StorageProviderTestSuite) SetupTest() {
}

//func (s *StorageProviderTestSuite) TestCreateStorageProvider() {
//	// Create a New SP
//	sp := s.BaseSuite.CreateNewStorageProvider()
//
//	// query sp by id
//	querySPResp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{
//		Id: sp.Info.Id,
//	})
//	s.Require().NoError(err)
//	s.Require().Equal(querySPResp.StorageProvider, querySPResp.StorageProvider)
//
//	// sp exit
//	msgSPExit := virtualgroupmoduletypes.MsgStorageProviderExit{
//		StorageProvider: sp.OperatorKey.GetAddr().String(),
//	}
//	s.SendTxBlock(sp.OperatorKey, &msgSPExit)
//
//	// 9 query sp status
//	querySPResp2, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
//	s.Require().NoError(err)
//	s.Require().Equal(querySPResp2.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)
//
//	// 10 complete sp exit
//	msgCompleteSPExit := virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
//		StorageProvider: sp.OperatorKey.GetAddr().String(),
//	}
//
//	s.SendTxBlock(sp.OperatorKey, &msgCompleteSPExit)
//
//	// 10 query sp
//	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
//	s.Require().Error(err)
//}

func (s *StorageProviderTestSuite) TestEditStorageProvider() {
	ctx := context.Background()
	sp := s.BaseSuite.PickStorageProvider()
	blsProof, _ := sp.BlsKey.Sign(tmhash.Sum(sp.BlsKey.PubKey().Bytes()))

	// 1. query previous storage provider
	querySPByOperatorAddressReq := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: sp.OperatorKey.GetAddr().String(),
	}

	querySPByOperatorAddressResp, err := s.Client.StorageProviderByOperatorAddress(ctx, &querySPByOperatorAddressReq)
	s.Require().NoError(err)
	prevSP := querySPByOperatorAddressResp.StorageProvider

	// 2. edit storage provider
	newBlsPubKeyBz, newBlsProofBz := sample.RandBlsPubKeyAndBlsProofBz()

	newSP := &sptypes.StorageProvider{
		OperatorAddress:    prevSP.OperatorAddress,
		FundingAddress:     prevSP.FundingAddress,
		SealAddress:        prevSP.SealAddress,
		ApprovalAddress:    prevSP.ApprovalAddress,
		GcAddress:          prevSP.GcAddress,
		MaintenanceAddress: prevSP.MaintenanceAddress,
		BlsKey:             newBlsPubKeyBz,
		Description: sptypes.Description{
			Moniker:  "sp_test_edit",
			Identity: "",
		},
		Endpoint:     "http://127.0.0.1:9034",
		TotalDeposit: prevSP.TotalDeposit,
	}
	msgEditSP := sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), newSP.Endpoint, &newSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr(), sp.MaintenanceKey.GetAddr(),
		hex.EncodeToString(newBlsPubKeyBz),
		hex.EncodeToString(newBlsProofBz),
	)

	txRes := s.SendTxBlock(sp.OperatorKey, msgEditSP)
	s.Require().Equal(txRes.Code, uint32(0))

	// 3. query modified storage provider
	querySPReq := sptypes.QueryStorageProviderRequest{
		Id: sp.Info.Id,
	}

	querySPResp, err := s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	newSP.Id = querySPResp.StorageProvider.Id
	s.Require().Equal(querySPResp.StorageProvider, newSP)

	// 4. revert storage provider info
	msgEditSP = sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), prevSP.Endpoint, &prevSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr(), sp.MaintenanceKey.GetAddr(),
		hex.EncodeToString(prevSP.BlsKey), hex.EncodeToString(blsProof))

	txRes = s.SendTxBlock(sp.OperatorKey, msgEditSP)
	s.Require().Equal(txRes.Code, uint32(0))

	// 5. query revert storage provider again
	querySPReq = sptypes.QueryStorageProviderRequest{
		Id: sp.Info.Id,
	}

	querySPResp, err = s.Client.StorageProvider(ctx, &querySPReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, prevSP)
}

func (s *StorageProviderTestSuite) TestDeposit() {
	sp := s.BaseSuite.PickStorageProvider()

	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}

	msgDeposit := sptypes.NewMsgDeposit(
		sp.FundingKey.GetAddr(), sp.OperatorKey.GetAddr(), deposit)
	txRes := s.SendTxBlock(sp.FundingKey, msgDeposit)
	s.Require().Equal(txRes.Code, uint32(0))
}

func (s *StorageProviderTestSuite) TestUpdateSpStoragePrice() {
	ctx := context.Background()
	defer s.revertParams()

	// query sp storage price by time before it exists, expect error
	_, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 1,
	})
	s.Require().Error(err)

	// update params
	params := s.queryParams()
	params.UpdateGlobalPriceInterval = 5
	s.updateParams(params)

	sp := s.BaseSuite.PickStorageProvider()
	spAddr := sp.OperatorKey.GetAddr().String()
	spStoragePrice, err := s.Client.QuerySpStoragePrice(ctx, &sptypes.QuerySpStoragePriceRequest{
		SpAddr: spAddr,
	})
	s.Require().NoError(err)
	s.T().Log(spStoragePrice)

	// update storage price - update is ok
	msgUpdateSpStoragePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     spAddr,
		ReadPrice:     spStoragePrice.SpStoragePrice.ReadPrice,
		StorePrice:    spStoragePrice.SpStoragePrice.StorePrice,
		FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
	}
	_ = s.SendTxBlock(sp.OperatorKey, msgUpdateSpStoragePrice)

	time.Sleep(6 * time.Second)

	// verify price is updated after interval
	globalPriceResBefore, _ := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{Timestamp: 0})
	s.T().Log("globalPriceResBefore", core.YamlString(globalPriceResBefore))
	priceChanged := false
	globalPriceResAfter1, _ := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{Timestamp: 0})
	s.T().Log("globalPriceResAfter1", core.YamlString(globalPriceResAfter1))
	for _, sp := range s.BaseSuite.StorageProviders {
		msgUpdateSpStoragePrice = &sptypes.MsgUpdateSpStoragePrice{
			SpAddress:     sp.OperatorKey.GetAddr().String(),
			ReadPrice:     spStoragePrice.SpStoragePrice.ReadPrice.MulInt64(10),
			StorePrice:    spStoragePrice.SpStoragePrice.StorePrice.MulInt64(10),
			FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
		}
		s.SendTxBlock(sp.OperatorKey, msgUpdateSpStoragePrice)

		globalPriceResAfter1, _ = s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{Timestamp: 0})
		s.T().Log("globalPriceResAfter1", core.YamlString(globalPriceResAfter1))
		if !globalPriceResAfter1.GlobalSpStorePrice.PrimaryStorePrice.Equal(globalPriceResBefore.GlobalSpStorePrice.PrimaryStorePrice) {
			_ = s.CheckGlobalSpStorePrice()
			priceChanged = true
			break
		}
	}

	time.Sleep(6 * time.Second)
	globalPriceResAfter2, _ := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{Timestamp: 0})
	s.T().Log("globalPriceResAfter2", core.YamlString(globalPriceResAfter2))

	checked := s.CheckGlobalSpStorePrice()
	s.Require().True(checked)

	if !priceChanged { //if price not changed, then after 6 seconds, it should change
		s.Require().NotEqual(globalPriceResAfter2.GlobalSpStorePrice.PrimaryStorePrice, globalPriceResBefore.GlobalSpStorePrice.PrimaryStorePrice)
		s.Require().NotEqual(globalPriceResAfter2.GlobalSpStorePrice.SecondaryStorePrice, globalPriceResBefore.GlobalSpStorePrice.SecondaryStorePrice)
		s.Require().NotEqual(globalPriceResAfter2.GlobalSpStorePrice.ReadPrice, globalPriceResBefore.GlobalSpStorePrice.ReadPrice)
	} else { //if price not changed already, then after 6 seconds, it should not change
		s.Require().Equal(globalPriceResAfter2.GlobalSpStorePrice.PrimaryStorePrice, globalPriceResAfter1.GlobalSpStorePrice.PrimaryStorePrice)
		s.Require().Equal(globalPriceResAfter2.GlobalSpStorePrice.SecondaryStorePrice, globalPriceResAfter1.GlobalSpStorePrice.SecondaryStorePrice)
		s.Require().Equal(globalPriceResAfter2.GlobalSpStorePrice.ReadPrice, globalPriceResAfter1.GlobalSpStorePrice.ReadPrice)
	}

	// update params
	now := time.Now().UTC()
	_, _, day := now.Date()
	params = s.queryParams()
	params.UpdateGlobalPriceInterval = 0 // update by month
	params.UpdatePriceDisallowedDays = uint32(31 - day + 1)
	s.updateParams(params)

	// update storage price - third update is not ok
	msgUpdateSpStoragePrice = &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     spAddr,
		ReadPrice:     spStoragePrice.SpStoragePrice.ReadPrice,
		StorePrice:    spStoragePrice.SpStoragePrice.StorePrice,
		FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
	}
	s.SendTxBlockWithExpectErrorString(msgUpdateSpStoragePrice, sp.OperatorKey, "update price is disallowed")
}

func (s *StorageProviderTestSuite) CheckGlobalSpStorePrice() bool {
	ctx := context.Background()
	queryGlobalSpStorePriceByTimeResp, err := s.Client.QueryGlobalSpStorePriceByTime(ctx, &sptypes.QueryGlobalSpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Logf("global SP store price: %s", core.YamlString(queryGlobalSpStorePriceByTimeResp.GlobalSpStorePrice))
	// query all sps
	sps, err := s.Client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{})
	s.Require().NoError(err)
	s.T().Logf("sps: %s", sps)
	spNum := int64(sps.Pagination.Total)
	storePrices := make([]sdk.Dec, 0)
	readPrices := make([]sdk.Dec, 0)
	for _, sp := range sps.Sps {
		if sp.Status == sptypes.STATUS_IN_SERVICE || sp.Status == sptypes.STATUS_IN_MAINTENANCE {
			spStoragePrice, err := s.Client.QuerySpStoragePrice(ctx, &sptypes.QuerySpStoragePriceRequest{
				SpAddr: sp.OperatorAddress,
			})
			s.Require().NoError(err)
			s.T().Logf("sp: %s, storage price: %s", sp.OperatorAddress, core.YamlString(spStoragePrice.SpStoragePrice))

			if spStoragePrice.SpStoragePrice.UpdateTimeSec >= queryGlobalSpStorePriceByTimeResp.GlobalSpStorePrice.UpdateTimeSec {
				s.T().Logf("cannot do the calculation for there is a new price update for %s", sp.OperatorAddress)
				return false
			}

			storePrices = append(storePrices, spStoragePrice.SpStoragePrice.StorePrice)
			readPrices = append(readPrices, spStoragePrice.SpStoragePrice.ReadPrice)
		}
	}

	sort.Slice(storePrices, func(i, j int) bool { return storePrices[i].LT(storePrices[j]) })
	var storeMedian sdk.Dec
	if spNum%2 == 0 {
		storeMedian = storePrices[spNum/2-1].Add(storePrices[spNum/2]).QuoInt64(2)
	} else {
		storeMedian = storePrices[spNum/2]
	}

	sort.Slice(readPrices, func(i, j int) bool { return readPrices[i].LT(readPrices[j]) })
	var readMedian sdk.Dec
	if spNum%2 == 0 {
		readMedian = readPrices[spNum/2-1].Add(readPrices[spNum/2]).QuoInt64(2)
	} else {
		readMedian = readPrices[spNum/2]
	}

	s.Require().Equal(storeMedian, queryGlobalSpStorePriceByTimeResp.GlobalSpStorePrice.PrimaryStorePrice)
	params, err := s.Client.SpQueryClient.Params(ctx, &sptypes.QueryParamsRequest{})
	s.Require().NoError(err)
	expectedSecondarySpStorePrice := params.Params.SecondarySpStorePriceRatio.Mul(storeMedian)
	s.Require().Equal(expectedSecondarySpStorePrice, queryGlobalSpStorePriceByTimeResp.GlobalSpStorePrice.SecondaryStorePrice)
	s.Require().Equal(readMedian, queryGlobalSpStorePriceByTimeResp.GlobalSpStorePrice.ReadPrice)
	return true
}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}

func (s *StorageProviderTestSuite) TestUpdateStorageProviderParams() {
	// 1. create proposal
	govAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()
	queryParamsResp, err := s.Client.SpQueryClient.Params(context.Background(), &sptypes.QueryParamsRequest{})
	s.Require().NoError(err)

	updatedParams := queryParamsResp.Params
	updatedParams.SecondarySpStorePriceRatio = sdk.NewDecFromBigIntWithPrec(big.NewInt(1), 18)
	msgUpdateParams := &sptypes.MsgUpdateParams{
		Authority: govAddr,
		Params:    updatedParams,
	}

	proposal, err := v1.NewMsgSubmitProposal([]sdk.Msg{msgUpdateParams}, sdk.NewCoins(sdk.NewCoin("BNB", sdk.NewInt(1000000000000000000))),
		s.Validator.GetAddr().String(), "", "update StorageProvider params", "Test update StorageProvider params")
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

	updatedQueryParamsResp, err := s.Client.SpQueryClient.Params(context.Background(), &sptypes.QueryParamsRequest{})
	s.Require().NoError(err)
	if reflect.DeepEqual(updatedQueryParamsResp.Params, updatedParams) {
		s.T().Logf("update params success")
	} else {
		s.T().Errorf("update params failed")
	}
}

func (s *StorageProviderTestSuite) TestUpdateStorageProviderStatus() {
	ctx := context.Background()
	var sp *core.StorageProvider
	for _, tempSP := range s.BaseSuite.StorageProviders {
		exists, err := s.BaseSuite.ExistsSPMaintenanceRecords(tempSP.OperatorKey.GetAddr().String())
		s.Require().NoError(err)
		if !exists {
			sp = tempSP
			break
		}
	}
	operatorAddr := sp.OperatorKey.GetAddr()

	// 1. query storage provider
	req := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: operatorAddr.String(),
	}
	spResp, err := s.Client.StorageProviderByOperatorAddress(ctx, &req)
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_IN_SERVICE, spResp.StorageProvider.Status)

	msg := sptypes.NewMsgUpdateStorageProviderStatus(
		operatorAddr,
		sptypes.STATUS_IN_MAINTENANCE,
		120, // seconds
	)
	txRes := s.SendTxBlock(sp.OperatorKey, msg)
	s.Require().Equal(txRes.Code, uint32(0))

	spResp, err = s.Client.StorageProviderByOperatorAddress(ctx, &req)
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_IN_MAINTENANCE, spResp.StorageProvider.Status)

	msg = sptypes.NewMsgUpdateStorageProviderStatus(
		operatorAddr,
		sptypes.STATUS_IN_SERVICE,
		0,
	)

	txRes = s.SendTxBlock(sp.OperatorKey, msg)
	s.Require().Equal(txRes.Code, uint32(0))
	spResp, err = s.Client.StorageProviderByOperatorAddress(ctx, &req)
	s.Require().NoError(err)
	s.Require().Equal(sptypes.STATUS_IN_SERVICE, spResp.StorageProvider.Status)
}

func (s *StorageProviderTestSuite) queryParams() sptypes.Params {
	queryParamsRequest := sptypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.SpQueryClient.Params(context.Background(), &queryParamsRequest)
	s.Require().NoError(err)
	s.T().Log("params", core.YamlString(queryParamsResponse.Params))
	return queryParamsResponse.Params
}

func (s *StorageProviderTestSuite) revertParams() {
	s.updateParams(s.defaultParams)
}

func (s *StorageProviderTestSuite) updateParams(params sptypes.Params) {
	var err error
	validator := s.Validator.GetAddr()

	ctx := context.Background()

	queryParamsRequest := &sptypes.QueryParamsRequest{}
	queryParamsResponse, err := s.Client.SpQueryClient.Params(ctx, queryParamsRequest)
	s.Require().NoError(err)
	s.T().Log("params before", core.YamlString(queryParamsResponse.Params))

	msgUpdateParams := &sptypes.MsgUpdateParams{
		Authority: authtypes.NewModuleAddress(govtypes.ModuleName).String(),
		Params:    params,
	}

	msgProposal, err := v1.NewMsgSubmitProposal(
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

	queryProposal := &v1.QueryProposalRequest{ProposalId: proposalId}
	_, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote := v1.NewMsgVote(validator, proposalId, v1.OptionYes, "test")
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := v1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(ctx, &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(1 * time.Second)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, v1.ProposalStatus_PROPOSAL_STATUS_PASSED)

	queryParamsRequest = &sptypes.QueryParamsRequest{}
	queryParamsResponse, err = s.Client.SpQueryClient.Params(ctx, queryParamsRequest)
	s.Require().NoError(err)
	s.T().Log("params after", core.YamlString(queryParamsResponse.Params))
}
