package types

import (
	authztypes "github.com/cosmos/cosmos-sdk/x/authz"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	feegranttypes "github.com/cosmos/cosmos-sdk/x/feegrant"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	oracletypes "github.com/cosmos/cosmos-sdk/x/oracle/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	bridgetypes "github.com/bnb-chain/greenfield/x/bridge/types"
	challengetypes "github.com/bnb-chain/greenfield/x/challenge/types"
	paymenttypes "github.com/bnb-chain/greenfield/x/payment/types"
	sptypes "github.com/bnb-chain/greenfield/x/sp/types"
	storagetypes "github.com/bnb-chain/greenfield/x/storage/types"
)

type (
	MsgGrant  = authztypes.MsgGrant
	MsgRevoke = authztypes.MsgRevoke

	MsgSend = banktypes.MsgSend

	MsgCreateValidator           = stakingtypes.MsgCreateValidator
	MsgEditValidator             = stakingtypes.MsgEditValidator
	MsgDelegate                  = stakingtypes.MsgDelegate
	MsgBeginRedelegate           = stakingtypes.MsgBeginRedelegate
	MsgUndelegate                = stakingtypes.MsgUndelegate
	MsgCancelUnbondingDelegation = stakingtypes.MsgCancelUnbondingDelegation

	MsgSetWithdrawAddress          = distrtypes.MsgSetWithdrawAddress
	MsgWithdrawDelegatorReward     = distrtypes.MsgWithdrawDelegatorReward
	MsgWithdrawValidatorCommission = distrtypes.MsgWithdrawValidatorCommission
	MsgFundCommunityPool           = distrtypes.MsgFundCommunityPool

	MsgSubmitProposal    = govv1.MsgSubmitProposal
	MsgExecLegacyContent = govv1.MsgExecLegacyContent
	MsgVote              = govv1.MsgVote
	MsgGovDeposit        = govv1.MsgDeposit
	MsgVoteWeighted      = govv1.MsgVoteWeighted

	MsgUnjail  = slashingtypes.MsgUnjail
	MsgImpeach = slashingtypes.MsgImpeach

	MsgGrantAllowance  = feegranttypes.MsgGrantAllowance
	MsgRevokeAllowance = feegranttypes.MsgRevokeAllowance

	MsgClaim = oracletypes.MsgClaim

	MsgTransferOut = bridgetypes.MsgTransferOut

	MsgCreatePaymentAccount = paymenttypes.MsgCreatePaymentAccount
	MsgPaymentDeposit       = paymenttypes.MsgDeposit
	MsgWithdraw             = paymenttypes.MsgWithdraw
	MsgDisableRefund        = paymenttypes.MsgDisableRefund

	MsgCreateStorageProvider = sptypes.MsgCreateStorageProvider
	MsgSpDeposit             = sptypes.MsgDeposit
	MsgEditStorageProvider   = sptypes.MsgEditStorageProvider

	MsgCreateBucket      = storagetypes.MsgCreateBucket
	MsgDeleteBucket      = storagetypes.MsgDeleteBucket
	MsgCreateObject      = storagetypes.MsgCreateObject
	MsgSealObject        = storagetypes.MsgSealObject
	MsgRejectSealObject  = storagetypes.MsgRejectSealObject
	MsgCopyObject        = storagetypes.MsgCopyObject
	MsgDeleteObject      = storagetypes.MsgDeleteObject
	MsgCreateGroup       = storagetypes.MsgCreateGroup
	MsgDeleteGroup       = storagetypes.MsgDeleteGroup
	MsgUpdateGroupMember = storagetypes.MsgUpdateGroupMember
	MsgLeaveGroup        = storagetypes.MsgLeaveGroup

	MsgSubmit = challengetypes.MsgSubmit
	MsgAttest = challengetypes.MsgAttest
)
