package types

import (
	"encoding/binary"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

const TypeMsgAttest = "attest"

var _ sdk.Msg = &MsgAttest{}

func NewMsgAttest(creator sdk.AccAddress, challengeId uint64, voteResult uint32, voteValidatorSet []uint64, voteAggSignature []byte) *MsgAttest {
	return &MsgAttest{
		Creator:          creator.String(),
		ChallengeId:      challengeId,
		VoteResult:       voteResult,
		VoteValidatorSet: voteValidatorSet,
		VoteAggSignature: voteAggSignature,
	}
}

func (msg *MsgAttest) Route() string {
	return RouterKey
}

func (msg *MsgAttest) Type() string {
	return TypeMsgAttest
}

func (msg *MsgAttest) GetSigners() []sdk.AccAddress {
	creator, err := sdk.AccAddressFromBech32(msg.Creator)
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

	if msg.VoteResult != ChallengeResultSucceed {
		return sdkerrors.Wrap(ErrInvalidVoteResult, "Only succeed challenge can submit attest")
	}

	if len(msg.VoteValidatorSet) == 0 {
		return sdkerrors.Wrap(ErrInvalidVoteValidatorSet, "Vote validator set cannot be empty")
	}

	if len(msg.VoteAggSignature) != 96 {
		return sdkerrors.Wrap(ErrInvalidVoteAggSignature, "Length of aggregated signature is invalid")
	}

	return nil
}

func (msg *MsgAttest) GetBlsSignBytes() [32]byte {
	idBz := make([]byte, 8)
	binary.BigEndian.PutUint64(idBz, msg.ChallengeId)
	resultBz := make([]byte, 8)
	binary.BigEndian.PutUint64(resultBz, uint64(msg.VoteResult))

	bs := make([]byte, 0)
	bs = append(bs, idBz...)
	bs = append(bs, resultBz...)
	hash := sdk.Keccak256Hash(bs)
	return hash
}
