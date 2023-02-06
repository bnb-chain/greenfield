package types

import (
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
	_, err := sdk.AccAddressFromBech32(msg.Creator)
	if err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid creator address (%s)", err)
	}
	return nil
}
