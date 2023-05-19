package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/e2e/core"
	"github.com/bnb-chain/greenfield/sdk/types"
)

type GashubTestSuite struct {
	core.BaseSuite
}

func (s *GashubTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *GashubTestSuite) SetupTest() {}

func (s *GashubTestSuite) TestUpdateParams() {
	ctx := context.Background()
	validator := s.Validator.GetAddr()

	// 1. submit MsgUpdateMsgGasParams
	typeUrl := sdk.MsgTypeURL(&banktypes.MsgSend{})
	msgSendGasParams := gashubtypes.NewMsgGasParamsWithFixedGas(typeUrl, 1e6)
	msgUpdateGasParams := gashubtypes.NewMsgSetMsgGasParams(authtypes.NewModuleAddress(gov.ModuleName).String(), []*gashubtypes.MsgGasParams{msgSendGasParams}, nil)
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateGasParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test",
		"update gas params",
		"pdate gas params",
	)
	s.Require().NoError(err)

	txRes := s.SendTxBlock(s.Validator, msgProposal)
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
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	for {
		proposalRes, err := s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
		s.Require().NoError(err)
		if proposalRes.Proposal.Status == govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED {
			break
		}
		s.T().Logf("waiting for proposal to be passed, now: %s", time.Now())
		time.Sleep(time.Second)
	}

	// 4. query new gas params
	queryRequest := &gashubtypes.QueryMsgGasParamsRequest{}
	queryRes, err := s.Client.GashubQueryClient.MsgGasParams(ctx, queryRequest)
	s.Require().NoError(err)

	for _, params := range queryRes.GetMsgGasParams() {
		if params.MsgTypeUrl == typeUrl {
			s.Require().True(params.GetFixedType().Equal(msgSendGasParams.GetFixedType()))
		}
	}
}

func TestGashubTestSuite(t *testing.T) {
	suite.Run(t, new(GashubTestSuite))
}
