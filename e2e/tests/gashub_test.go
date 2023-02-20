package tests

import (
	"context"
	"strconv"
	"testing"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	grpctypes "github.com/cosmos/cosmos-sdk/types/grpc"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gashubtypes "github.com/cosmos/cosmos-sdk/x/gashub/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"

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
	msgUpdateGasParams := gashubtypes.NewMsgUpdateMsgGasParams(authtypes.NewModuleAddress(gov.ModuleName), []*gashubtypes.MsgGasParams{msgSendGasParams})
	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgUpdateGasParams},
		sdk.Coins{sdk.NewCoin(s.BaseSuite.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
		validator.String(),
		"test",
	)
	s.Require().NoError(err)

	res := s.SendTxBlock(msgProposal, s.Validator)
	s.Require().Equal(res.Code, uint32(0))

	// 2. query proposal
	var proposalId uint64
	for _, event := range res.Logs[0].Events {
		if event.Type == "submit_proposal" {
			for _, attr := range event.Attributes {
				if attr.Key == "proposal_id" {
					proposalId, err = strconv.ParseUint(attr.Value, 10, 0)
					s.Require().NoError(err)
				}
			}
		}
	}
	s.Require().True(proposalId != 0)

	queryProposal := &govtypesv1.QueryProposalRequest{ProposalId: proposalId}
	_, err = s.Client.GovQueryClientV1.Proposal(ctx, queryProposal)
	s.Require().NoError(err)

	// 3. submit MsgVote
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	res = s.SendTxBlock(msgVote, s.Validator)
	s.Require().Equal(res.Code, uint32(0))

	// 4. query new gas params
	var header metadata.MD
	queryRequest := &gashubtypes.QueryParamsRequest{}
	_, err = s.Client.GashubQueryClient.Params(ctx, queryRequest, grpc.Header(&header))
	s.Require().NoError(err)
	blockHeightHeader := header.Get(grpctypes.GRPCBlockHeightHeader)
	preHeight, err := strconv.ParseUint(blockHeightHeader[0], 10, 64)
	s.Require().NoError(err)
	// wait for the block end
	var queryRes *gashubtypes.QueryParamsResponse
	for {
		time.Sleep(10 * time.Second)
		queryRes, err = s.Client.GashubQueryClient.Params(ctx, queryRequest, grpc.Header(&header))
		s.Require().NoError(err)
		blockHeightHeader = header.Get(grpctypes.GRPCBlockHeightHeader)
		curHeight, err := strconv.ParseUint(blockHeightHeader[0], 10, 64)
		s.Require().NoError(err)
		if curHeight > preHeight {
			break
		}
	}

	s.Require().NoError(err)
	for _, params := range queryRes.GetParams().MsgGasParamsSet {
		if params.MsgTypeUrl == typeUrl {
			s.Require().True(params.GetFixedType().Equal(msgSendGasParams.GetFixedType()))
		}
	}
}

func TestGashubTestSuite(t *testing.T) {
	suite.Run(t, new(GashubTestSuite))
}
