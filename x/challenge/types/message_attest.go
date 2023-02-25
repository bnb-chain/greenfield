package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAttest = "attest"

var _ sdk.Msg = &MsgAttest{}

func NewMsgAttest(creator sdk.AccAddress, challengeId, objectId uint64, spOperatorAddress sdk.AccAddress, voteResult uint32,
	voteValidatorSet []uint64, voteAggSignature []byte) *MsgAttest {
	return &MsgAttest{
		Creator:           creator.String(),
		ChallengeId:       challengeId,
		ObjectId:          objectId,
		SpOperatorAddress: spOperatorAddress.String(),
		VoteResult:        voteResult,
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
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgAttest) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgAttest) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	_, err = sdk.AccAddressFromHexUnsafe(msg.SpOperatorAddress)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid sp operator address (%s)", err)
	}

	if msg.VoteResult != ChallengeResultSucceed {
		return sdkerrors.Wrap(ErrInvalidVoteResult, "only succeed challenge can submit attest")
	}

	if len(msg.VoteValidatorSet) == 0 {
		return sdkerrors.Wrap(ErrInvalidVoteValidatorSet, "vote validator set cannot be empty")
	}

	if len(msg.VoteAggSignature) != BlsSignatureLength {
		return sdkerrors.Wrap(ErrInvalidVoteAggSignature, "length of aggregated signature is invalid")
	}

	return nil
}

func (msg *MsgAttest) GetBlsSignBytes() [32]byte {
	challengeIdBz := make([]byte, 8)
	binary.BigEndian.PutUint64(challengeIdBz, msg.ChallengeId)
	objectIdBz := make([]byte, 8)
	binary.BigEndian.PutUint64(objectIdBz, uint64(msg.ObjectId))
	resultBz := make([]byte, 8)
	binary.BigEndian.PutUint64(resultBz, uint64(msg.VoteResult))

	bs := make([]byte, 0)
	bs = append(bs, challengeIdBz...)
	bs = append(bs, objectIdBz...)
	bs = append(bs, resultBz...)
	bs = append(bs, []byte(msg.SpOperatorAddress)...)
	hash := sdk.Keccak256Hash(bs)
	return hash
}
