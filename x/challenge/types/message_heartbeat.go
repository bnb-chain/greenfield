package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgHeartbeat = "heartbeat"

var _ sdk.Msg = &MsgHeartbeat{}

func NewMsgHeartbeat(creator sdk.AccAddress, challengeId uint64, voteValidatorSet []uint64, voteAggSignature []byte) *MsgHeartbeat {
	return &MsgHeartbeat{
		Creator:          creator.String(),
		ChallengeId:      challengeId,
		VoteValidatorSet: voteValidatorSet,
		VoteAggSignature: voteAggSignature,
	}
}

func (msg *MsgHeartbeat) Route() string {
	return RouterKey
}

func (msg *MsgHeartbeat) Type() string {
	return TypeMsgHeartbeat
}

func (msg *MsgHeartbeat) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		panic(err)
	}
	return []sdk.AccAddress{creator}
}

func (msg *MsgHeartbeat) GetSignBytes() []byte {
	bz := ModuleCdc.MustMarshalJSON(msg)
	return sdk.MustSortJSON(bz)
}

func (msg *MsgHeartbeat) ValidateBasic() error {
	_, err := sdk.AccAddressFromHexUnsafe(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}

	if len(msg.VoteValidatorSet) == 0 {
		return sdkerrors.Wrap(ErrInvalidVoteValidatorSet, "vote validator set cannot be empty")
	}

	if len(msg.VoteAggSignature) != BlsSignatureLength {
		return sdkerrors.Wrap(ErrInvalidVoteAggSignature, "length of aggregated signature is invalid")
	}

	return nil
}

func (msg *MsgHeartbeat) GetBlsSignBytes() [32]byte {
	idBz := make([]byte, 8)
	binary.BigEndian.PutUint64(idBz, msg.ChallengeId)

	hash := sdk.Keccak256Hash(idBz)
	return hash
}
