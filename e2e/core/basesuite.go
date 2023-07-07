package core

import (
	"bytes"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/cometbft/cometbft/crypto/tmhash"
	"math"
	"strconv"
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	gov "github.com/cosmos/cosmos-sdk/x/gov/types"
	govtypesv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/prysmaticlabs/prysm/crypto/bls"
	"github.com/stretchr/testify/suite"

	"github.com/bnb-chain/greenfield/sdk/client"
	"github.com/bnb-chain/greenfield/sdk/keys"
	"github.com/bnb-chain/greenfield/sdk/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
	virtualgroupmoduletypes "github.com/bnb-chain/greenfield/x/virtualgroup/types"
)

type StorageProvider struct {
	OperatorKey                keys.KeyManager
	SealKey                    keys.KeyManager
	FundingKey                 keys.KeyManager
	ApprovalKey                keys.KeyManager
	GcKey                      keys.KeyManager
	BlsKey                     keys.KeyManager
	Info                       *sptypes.StorageProvider
	GlobalVirtualGroupFamilies map[uint32][]*virtualgroupmoduletypes.GlobalVirtualGroup
}

type BaseSuite struct {
	suite.Suite
	Config           *Config
	Client           *client.GreenfieldClient
	TmClient         *client.TendermintClient
	Validator        keys.KeyManager
	ValidatorBLS     keys.KeyManager
	Relayer          keys.KeyManager
	Challenger       keys.KeyManager
	StorageProviders []StorageProvider
}

func (s *BaseSuite) SetupSuite() {
	s.Config = InitConfig()
	s.Client, _ = client.NewGreenfieldClient(s.Config.TendermintAddr, s.Config.ChainId)
	tmClient := client.NewTendermintClient(s.Config.TendermintAddr)
	s.TmClient = &tmClient
	var err error
	s.Validator, err = keys.NewMnemonicKeyManager(s.Config.ValidatorMnemonic)
	s.Require().NoError(err)
	s.ValidatorBLS, err = keys.NewBlsMnemonicKeyManager(s.Config.ValidatorBlsMnemonic)
	s.Require().NoError(err)
	s.Relayer, err = keys.NewMnemonicKeyManager(s.Config.RelayerMnemonic)
	s.Require().NoError(err)
	s.Challenger, err = keys.NewMnemonicKeyManager(s.Config.ChallengerMnemonic)
	s.Require().NoError(err)

	var spIDs []uint32
	for i, spMnemonics := range s.Config.SPMnemonics {
		sp := StorageProvider{}
		sp.OperatorKey, err = keys.NewMnemonicKeyManager(spMnemonics.OperatorMnemonic)
		s.Require().NoError(err)
		sp.SealKey, err = keys.NewMnemonicKeyManager(spMnemonics.SealMnemonic)
		s.Require().NoError(err)
		sp.FundingKey, err = keys.NewMnemonicKeyManager(spMnemonics.FundingMnemonic)
		s.Require().NoError(err)
		sp.ApprovalKey, err = keys.NewMnemonicKeyManager(spMnemonics.ApprovalMnemonic)
		s.Require().NoError(err)
		sp.GcKey, err = keys.NewMnemonicKeyManager(spMnemonics.GcMnemonic)
		s.Require().NoError(err)
		sp.BlsKey, err = keys.NewBlsMnemonicKeyManager(s.Config.SPBLSMnemonic[i])
		s.Require().NoError(err)
		var resp *sptypes.QueryStorageProviderByOperatorAddressResponse
		resp, err = s.Client.StorageProviderByOperatorAddress(context.Background(), &sptypes.QueryStorageProviderByOperatorAddressRequest{
			OperatorAddress: sp.OperatorKey.GetAddr().String(),
		})
		s.Require().NoError(err)
		sp.Info = resp.StorageProvider
		sp.GlobalVirtualGroupFamilies = make(map[uint32][]*virtualgroupmoduletypes.GlobalVirtualGroup)
		s.StorageProviders = append(s.StorageProviders, sp)

		spIDs = append(spIDs, sp.Info.Id)
	}

	for i, sp := range s.StorageProviders {
		var gvgFamilies []*virtualgroupmoduletypes.GlobalVirtualGroupFamily
		resp1, err1 := s.Client.GlobalVirtualGroupFamilies(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamiliesRequest{StorageProviderId: sp.Info.Id})
		s.Require().NoError(err1)
		if len(resp1.GlobalVirtualGroupFamilies) == 0 {
			// Create a GVG for each sp by default
			deposit := sdk.Coin{
				Denom:  s.Config.Denom,
				Amount: types.NewIntFromInt64WithDecimal(1, types.DecimalBNB),
			}
			secondaryIds := append(spIDs[:i], spIDs[i+1:]...)
			msgCreateGVG := &virtualgroupmoduletypes.MsgCreateGlobalVirtualGroup{
				StorageProvider: sp.OperatorKey.GetAddr().String(),
				SecondarySpIds:  secondaryIds,
				Deposit:         deposit,
			}
			s.SendTxBlock(sp.OperatorKey, msgCreateGVG)
			resp2, err2 := s.Client.GlobalVirtualGroupFamilies(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupFamiliesRequest{StorageProviderId: sp.Info.Id})
			s.Require().NoError(err2)

			gvgFamilies = resp2.GlobalVirtualGroupFamilies
		} else {
			gvgFamilies = resp1.GlobalVirtualGroupFamilies

		}

		for _, family := range gvgFamilies {
			gvgsResp, err3 := s.Client.GlobalVirtualGroupByFamilyID(context.Background(), &virtualgroupmoduletypes.QueryGlobalVirtualGroupByFamilyIDRequest{
				StorageProviderId:          sp.Info.Id,
				GlobalVirtualGroupFamilyId: family.Id,
			})
			s.Require().NoError(err3)
			sp.GlobalVirtualGroupFamilies[family.Id] = gvgsResp.GlobalVirtualGroups
			s.StorageProviders[i] = sp
		}
	}
}

func (s *BaseSuite) SendTxBlock(from keys.KeyManager, msg ...sdk.Msg) *sdk.TxResponse {
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	response, err := s.Client.BroadcastTx(context.Background(), append([]sdk.Msg{}, msg...), txOpt)
	s.Require().NoError(err)

	s.Require().NoError(s.CheckTxCode(response.TxResponse.TxHash, uint32(0)), "tx failed")
	getTxRes, err := s.Client.GetTx(context.Background(), &tx.GetTxRequest{
		Hash: response.TxResponse.TxHash,
	})
	s.Require().NoError(err)

	s.T().Logf("block_height: %d, tx_hash: 0x%s", getTxRes.TxResponse.Height, response.TxResponse.TxHash)
	return getTxRes.TxResponse
}

func (s *BaseSuite) SendTxWithTxOpt(msg sdk.Msg, from keys.KeyManager, txOpt types.TxOption) {
	s.Client.SetKeyManager(from)
	_, err := s.Client.BroadcastTx(context.Background(), []sdk.Msg{msg}, &txOpt)
	s.Require().NoError(err)
}

func (s *BaseSuite) SimulateTx(msg sdk.Msg, from keys.KeyManager) (txRes *tx.SimulateResponse) {
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	response, err := s.Client.SimulateTx(context.Background(), []sdk.Msg{msg}, txOpt)
	s.Require().NoError(err)
	return response
}

func (s *BaseSuite) SendTxBlockWithoutCheck(msg sdk.Msg, from keys.KeyManager) (*tx.BroadcastTxResponse, error) {
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	return s.Client.BroadcastTx(context.Background(), []sdk.Msg{msg}, txOpt)
}

func (s *BaseSuite) SendTxBlockWithExpectErrorString(msg sdk.Msg, from keys.KeyManager, expectErrorString string) {
	mode := tx.BroadcastMode_BROADCAST_MODE_SYNC
	txOpt := &types.TxOption{
		Mode: &mode,
		Memo: "",
	}
	s.Client.SetKeyManager(from)
	_, err := s.Client.BroadcastTx(context.Background(), []sdk.Msg{msg}, txOpt)
	s.T().Logf("tx failed, err: %s, expect error string: %s", err, expectErrorString)
	s.Require().Error(err)
	s.Require().True(strings.Contains(err.Error(), expectErrorString))
}

func (s *BaseSuite) GenAndChargeAccounts(n int, balance int64) (accounts []keys.KeyManager) {
	var outputs []banktypes.Output
	denom := s.Config.Denom
	for i := 0; i < n; i++ {
		km := GenRandomKeyManager()
		accounts = append(accounts, km)
		outputs = append(outputs, banktypes.Output{
			Address: km.GetAddr().String(),
			Coins:   []sdk.Coin{{Denom: denom, Amount: types.NewIntFromInt64WithDecimal(balance, types.DecimalBNB)}},
		})
	}
	if balance == 0 {
		return
	}
	// prevent int64 multiplication overflow
	balanceInt := types.NewIntFromInt64WithDecimal(balance, types.DecimalBNB)
	nInt := sdkmath.NewInt(int64(n))
	in := banktypes.Input{
		Address: s.Validator.GetAddr().String(),
		Coins:   []sdk.Coin{{Denom: denom, Amount: balanceInt.Mul(nInt)}},
	}
	msg := banktypes.MsgMultiSend{
		Inputs:  []banktypes.Input{in},
		Outputs: outputs,
	}
	_ = s.SendTxBlock(s.Validator, &msg)
	return accounts
}

func (s *BaseSuite) GenRandomBlsKeyManager() keys.KeyManager {
	blsPrivKey, err := bls.RandKey()
	if err != nil {
		panic("failed to init bls key")
	}
	km, err := keys.NewBlsPrivateKeyManager(hex.EncodeToString(blsPrivKey.Marshal()))
	if err != nil {
		panic("failed to init bls key manager")
	}
	return km
}

func (s *BaseSuite) CheckTxCode(txHash string, expectedCode uint32) error {
	// wait for 2 blocks
	for i := 0; i < 2; i++ {
		if err := s.WaitForNextBlock(); err != nil {
			return fmt.Errorf("failed to wait for next block: %w", err)
		}
	}

	res, err := s.Client.GetTx(context.Background(), &tx.GetTxRequest{
		Hash: txHash,
	})
	if err != nil {
		return err
	}

	if res.TxResponse.Code != expectedCode {
		return fmt.Errorf("expected code %d, got %d", expectedCode, res.TxResponse.Code)
	}

	return nil
}

func (s *BaseSuite) WaitForNextBlock() error {
	lastBlock, err := s.LatestHeight()
	if err != nil {
		return err
	}

	_, err = s.WaitForHeightWithTimeout(lastBlock+1, 10*time.Second)
	if err != nil {
		return err
	}

	return err
}

func (s *BaseSuite) WaitForHeightWithTimeout(h int64, t time.Duration) (int64, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(t)
	defer timeout.Stop()

	var latestHeight int64
	queryClient := s.TmClient.TmClient

	for {
		select {
		case <-timeout.C:
			return latestHeight, errors.New("timeout exceeded waiting for block")
		case <-ticker.C:
			res, err := queryClient.Block(context.Background(), nil)
			if err == nil && res != nil {
				latestHeight = res.Block.Height
				if latestHeight >= h {
					return latestHeight, nil
				}
			}
		}
	}
}

func (s *BaseSuite) WaitForTx(hash string) (*sdk.TxResponse, error) {
	for {
		txResponse, err := s.Client.GetTx(context.Background(), &tx.GetTxRequest{Hash: hash})
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				// Tx not found, wait for next block and try again
				err := s.WaitForNextBlock()
				if err != nil {
					return nil, err
				}
				continue
			}
			return nil, err
		}
		// Tx found
		return txResponse.TxResponse, nil
	}
}

func (s *BaseSuite) LatestHeight() (int64, error) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.NewTimer(time.Second * 5)
	defer timeout.Stop()

	var latestHeight int64
	queryClient := s.TmClient.TmClient

	for {
		select {
		case <-timeout.C:
			return latestHeight, errors.New("timeout exceeded waiting for block")
		case <-ticker.C:
			res, err := queryClient.Block(context.Background(), nil)
			if err == nil && res != nil {
				return res.Block.Height, nil
			}
		}
	}
}

func (sp *StorageProvider) GetFirstGlobalVirtualGroup() (*virtualgroupmoduletypes.GlobalVirtualGroup, bool) {
	for _, family := range sp.GlobalVirtualGroupFamilies {
		if len(family) != 0 {
			return family[0], true
		}
	}
	return nil, false
}

func (s *BaseSuite) NewSpAcc() *StorageProvider {
	userAccs := s.GenAndChargeAccounts(5, 1000000)
	operatorAcc := userAccs[0]
	fundingAcc := userAccs[1]
	approvalAcc := userAccs[2]
	sealAcc := userAccs[3]
	gcAcc := userAccs[4]

	blsKm := s.GenRandomBlsKeyManager()
	return &StorageProvider{OperatorKey: operatorAcc, SealKey: fundingAcc,
		FundingKey: approvalAcc, ApprovalKey: sealAcc, GcKey: gcAcc, BlsKey: blsKm}
}

func (s *BaseSuite) CreateNewStorageProvider() *StorageProvider {
	validator := s.Validator.GetAddr()

	// 1. create new newStorageProvider
	newSP := s.NewSpAcc()

	// 2. grant deposit authorization of sp to gov module account
	coins := sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB))
	authorization := sptypes.NewDepositAuthorization(newSP.OperatorKey.GetAddr(), &coins)

	govAddr := authtypes.NewModuleAddress(gov.ModuleName)
	now := time.Now().Add(24 * time.Hour)
	grantMsg, err := authz.NewMsgGrant(
		newSP.FundingKey.GetAddr(), govAddr, authorization, &now)
	s.Require().NoError(err)
	s.SendTxBlock(newSP.FundingKey, grantMsg)

	// 2. submit CreateStorageProvider proposal
	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(10000, types.DecimalBNB),
	}
	description := sptypes.Description{
		Moniker:  "sp_test",
		Identity: "",
	}

	endpoint := "http://127.0.0.1:9034"
	newReadPrice := sdk.NewDec(RandInt64(100, 200))
	newStorePrice := sdk.NewDec(RandInt64(10000, 20000))

	// bls pub key
	newSpBlsKm := newSP.BlsKey
	blsProofBz, err := newSpBlsKm.Sign(tmhash.Sum(newSpBlsKm.PubKey().Bytes()))
	s.Require().NoError(err)

	msgCreateSP, _ := sptypes.NewMsgCreateStorageProvider(govAddr,
		newSP.OperatorKey.GetAddr(), newSP.FundingKey.GetAddr(),
		newSP.SealKey.GetAddr(),
		newSP.ApprovalKey.GetAddr(),
		newSP.GcKey.GetAddr(), description,
		endpoint, deposit, newReadPrice, 10000, newStorePrice,
		hex.EncodeToString(newSP.BlsKey.PubKey().Bytes()),
		hex.EncodeToString(blsProofBz),
	)

	msgProposal, err := govtypesv1.NewMsgSubmitProposal(
		[]sdk.Msg{msgCreateSP},
		sdk.Coins{sdk.NewCoin(s.Config.Denom, types.NewIntFromInt64WithDecimal(100, types.DecimalBNB))},
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

	queryProposal := &govtypesv1.QueryProposalRequest{ProposalId: proposalId}
	_, err = s.Client.GovQueryClientV1.Proposal(context.Background(), queryProposal)
	s.Require().NoError(err)

	// 4. submit MsgVote and wait the proposal exec
	msgVote := govtypesv1.NewMsgVote(validator, proposalId, govtypesv1.OptionYes, "test")
	txRes = s.SendTxBlock(s.Validator, msgVote)
	s.Require().Equal(txRes.Code, uint32(0))

	queryVoteParamsReq := govtypesv1.QueryParamsRequest{ParamsType: "voting"}
	queryVoteParamsResp, err := s.Client.GovQueryClientV1.Params(context.Background(), &queryVoteParamsReq)
	s.Require().NoError(err)

	// 5. wait a voting period and confirm that the proposal success.
	s.T().Logf("voting period %s", *queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(*queryVoteParamsResp.Params.VotingPeriod)
	time.Sleep(1 * time.Second)
	proposalRes, err := s.Client.GovQueryClientV1.Proposal(context.Background(), queryProposal)
	s.Require().NoError(err)
	s.Require().Equal(proposalRes.Proposal.Status, govtypesv1.ProposalStatus_PROPOSAL_STATUS_PASSED)

	// 6. query storage provider
	querySPByOperatorAddrReq := sptypes.QueryStorageProviderByOperatorAddressRequest{
		OperatorAddress: newSP.OperatorKey.GetAddr().String(),
	}
	querySPByOperatorAddrResp, err := s.Client.StorageProviderByOperatorAddress(context.Background(), &querySPByOperatorAddrReq)
	s.Require().NoError(err)
	s.Require().Equal(querySPByOperatorAddrResp.StorageProvider.OperatorAddress, newSP.OperatorKey.GetAddr().String())
	s.Require().Equal(querySPByOperatorAddrResp.StorageProvider.FundingAddress, newSP.FundingKey.GetAddr().String())
	s.Require().Equal(querySPByOperatorAddrResp.StorageProvider.SealAddress, newSP.SealKey.GetAddr().String())
	s.Require().Equal(querySPByOperatorAddrResp.StorageProvider.ApprovalAddress, newSP.ApprovalKey.GetAddr().String())
	s.Require().Equal(querySPByOperatorAddrResp.StorageProvider.Endpoint, endpoint)
	newSP.Info = querySPByOperatorAddrResp.StorageProvider
	return newSP
}

func (s *BaseSuite) CreateObject(user keys.KeyManager, primarySP *StorageProvider, gvgID uint32, bucketName, objectName string) (secondarySps []*StorageProvider, familyID, resGVGID uint32, bucketInfo storagetypes.BucketInfo) {
	// GetGVG
	resp, err := s.Client.GlobalVirtualGroup(
		context.Background(),
		&virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{GlobalVirtualGroupId: gvgID})
	s.Require().NoError(err)
	gvg := resp.GlobalVirtualGroup

	// CreateBucket
	msgCreateBucket := storagetypes.NewMsgCreateBucket(
		user.GetAddr(), bucketName, storagetypes.VISIBILITY_TYPE_PRIVATE, primarySP.OperatorKey.GetAddr(),
		nil, math.MaxUint, nil, 0)
	msgCreateBucket.PrimarySpApproval.GlobalVirtualGroupFamilyId = gvg.FamilyId
	msgCreateBucket.PrimarySpApproval.Sig, err = primarySP.ApprovalKey.Sign(msgCreateBucket.GetApprovalBytes())
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
	bucketInfo = *queryHeadBucketResponse.BucketInfo

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
	msgCreateObject.PrimarySpApproval.Sig, err = primarySP.ApprovalKey.Sign(msgCreateObject.GetApprovalBytes())
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

	// query gvg
	queryGlobalvirtualGroupResp, err := s.Client.GlobalVirtualGroup(ctx, &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	originGVG := queryGlobalvirtualGroupResp.GlobalVirtualGroup
	// SealObject
	gvgId := gvg.Id
	msgSealObject := storagetypes.NewMsgSealObject(primarySP.SealKey.GetAddr(), bucketName, objectName, gvg.Id, nil)

	secondarySigs := make([][]byte, 0)
	secondarySPBlsPubKeys := make([]bls.PublicKey, 0)
	blsSignHash := storagetypes.NewSecondarySpSealObjectSignDoc(s.GetChainID(), gvgId, queryHeadObjectResponse.ObjectInfo.Id, storagetypes.GenerateHash(queryHeadObjectResponse.ObjectInfo.Checksums[:])).GetBlsSignHash()
	// every secondary sp signs the checksums
	for i := 1; i < len(s.StorageProviders); i++ {
		sig, err := BlsSignAndVerify(s.StorageProviders[i], blsSignHash)
		s.Require().NoError(err)
		secondarySigs = append(secondarySigs, sig)
		pk, err := bls.PublicKeyFromBytes(s.StorageProviders[i].BlsKey.PubKey().Bytes())
		s.Require().NoError(err)
		secondarySPBlsPubKeys = append(secondarySPBlsPubKeys, pk)
		if s.StorageProviders[i].Info.Id != primarySP.Info.Id {
			ssp := s.StorageProviders[i]
			secondarySps = append(secondarySps, &ssp)
		}
	}
	aggBlsSig, err := BlsAggregateAndVerify(secondarySPBlsPubKeys, blsSignHash, secondarySigs)
	s.Require().NoError(err)
	msgSealObject.SecondarySpBlsAggSignatures = aggBlsSig
	s.SendTxBlock(primarySP.SealKey, msgSealObject)

	queryHeadObjectResponse, err = s.Client.HeadObject(ctx, &queryHeadObjectRequest)
	s.Require().NoError(err)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectName, objectName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.BucketName, bucketName)
	s.Require().Equal(queryHeadObjectResponse.ObjectInfo.ObjectStatus, storagetypes.OBJECT_STATUS_SEALED)

	// verify gvg store size
	queryGlobalvirtualGroupResp, err = s.Client.GlobalVirtualGroup(ctx, &virtualgroupmoduletypes.QueryGlobalVirtualGroupRequest{
		GlobalVirtualGroupId: gvgID,
	})
	s.Require().NoError(err)
	s.Require().Equal(originGVG.StoredSize+uint64(payloadSize), queryGlobalvirtualGroupResp.GlobalVirtualGroup.StoredSize)

	return secondarySps, gvg.FamilyId, gvg.Id, bucketInfo
}

func (s *BaseSuite) CreateGlobalVirtualGroup(sp *StorageProvider, familyID uint32, secondarySPIDs []uint32, depositAmount int64) (uint32, uint32) {
	// Create a GVG for each sp by default
	deposit := sdk.Coin{
		Denom:  s.Config.Denom,
		Amount: types.NewIntFromInt64WithDecimal(depositAmount, types.DecimalBNB),
	}
	msgCreateGVG := &virtualgroupmoduletypes.MsgCreateGlobalVirtualGroup{
		StorageProvider: sp.OperatorKey.GetAddr().String(),
		SecondarySpIds:  secondarySPIDs,
		Deposit:         deposit,
		FamilyId:        familyID,
	}
	resp := s.SendTxBlock(sp.OperatorKey, msgCreateGVG)

	// wait for the tx execute
	resp2, err := s.WaitForTx(resp.TxHash)
	s.Require().NoError(err)

	var gvgID uint32
	var newFamilyID uint32
	for _, e := range resp2.Events {
		s.T().Logf("Event: %s", e.String())
		if e.Type == "greenfield.virtualgroup.EventCreateGlobalVirtualGroup" {
			for _, a := range e.Attributes {
				if a.Key == "id" {
					num, err := strconv.ParseUint(a.Value, 10, 32)
					s.Require().NoError(err)
					gvgID = uint32(num)
				}
				if a.Key == "family_id" {
					num, err := strconv.ParseUint(a.Value, 10, 32)
					s.Require().NoError(err)
					newFamilyID = uint32(num)
				}
			}
		}
	}
	s.T().Logf("gvgID: %d, familyID: %d", gvgID, newFamilyID)
	return gvgID, newFamilyID
}

func (s *BaseSuite) GetChainID() string {
	return s.Config.ChainId
}
