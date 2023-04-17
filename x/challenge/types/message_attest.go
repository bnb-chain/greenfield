package types

import (
	"encoding/binary"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAttest = "attest"

var _ sdk.Msg = &MsgAttest{}

func NewMsgAttest(submitter sdk.AccAddress, challengeId uint64, objectId Uint, spOperatorAddress string,
	voteResult VoteResult, challenger string, voteValidatorSet []uint64, voteAggSignature []byte) *MsgAttest {
	return &MsgAttest{
		Submitter:         submitter.String(),
		ChallengeId:       challengeId,
		ObjectId:          objectId,
		SpOperatorAddress: spOperatorAddress,
		VoteResult:        VoteResult(voteResult),
		ChallengerAddress: challenger,
		VoteValidatorSet:  voteValidatorSet,
		VoteAggSignature:  voteAggSignature,
	}
}

func (msg *MsgAttest) Route() string {
	return RouterKey
}

func (msg *MsgAttest) Type() string {
	return TypeMsgAttest
}

func (msg *MsgAttest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Submitter)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAttest) GetSignBytes() []byte {
	panic("GetSignBytes")
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAttest) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Submitter)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid submitter address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp operator address (%s)", err)
	}

	if msg.VoteResult != CHALLENGE_SUCCEED && msg.VoteResult != CHALLENGE_FAILED {
		return errors.Wrap(ErrInvalidVoteResult, "vote result should be 0 or 1")
	}

	if msg.ChallengerAddress != "" {
		_, err = sdk.AccAddressFromHexUnsafe(msg.ChallengerAddress)
		if err != nil {
			return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid challenger address (%s)", err)
		}
	}

	if len(msg.VoteValidatorSet) == 0 {
		return errors.Wrap(ErrInvalidVoteValidatorSet, "vote validator set cannot be empty")
	}

	if len(msg.VoteAggSignature) != BlsSignatureLength {
		return errors.Wrap(ErrInvalidVoteAggSignature, "length of aggregated signature is invalid")
	}

	return nil
}

func (msg *MsgAttest) GetBlsSignBytes() [32]byte {
	challengeIdBz := make([]byte, 8)
	binary.BigEndian.PutUint64(challengeIdBz, msg.ChallengeId)
	objectIdBz := msg.ObjectId.Bytes()
	resultBz := make([]byte, 8)
	binary.BigEndian.PutUint64(resultBz, uint64(msg.VoteResult))

	spOperatorBz := sdk.MustAccAddressFromHex(msg.SpOperatorAddress).Bytes()
	challengerBz := make([]byte, 0)
	if msg.ChallengerAddress != "" {
		challengerBz = sdk.MustAccAddressFromHex(msg.ChallengerAddress).Bytes()
	}

	bs := make([]byte, 0)
	bs = append(bs, challengeIdBz...)
	bs = append(bs, objectIdBz...)
	bs = append(bs, resultBz...)
	bs = append(bs, spOperatorBz...)
	bs = append(bs, challengerBz...)
	hash := sdk.Keccak256Hash(bs)
	return hash
}
