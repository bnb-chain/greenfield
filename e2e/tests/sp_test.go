package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"
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
	deposit := sdk.Coin{
		Denom:  "bnb",
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}
	description := sptypes.Description{
		Moniker:  "sp0",
		Identity: "",
	}
	// CreateStorageProvider
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	msgCreateStorageProvider, err := sptypes.NewMsgCreateStorageProvider(sdk.AccAddress("0x7b5Fe22B5446f7C62Ea27B8BD71CeF94e03f3dF2"), s.StorageProvider.OperatorKey.GetAddr(), s.StorageProvider.FundingKey.GetAddr(),
		s.StorageProvider.SealKey.GetAddr(),
		s.StorageProvider.ApprovalKey.GetAddr(), description,
		"sp0.greenfield.io", deposit)
	//msgCreateSP

	s.Require().NoError(err)
	s.SendTxBlock(msgCreateStorageProvider, user)

	ctx := context.Background()
	validator := s.Validator.GetAddr()

	// 1. submit CreateStorageProviderParams
	typeUrl := sdk.MsgTypeURL(&banktypes.MsgSend{})
	msgSendGasParams := gashubtypes.NewMsgGasParamsWithFixedGas(typeUrl, 1e6)
	msgUpdateGasParams := gashubtypes.NewMsgUpdateMsgGasParams(authtypes.NewModuleAddress(gov.ModuleName), []*gashubtypes.MsgGasParams{msgSendGasParams})
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateGasParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(msgProposal, s.Validator)
	s.Require().Equal(txRes.Code, uint32(0))

	// 2. query proposal
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

	// 3. submit MsgVote and wait the proposal exec
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(msgVote, s.Validator)
	s.Require().Equal(txRes.Code, uint32(0))

	for {
		time.Sleep(60 * time.Second)
		proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
		s.Require().NoError(err)
		if proposalRes.Proposal.Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		}
	}

	// 4. query new gas params
	queryRequest := &gashubtypes.QueryParamsRequest{}
	queryRes, err := s.Client.GashubQueryClient.Params(ctx, queryRequest)
	s.Require().NoError(err)

	for _, params := range queryRes.GetParams().MsgGasParamsSet {
		if params.MsgTypeUrl == typeUrl {
			s.Require().True(params.GetFixedType().Equal(msgSendGasParams.GetFixedType()))
		}
	}
}

func TestStorageProviderTestSuite(t *testing.T) {
	suite.Run(t, new(StorageProviderTestSuite))
}
