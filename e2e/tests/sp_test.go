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
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type StorageProviderTestSuite struct {
	core.BaseSuite
}

func (s *StorageProviderTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *StorageProviderTestSuite) SetupTest() {
}

func (s *StorageProviderTestSuite) TestCreateStorageProvider() {
	// Create a New SP
	sp := s.BaseSuite.CreateNewStorageProvider()

	// query sp by id
	querySPResp, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{
		Id: sp.Info.Id,
	})
	s.Require().NoError(err)
	s.Require().Equal(querySPResp.StorageProvider, querySPResp.StorageProvider)

	// sp exit
	msgSPExit := virtualgroupmoduletypes.MsgStorageProviderExit{
		StorageProvider: sp.OperatorKey.GetAddr().String(),
	}
	s.SendTxBlock(sp.OperatorKey, &msgSPExit)

	// 9 query sp status
	querySPResp2, err := s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
	s.Require().NoError(err)
	s.Require().Equal(querySPResp2.StorageProvider.Status, sptypes.STATUS_GRACEFUL_EXITING)

	// 10 complete sp exit
	msgCompleteSPExit := virtualgroupmoduletypes.MsgCompleteStorageProviderExit{
		StorageProvider: sp.OperatorKey.GetAddr().String(),
	}

	s.SendTxBlock(sp.OperatorKey, &msgCompleteSPExit)

	// 10 query sp
	_, err = s.Client.StorageProvider(context.Background(), &sptypes.QueryStorageProviderRequest{Id: sp.Info.Id})
	s.Require().Error(err)
}

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
		OperatorAddress: prevSP.OperatorAddress,
		FundingAddress:  prevSP.FundingAddress,
		SealAddress:     prevSP.SealAddress,
		ApprovalAddress: prevSP.ApprovalAddress,
		GcAddress:       prevSP.GcAddress,
		BlsKey:          newBlsPubKeyBz,
		Description: sptypes.Description{
			Moniker:  "sp_test_edit",
			Identity: "",
		},
		Endpoint:     "http://127.0.0.1:9034",
		TotalDeposit: prevSP.TotalDeposit,
	}
	msgEditSP := sptypes.NewMsgEditStorageProvider(
		sp.OperatorKey.GetAddr(), newSP.Endpoint, &newSP.Description,
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr(),
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
		sp.SealKey.GetAddr(), sp.ApprovalKey.GetAddr(), sp.GcKey.GetAddr(),
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

func (s *StorageProviderTestSuite) TestSpStoragePrice() {
	ctx := context.Background()
	s.CheckSecondarySpPrice()
	sp := s.BaseSuite.PickStorageProvider()
	spAddr := sp.OperatorKey.GetAddr().String()
	spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log(spStoragePrice)
	// update storage price
	newReadPrice := sdk.NewDec(core.RandInt64(100, 200))
	newStorePrice := sdk.NewDec(core.RandInt64(10000, 20000))
	msgUpdateSpStoragePrice := &sptypes.MsgUpdateSpStoragePrice{
		SpAddress:     spAddr,
		ReadPrice:     newReadPrice,
		StorePrice:    newStorePrice,
		FreeReadQuota: spStoragePrice.SpStoragePrice.FreeReadQuota,
	}
	_ = s.SendTxBlock(sp.OperatorKey, msgUpdateSpStoragePrice)
	// query and assert
	spStoragePrice2, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Log(spStoragePrice2)
	// check price changed as expected
	s.Require().Equal(newReadPrice, spStoragePrice2.SpStoragePrice.ReadPrice)
	s.Require().Equal(newStorePrice, spStoragePrice2.SpStoragePrice.StorePrice)
	s.CheckSecondarySpPrice()
	// query sp storage price by time before it exists, expect error
	_, err = s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: 1,
	})
	s.Require().Error(err)
	_, err = s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
		SpAddr:    spAddr,
		Timestamp: 1,
	})
	s.Require().Error(err)
}

func (s *StorageProviderTestSuite) CheckSecondarySpPrice() {
	ctx := context.Background()
	queryGetSecondarySpStorePriceByTimeResp, err := s.Client.QueryGetSecondarySpStorePriceByTime(ctx, &sptypes.QueryGetSecondarySpStorePriceByTimeRequest{
		Timestamp: 0,
	})
	s.Require().NoError(err)
	s.T().Logf("Secondary SP store price: %s", core.YamlString(queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice))
	// query all sps
	sps, err := s.Client.StorageProviders(ctx, &sptypes.QueryStorageProvidersRequest{})
	s.Require().NoError(err)
	s.T().Logf("sps: %s", sps)
	spNum := int64(sps.Pagination.Total)
	prices := make([]sdk.Dec, 0)
	for _, sp := range sps.Sps {
		spStoragePrice, err := s.Client.QueryGetSpStoragePriceByTime(ctx, &sptypes.QueryGetSpStoragePriceByTimeRequest{
			SpAddr:    sp.OperatorAddress,
			Timestamp: 0,
		})
		s.Require().NoError(err)
		s.T().Logf("sp: %s, storage price: %s", sp.OperatorAddress, core.YamlString(spStoragePrice.SpStoragePrice))
		prices = append(prices, spStoragePrice.SpStoragePrice.StorePrice)
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].LT(prices[j]) })
	var median sdk.Dec
	if spNum%2 == 0 {
		median = prices[spNum/2-1].Add(prices[spNum/2]).QuoInt64(2)
	} else {
		median = prices[spNum/2]
	}

	params, err := s.Client.SpQueryClient.Params(ctx, &sptypes.QueryParamsRequest{})
	s.Require().NoError(err)
	expectedSecondarySpStorePrice := params.Params.SecondarySpStorePriceRatio.Mul(median)
	s.Require().Equal(expectedSecondarySpStorePrice, queryGetSecondarySpStorePriceByTimeResp.SecondarySpStorePrice.StorePrice)
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
