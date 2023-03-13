package tests

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"math"
	"strconv"
	"strings"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/bits-and-blooms/bitset"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
	ctypes "github.com/tendermint/tendermint/rpc/core/types"

	"github.com/bnb-chain/greenfield/e2e/core"
	storagetestutil "github.com/bnb-chain/greenfield/testutil/storage"
	challengetypes "github.com/bnb-chain/greenfield/x/challenge/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type ChallengeTestSuite struct {
	core.BaseSuite
}

func (s *ChallengeTestSuite) SetupSuite() {
	s.BaseSuite.SetupSuite()
}

func (s *ChallengeTestSuite) SetupTest() {
}

func TestChallengeTestSuite(t *testing.T) {
	suite.Run(t, new(ChallengeTestSuite))
}

func (s *ChallengeTestSuite) createObject() (string, string, sdk.AccAddress, []sdk.AccAddress) {
	var err error
	sp := s.StorageProviders[0]
	// CreateBucket
	user := s.GenAndChargeAccounts(1, 1000000)[0]
	bucketName := "ch" + storagetestutil.GenRandomBucketName()
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, false, sp.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil)
	msgCreateBucket.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateBucket.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateBucket, user)

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
	msgCreateObject := storagetypes.NewMsgCreateObject(user.GetAddr(), bucketName, objectName, uint64(payloadSize), false, expectChecksum, contextType, storagetypes.REDUNDANCY_EC_TYPE, math.MaxUint, nil, nil)
	msgCreateObject.PrimarySpApproval.Sig, err = sp.ApprovalKey.GetPrivKey().Sign(msgCreateObject.GetApprovalBytes())
	s.Require().NoError(err)
	s.SendTxBlock(msgCreateObject, user)

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
	secondarySPs := []sdk.AccAddress{
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
		sp.OperatorKey.GetAddr(), sp.OperatorKey.GetAddr(),
	}
	msgSealObject := storagetypes.NewMsgSealObject(sp.SealKey.GetAddr(), bucketName, objectName, secondarySPs, nil)
	sr := storagetypes.NewSecondarySpSignDoc(sp.OperatorKey.GetAddr(), queryHeadObjectResponse.ObjectInfo.Id, checksum)
	secondarySig, err := sp.ApprovalKey.GetPrivKey().Sign(sr.GetSignBytes())
	s.Require().NoError(err)
	err = storagetypes.VerifySignature(sp.ApprovalKey.GetAddr(), sdk.Keccak256(sr.GetSignBytes()), secondarySig)
	s.Require().NoError(err)

	s.Require().NoError(err)

	secondarySigs := [][]byte{secondarySig, secondarySig, secondarySig, secondarySig, secondarySig, secondarySig}
	msgSealObject.SecondarySpSignatures = secondarySigs
	s.SendTxBlock(msgSealObject, sp.SealKey)

	queryHeadObjectResponse, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	return bucketName, objectName, sp.OperatorKey.GetAddr(), secondarySPs
}

func (s *ChallengeTestSuite) TestSubmit() {
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	bucketName, objectName, primarySp, _ := s.createObject()
	msgSubmit := challengetypes.NewMsgSubmit(user.GetAddr(), primarySp, bucketName, objectName, true, 1000)
	txRes := s.SendTxBlock(msgSubmit, user)
	event := filterEventFromTx(txRes) // secondary sps are faked with primary sp, redundancy check is meaningless here
	s.Require().GreaterOrEqual(event.ChallengeId, uint64(0))
	s.Require().NotEqual(event.SegmentIndex, uint32(100))
	s.Require().Equal(event.SpOperatorAddress, primarySp.String())

	bucketName, objectName, _, secondarySps := s.createObject()
	msgSubmit = challengetypes.NewMsgSubmit(user.GetAddr(), secondarySps[0], bucketName, objectName, false, 0)
	txRes = s.SendTxBlock(msgSubmit, user)
	event = filterEventFromTx(txRes)
	s.Require().GreaterOrEqual(event.ChallengeId, uint64(0))
	s.Require().Equal(event.SegmentIndex, uint32(0))
	s.Require().Equal(event.SpOperatorAddress, secondarySps[0].String())
}

func (s *ChallengeTestSuite) calculateValidatorBitSet(height int64, relayerKey string) *bitset.BitSet {
	valBitSet := bitset.New(256)

	page := 1
	size := 10
	valRes, err := s.TmClient.TmClient.Validators(context.Background(), &height, &page, &size)
	if err != nil {
		panic(err)
	}

	for idx, val := range valRes.Validators {
		if strings.EqualFold(relayerKey, hex.EncodeToString(val.RelayerBlsKey[:])) {
			valBitSet.Set(uint(idx))
		}
	}

	return valBitSet
}

func (s *ChallengeTestSuite) TestNormalAttest() {
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	bucketName, objectName, primarySp, _ := s.createObject()
	msgSubmit := challengetypes.NewMsgSubmit(user.GetAddr(), primarySp, bucketName, objectName, true, 1000)
	txRes := s.SendTxBlock(msgSubmit, user)
	event := filterEventFromTx(txRes)

	statusRes, err := s.TmClient.TmClient.Status(context.Background())
	s.Require().NoError(err)
	height := statusRes.SyncInfo.LatestBlockHeight

	valBitset := s.calculateValidatorBitSet(height, s.Relayer.GetPrivKey().PubKey().String())

	msgAttest := challengetypes.NewMsgAttest(user.GetAddr(), event.ChallengeId, event.ObjectId, primarySp.String(),
		challengetypes.CHALLENGE_SUCCEED, user.GetAddr().String(), valBitset.Bytes(), nil)
	toSign := msgAttest.GetBlsSignBytes()

	voteAggSignature, err := s.Relayer.GetPrivKey().Sign(toSign[:])
	if err != nil {
		panic(err)
	}
	msgAttest.VoteAggSignature = voteAggSignature

	txRes = s.SendTxBlock(msgAttest, user)
	s.Require().True(txRes.Code == 0)

	queryRes, err := s.Client.ChallengeQueryClient.LatestAttestedChallenge(context.Background(), &challengetypes.QueryLatestAttestedChallengeRequest{})
	s.Require().NoError(err)
	s.Require().True(queryRes.ChallengeId == event.ChallengeId)
}

func (s *ChallengeTestSuite) TestHeartbeatAttest() {
	for i := 0; i < 3; i++ {
		s.createObject()
	}

	heartbeatInterval := uint64(100)
	user := s.GenAndChargeAccounts(1, 1000000)[0]

	var event challengetypes.EventStartChallenge
	found := false
	height := int64(0)
	for {
		statusRes, err := s.TmClient.TmClient.Status(context.Background())
		s.Require().NoError(err)
		height = statusRes.SyncInfo.LatestBlockHeight

		time.Sleep(10 * time.Millisecond)
		blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &height)
		s.Require().NoError(err)
		events := filterEventFromBlock(blockRes)

		for _, e := range events {
			if e.ChallengeId%heartbeatInterval == 0 {
				event = e
				found = true
				break
			}
		}
		if found == true {
			break
		}

		if len(events) > 0 {
			s.T().Logf("current challenge id: %d", events[len(events)-1].ChallengeId)
		}
		time.Sleep(1 * time.Second)
	}

	valBitset := s.calculateValidatorBitSet(height, s.Relayer.GetPrivKey().PubKey().String())

	msgAttest := challengetypes.NewMsgAttest(user.GetAddr(), event.ChallengeId, event.ObjectId,
		event.SpOperatorAddress, challengetypes.CHALLENGE_FAILED, "", valBitset.Bytes(), nil)
	toSign := msgAttest.GetBlsSignBytes()

	voteAggSignature, err := s.Relayer.GetPrivKey().Sign(toSign[:])
	if err != nil {
		panic(err)
	}
	msgAttest.VoteAggSignature = voteAggSignature

	txRes := s.SendTxBlock(msgAttest, user)
	s.Require().True(txRes.Code == 0)

	queryRes, err := s.Client.ChallengeQueryClient.LatestAttestedChallenge(context.Background(), &challengetypes.QueryLatestAttestedChallengeRequest{})
	s.Require().NoError(err)
	s.Require().True(queryRes.ChallengeId == event.ChallengeId)
}

func (s *ChallengeTestSuite) TestEndBlock() {
	for i := 0; i < 3; i++ {
		s.createObject()
	}

	statusRes, err := s.TmClient.TmClient.Status(context.Background())
	s.Require().NoError(err)
	height := statusRes.SyncInfo.LatestBlockHeight

	blockRes, err := s.TmClient.TmClient.BlockResults(context.Background(), &height)
	s.Require().NoError(err)
	events := filterEventFromBlock(blockRes)
	s.Require().True(len(events) > 0)
}

func filterEventFromBlock(blockRes *ctypes.ResultBlockResults) []challengetypes.EventStartChallenge {
	challengeEvents := make([]challengetypes.EventStartChallenge, 0)

	for _, event := range blockRes.EndBlockEvents {
		if event.Type == "bnbchain.greenfield.challenge.EventStartChallenge" {

			challengeIdStr, objectIdStr, redundancyIndexStr, segmentIndexStr, spOpAddress := "", "", "", "", ""
			for _, attr := range event.Attributes {
				if string(attr.Key) == "challenge_id" {
					challengeIdStr = strings.Trim(string(attr.Value), `"`)
				} else if string(attr.Key) == "object_id" {
					objectIdStr = strings.Trim(string(attr.Value), `"`)
				} else if string(attr.Key) == "redundancy_index" {
					redundancyIndexStr = strings.Trim(string(attr.Value), `"`)
				} else if string(attr.Key) == "segment_index" {
					segmentIndexStr = strings.Trim(string(attr.Value), `"`)
				} else if string(attr.Key) == "sp_operator_address" {
					spOpAddress = strings.Trim(string(attr.Value), `"`)
				}
			}
			challengeId, _ := strconv.ParseInt(challengeIdStr, 10, 64)
			objectId := sdkmath.NewUintFromString(objectIdStr)
			redundancyIndex, _ := strconv.ParseInt(redundancyIndexStr, 10, 32)
			segmentIndex, _ := strconv.ParseInt(segmentIndexStr, 10, 32)
			challengeEvents = append(challengeEvents, challengetypes.EventStartChallenge{
				ChallengeId:       uint64(challengeId),
				ObjectId:          objectId,
				SegmentIndex:      uint32(segmentIndex),
				SpOperatorAddress: spOpAddress,
				RedundancyIndex:   int32(redundancyIndex),
			})
		}
	}
	return challengeEvents
}

func filterEventFromTx(txRes *sdk.TxResponse) challengetypes.EventStartChallenge {
	challengeIdStr, objectIdStr, redundancyIndexStr, segmentIndexStr, spOpAddress := "", "", "", "", ""
	for _, event := range txRes.Logs[0].Events {
		if event.Type == "bnbchain.greenfield.challenge.EventStartChallenge" {
			for _, attr := range event.Attributes {
				if attr.Key == "challenge_id" {
					challengeIdStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "object_id" {
					objectIdStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "redundancy_index" {
					redundancyIndexStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "segment_index" {
					segmentIndexStr = strings.Trim(attr.Value, `"`)
				} else if attr.Key == "sp_operator_address" {
					spOpAddress = strings.Trim(attr.Value, `"`)
				}
			}
		}
	}
	challengeId, _ := strconv.ParseInt(challengeIdStr, 10, 64)
	objectId := sdkmath.NewUintFromString(objectIdStr)
	redundancyIndex, _ := strconv.ParseInt(redundancyIndexStr, 10, 32)
	segmentIndex, _ := strconv.ParseInt(segmentIndexStr, 10, 32)
	return challengetypes.EventStartChallenge{
		ChallengeId:       uint64(challengeId),
		ObjectId:          objectId,
		SegmentIndex:      uint32(segmentIndex),
		SpOperatorAddress: spOpAddress,
		RedundancyIndex:   int32(redundancyIndex),
	}
}
